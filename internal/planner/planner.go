package planner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cyperx84/multiplan/internal/config"
	"github.com/cyperx84/multiplan/internal/eval"
	"github.com/cyperx84/multiplan/internal/models"
)

// RunResult is returned by Run.
type RunResult struct {
	RunID      string
	OutputDir  string
	Plans      []models.ModelResult
	Debate     string
	FinalPlan  string
	EvalScores map[string]float64
}

// configProviderAdapter wraps a config.Provider so it satisfies models.Provider.
type configProviderAdapter struct {
	p config.Provider
}

func (a *configProviderAdapter) ID() string   { return a.p.ID() }
func (a *configProviderAdapter) Name() string { return a.p.Name() }
func (a *configProviderAdapter) Available(ctx context.Context) bool {
	return a.p.Available(ctx)
}
func (a *configProviderAdapter) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	return a.p.Plan(ctx, prompt, timeout)
}

// getProvider returns a models.Provider using the factory (if set) or global registry.
func getProvider(cfg *config.Config, id string) (models.Provider, bool) {
	if cfg.ProviderFactory != nil {
		p, ok := cfg.ProviderFactory(id)
		if !ok {
			return nil, false
		}
		return &configProviderAdapter{p}, true
	}
	return models.GetProvider(id)
}

// Run executes the full 3-phase multiplan pipeline.
func Run(cfg *config.Config) (*RunResult, error) {
	start := time.Now()
	ctx := context.Background()

	runID := generateRunID()
	requirements := cfg.GetRequirements()
	constraints := cfg.GetConstraints()
	timeoutMs := cfg.GetTimeoutMs()

	// Determine output directory
	outputDir := cfg.OutputDir
	if outputDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		outputDir = filepath.Join(home, ".multiplan", "runs", runID)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	log := func(format string, args ...interface{}) {
		if !cfg.Quiet && cfg.Verbose {
			fmt.Printf(format+"\n", args...)
		}
	}

	logf := func(format string, args ...interface{}) {
		if !cfg.Quiet {
			fmt.Printf(format, args...)
		}
	}

	// Phase 0: Lattice mental model framing
	var latticeModels []string
	if !cfg.SkipLattice {
		latticeCmd := cfg.LatticeCmd
		if latticeCmd == "" {
			latticeCmd = "lattice"
		}
		if _, err := exec.LookPath(latticeCmd); err == nil {
			logf("[multiplan] Phase 0 — Lattice mental model framing...\n")
			latticeModels = runLatticeFraming(latticeCmd, cfg.Task, outputDir, cfg.Verbose)
			if len(latticeModels) > 0 {
				logf("[multiplan]   ✓ Models: %s\n", strings.Join(latticeModels, ", "))
			}
		} else if cfg.Verbose {
			log("[multiplan] Phase 0 — Skipped (lattice not on PATH)")
		}
	}

	log("[multiplan] Run ID: %s", runID)
	log("[multiplan] Output: %s", outputDir)

	// ── Phase 1: Parallel planning ───────────────────────────────────────────
	log("[multiplan] Phase 1 — Running models in parallel with lens-based prompts...")

	// Wire claude_cmd / claude_model config into the provider
	if cfg.ProviderFactory == nil {
		if p, ok := models.GetProvider("claude"); ok {
			if cp, ok := p.(*models.ClaudeProvider); ok {
				if cfg.ClaudeCmd != "" {
					cp.ClaudeCmd = cfg.ClaudeCmd
				}
				if cfg.ClaudeModel != "" {
					cp.ClaudeModel = cfg.ClaudeModel
				}
			}
		}
	}

	modelIDs := cfg.Models
	if len(modelIDs) == 0 {
		modelIDs = models.GetAvailableModels(ctx)
	}

	results := make([]models.ModelResult, 0, len(modelIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup

	resultChan := make(chan models.ModelResult, len(modelIDs))

	for _, modelID := range modelIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			provider, ok := getProvider(cfg, id)
			if !ok {
				return
			}

			// Check availability — mark as skipped if unavailable
			if !provider.Available(ctx) {
				result := models.ModelResult{
					ModelID:   id,
					ModelName: provider.Name(),
					Skipped:   true,
					Error:     "unavailable",
				}
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				resultChan <- result
				return
			}

			startTime := time.Now()
			prompt := GetLensPrompt(id, cfg.Task, requirements, constraints)

			// Inject lattice mental model framing if available
			if len(latticeModels) > 0 {
				prompt += "\n\nRelevant mental models: [" + strings.Join(latticeModels, ", ") + "]\nConsider these frameworks in your plan."
			}

			timeout := time.Duration(timeoutMs) * time.Millisecond

			var plan string
			var inputTok, outputTok int
			var runErr error

			if pt, ok := provider.(models.ProviderWithTokens); ok {
				plan, inputTok, outputTok, runErr = pt.PlanWithTokens(ctx, prompt, timeout)
			} else {
				plan, runErr = provider.Plan(ctx, prompt, timeout)
			}

			durationMs := time.Since(startTime).Milliseconds()

			result := models.ModelResult{
				ModelID:      id,
				ModelName:    provider.Name(),
				DurationMs:   durationMs,
				InputTokens:  inputTok,
				OutputTokens: outputTok,
			}

			if runErr != nil {
				result.Error = runErr.Error()
				result.Plan = fmt.Sprintf("[Error: %s]", runErr.Error())
			} else {
				result.Plan = plan
				filename := filepath.Join(outputDir, fmt.Sprintf("plan-%s.md", id))
				_ = os.WriteFile(filename, []byte(plan), 0644)
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			resultChan <- result
		}(modelID)
	}

	// Close channel when all goroutines finish.
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Stream progress as each model finishes.
	for r := range resultChan {
		if r.Skipped {
			if cfg.Verbose {
				log("[multiplan]   ⊘ %s skipped (unavailable)", r.ModelName)
			} else {
				logf("⏳ %s... skipped\n", r.ModelName)
			}
		} else if cfg.Verbose {
			if r.Error == "" {
				log("[multiplan]   ✓ %s done (%dms)", r.ModelName, r.DurationMs)
			} else {
				log("[multiplan]   ✗ %s failed: %s", r.ModelName, r.Error)
			}
		} else {
			durationSec := float64(r.DurationMs) / 1000.0
			if r.Error == "" {
				logf("⏳ %s... done (%.1fs)\n", r.ModelName, durationSec)
			} else {
				logf("⏳ %s... failed\n", r.ModelName)
			}
		}
	}

	log("[multiplan] Phase 1 complete")

	// ── Phase 1.5: Structural eval ───────────────────────────────────────────
	log("[multiplan] Phase 1.5 — Evaluating plans...")

	evalCase := &eval.EvalCase{Task: cfg.Task}
	planScores := make(map[string]float64)
	planTexts := make(map[string]string)

	for _, result := range results {
		if result.Error == "" {
			report, err := eval.EvalPlan(result.Plan, evalCase, &eval.EvalOptions{Judge: ""})
			if err == nil {
				planScores[result.ModelID] = report.OverallScore
				planTexts[result.ModelID] = result.Plan
				log("[multiplan]   %s: %.1f/10", result.ModelName, report.OverallScore*10)
			}
		}
	}

	// ── Phase 2: Debate ──────────────────────────────────────────────────────
	log("[multiplan] Phase 2 — Cross-examination...")

	debateProvider, ok := getProvider(cfg, cfg.DebateModel)
	if !ok {
		debateProvider, _ = getProvider(cfg, "claude")
	}

	debatePrompt := GetDebatePrompt(cfg.Task, planTexts)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	var debate string
	if debateProvider != nil {
		var err error
		debate, err = debateProvider.Plan(ctx, debatePrompt, timeout)
		if err != nil {
			debate = fmt.Sprintf("[Debate failed: %s]", err.Error())
			log("[multiplan]   ✗ Debate failed: %s", err.Error())
		} else {
			log("[multiplan]   ✓ Debate complete (via %s)", debateProvider.Name())
		}
	} else {
		debate = "[Debate skipped: no provider available]"
	}

	if err := os.WriteFile(filepath.Join(outputDir, "debate.md"), []byte(debate), 0644); err != nil {
		return nil, err
	}

	// ── Phase 3: Convergence ─────────────────────────────────────────────────
	log("[multiplan] Phase 3 — Convergence with eval scores...")

	convergeProvider, ok := getProvider(cfg, cfg.ConvergeModel)
	if !ok {
		convergeProvider, _ = getProvider(cfg, "claude")
	}

	convergePrompt := GetConvergePrompt(cfg.Task, planTexts, debate, planScores)
	var finalPlan string
	if convergeProvider != nil {
		var err error
		finalPlan, err = convergeProvider.Plan(ctx, convergePrompt, timeout)
		if err != nil {
			finalPlan = fmt.Sprintf("[Convergence failed: %s]", err.Error())
			log("[multiplan]   ✗ Convergence failed: %s", err.Error())
		} else {
			log("[multiplan]   ✓ Convergence complete (via %s)", convergeProvider.Name())
		}
	} else {
		finalPlan = "[Convergence skipped: no provider available]"
	}

	// Build header and full plan.
	modelNames := make([]string, 0, len(results))
	for _, r := range results {
		modelNames = append(modelNames, r.ModelName)
	}
	header := fmt.Sprintf("# Multimodel Plan: %s\n\n> Generated: %s\n> Models: %v\n\n---\n\n",
		cfg.Task, time.Now().Format(time.RFC3339), modelNames)
	fullPlan := header + finalPlan

	if err := os.WriteFile(filepath.Join(outputDir, "final-plan.md"), []byte(fullPlan), 0644); err != nil {
		return nil, err
	}

	log("[multiplan] Phase 3 complete")

	// ── Token cost summary ───────────────────────────────────────────────────
	var totalIn, totalOut int
	var totalCost float64
	for _, r := range results {
		totalIn += r.InputTokens
		totalOut += r.OutputTokens
		totalCost += models.EstimateCost(r.ModelID, r.InputTokens, r.OutputTokens)
	}
	if totalIn+totalOut > 0 && !cfg.Quiet {
		logf("📊 Token usage: %s input / %s output (~$%.2f estimated)\n",
			formatInt(totalIn), formatInt(totalOut), totalCost)
	}

	runResult := &RunResult{
		RunID:      runID,
		OutputDir:  outputDir,
		Plans:      results,
		Debate:     debate,
		FinalPlan:  fullPlan,
		EvalScores: planScores,
	}

	printRunSummary(runResult, time.Since(start), cfg.Quiet, cfg.JSON)

	return runResult, nil
}

func generateRunID() string {
	now := time.Now()
	timestamp := now.Format("20060102T150405")
	return fmt.Sprintf("%s-%06d", timestamp, now.Nanosecond()/1000)
}

// formatInt formats an integer with comma separators.
func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	result := []byte{}
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}
	return string(result)
}

// printRunSummary prints a human-readable summary after all phases complete.
func printRunSummary(result *RunResult, elapsed time.Duration, quiet, jsonMode bool) {
	if quiet || jsonMode {
		return
	}

	succeeded := 0
	for _, p := range result.Plans {
		if p.Error == "" {
			succeeded++
		}
	}
	total := len(result.Plans)

	fmt.Printf("\n✅ multiplan complete — %d/%d models succeeded (%.0fs)\n\n", succeeded, total, elapsed.Seconds())

	hasScores := len(result.EvalScores) > 0

	for _, p := range result.Plans {
		if p.Skipped {
			fmt.Printf("  %-10s SKIPPED (unavailable)\n", p.ModelID)
		} else if p.Error != "" {
			// Determine failure reason
			reason := p.Error
			if strings.Contains(strings.ToLower(reason), "timeout") || strings.Contains(strings.ToLower(reason), "deadline") {
				reason = "timeout"
			} else if strings.Contains(strings.ToLower(reason), "api") || strings.Contains(strings.ToLower(reason), "status") {
				reason = "API error"
			}
			fmt.Printf("  %-10s FAILED (%s)\n", p.ModelID, reason)
		} else if hasScores {
			if score, ok := result.EvalScores[p.ModelID]; ok {
				fmt.Printf("  %-10s %-6.1f plan-%s.md\n", p.ModelID, score*10, p.ModelID)
			} else {
				fmt.Printf("  %-10s %-6s plan-%s.md\n", p.ModelID, "—", p.ModelID)
			}
		} else {
			fmt.Printf("  %-10s plan-%s.md\n", p.ModelID, p.ModelID)
		}
	}

	finalPath := filepath.Join(result.OutputDir, "final-plan.md")
	fmt.Printf("\n📄 Final plan: %s\n", finalPath)
}

// latticeThinkResult is the JSON structure returned by `lattice think --json`.
type latticeThinkResult struct {
	Problem string `json:"problem"`
	Models  []struct {
		ModelName string `json:"model_name"`
		ModelSlug string `json:"model_slug"`
		Category  string `json:"category"`
	} `json:"models"`
	Summary string `json:"summary"`
}

// latticeSearchResult is a single model object returned by `lattice search --json`.
type latticeSearchResult struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

func runLatticeFraming(latticeCmd, task, outputDir string, verbose bool) []string {
	// Step 1: Search for relevant models
	slugs := latticeSearch(latticeCmd, task, verbose)
	if len(slugs) == 0 {
		// Fallback: search for generic planning models
		if verbose {
			fmt.Printf("[multiplan]   ↻ No models for task, trying fallback search...\n")
		}
		slugs = latticeSearch(latticeCmd, "planning", verbose)
	}
	if len(slugs) == 0 {
		if verbose {
			fmt.Printf("[multiplan]   ✗ Lattice search returned no models\n")
		}
		return nil
	}

	// Cap at 5 slugs
	if len(slugs) > 5 {
		slugs = slugs[:5]
	}

	// Step 2: Think with the discovered models (15s timeout — it calls an LLM)
	modelsArg := strings.Join(slugs, ",")
	thinkCtx, thinkCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer thinkCancel()
	cmd := exec.CommandContext(thinkCtx, latticeCmd, "think", task, "--models", modelsArg, "--json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if verbose {
			fmt.Printf("[multiplan]   ✗ Lattice think failed: %s\n", err)
		}
		return nil
	}

	var result latticeThinkResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		if verbose {
			fmt.Printf("[multiplan]   ✗ Lattice JSON parse failed: %s\n", err)
		}
		return nil
	}

	var modelNames []string
	for _, m := range result.Models {
		modelNames = append(modelNames, m.ModelName)
	}

	// Write lattice framing to output dir
	if len(modelNames) > 0 {
		var framingBuf strings.Builder
		framingBuf.WriteString("# Lattice Mental Model Framing\n\n")
		framingBuf.WriteString(fmt.Sprintf("Problem: %s\n\n", task))
		framingBuf.WriteString("## Models Applied\n\n")
		for _, m := range result.Models {
			framingBuf.WriteString(fmt.Sprintf("- **%s** (%s)\n", m.ModelName, m.Category))
		}
		if result.Summary != "" {
			framingBuf.WriteString(fmt.Sprintf("\n## Synthesis\n\n%s\n", result.Summary))
		}
		_ = os.WriteFile(filepath.Join(outputDir, "lattice_framing.md"), []byte(framingBuf.String()), 0644)
	}

	return modelNames
}

// latticeSearch runs `lattice search <query> --json` and returns up to 5 slugs.
func latticeSearch(latticeCmd, query string, verbose bool) []string {
	cmd := exec.Command(latticeCmd, "search", query, "--json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if verbose {
			fmt.Printf("[multiplan]   ✗ Lattice search failed: %s\n", err)
		}
		return nil
	}

	var results []latticeSearchResult
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		if verbose {
			fmt.Printf("[multiplan]   ✗ Lattice search JSON parse failed: %s\n", err)
		}
		return nil
	}

	var slugs []string
	for _, r := range results {
		if r.Slug != "" {
			slugs = append(slugs, r.Slug)
		}
	}
	return slugs
}
