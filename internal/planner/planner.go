package planner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cyperx84/multiplan/internal/config"
	"github.com/cyperx84/multiplan/internal/eval"
	"github.com/cyperx84/multiplan/internal/models"
)

type RunResult struct {
	RunID      string
	OutputDir  string
	Plans      []models.ModelResult
	Debate     string
	FinalPlan  string
}

func Run(cfg *config.Config) (*RunResult, error) {
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

	if cfg.Verbose {
		fmt.Printf("[multiplan] Run ID: %s\n", runID)
		fmt.Printf("[multiplan] Output: %s\n", outputDir)
	}

	// Phase 1: Parallel planning with lens-based prompts
	if cfg.Verbose {
		fmt.Println("[multiplan] Phase 1 — Running models in parallel with lens-based prompts...")
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

			provider, ok := models.GetProvider(id)
			if !ok {
				return
			}

			startTime := time.Now()

			// Get lens-based prompt for this model
			prompt := GetLensPrompt(id, cfg.Task, requirements, constraints)

			timeout := time.Duration(timeoutMs) * time.Millisecond
			plan, err := provider.Plan(ctx, prompt, timeout)
			durationMs := time.Since(startTime).Milliseconds()

			result := models.ModelResult{
				ModelID:    id,
				ModelName:  provider.Name(),
				DurationMs: durationMs,
			}

			if err != nil {
				result.Error = err.Error()
				result.Plan = fmt.Sprintf("[Error: %s]", err.Error())
			} else {
				result.Plan = plan
				// Write plan to disk
				filename := filepath.Join(outputDir, fmt.Sprintf("plan-%s.md", id))
				if err := os.WriteFile(filename, []byte(plan), 0644); err == nil {
					if cfg.Verbose {
						fmt.Printf("[multiplan]   ✓ %s done (%dms)\n", provider.Name(), durationMs)
					}
				}
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			resultChan <- result
		}(modelID)
	}

	// Wait for all models to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Stream progress as models finish
	if cfg.Verbose {
		for range resultChan {
			// Results already printed in goroutine
		}
	} else {
		for range resultChan {
			// Drain channel
		}
	}

	if cfg.Verbose {
		fmt.Println("[multiplan] Phase 1 complete")
	}

	// Phase 1.5: Eval all plans (NEW for v0.2.0)
	if cfg.Verbose {
		fmt.Println("[multiplan] Phase 1.5 — Evaluating plans...")
	}

	evalCase := &eval.EvalCase{
		Task: cfg.Task,
	}

	planScores := make(map[string]float64)
	planTexts := make(map[string]string)

	for _, result := range results {
		if result.Error == "" {
			// Eval this plan (structural only, no LLM judge for speed)
			report, err := eval.EvalPlan(result.Plan, evalCase, &eval.EvalOptions{Judge: ""})
			if err == nil {
				planScores[result.ModelID] = report.OverallScore
				planTexts[result.ModelID] = result.Plan
				if cfg.Verbose {
					fmt.Printf("[multiplan]   %s: %.1f/10\n", result.ModelName, report.OverallScore*10)
				}
			}
		}
	}

	// Phase 2: Cross-examination (debate)
	if cfg.Verbose {
		fmt.Println("[multiplan] Phase 2 — Cross-examination...")
	}

	debateProvider, ok := models.GetProvider(cfg.DebateModel)
	if !ok {
		debateProvider, _ = models.GetProvider("claude")
	}

	debatePrompt := GetDebatePrompt(cfg.Task, planTexts)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	debate, err := debateProvider.Plan(ctx, debatePrompt, timeout)
	if err != nil {
		debate = fmt.Sprintf("[Debate failed: %s]", err.Error())
		if cfg.Verbose {
			fmt.Printf("[multiplan]   ✗ Debate failed: %s\n", err.Error())
		}
	} else {
		if cfg.Verbose {
			fmt.Printf("[multiplan]   ✓ Debate complete (via %s)\n", debateProvider.Name())
		}
	}

	if err := os.WriteFile(filepath.Join(outputDir, "debate.md"), []byte(debate), 0644); err != nil {
		return nil, err
	}

	// Phase 3: Convergence with eval scores
	if cfg.Verbose {
		fmt.Println("[multiplan] Phase 3 — Convergence with eval scores...")
	}

	convergeProvider, ok := models.GetProvider(cfg.ConvergeModel)
	if !ok {
		convergeProvider, _ = models.GetProvider("claude")
	}

	convergePrompt := GetConvergePrompt(cfg.Task, planTexts, debate, planScores)

	finalPlan, err := convergeProvider.Plan(ctx, convergePrompt, timeout)
	if err != nil {
		finalPlan = fmt.Sprintf("[Convergence failed: %s]", err.Error())
		if cfg.Verbose {
			fmt.Printf("[multiplan]   ✗ Convergence failed: %s\n", err.Error())
		}
	} else {
		if cfg.Verbose {
			fmt.Printf("[multiplan]   ✓ Convergence complete (via %s)\n", convergeProvider.Name())
		}
	}

	// Add metadata header
	header := fmt.Sprintf("# Multimodel Plan: %s\n\n> Generated: %s\n> Models: ", cfg.Task, time.Now().Format(time.RFC3339))
	modelNames := []string{}
	for _, r := range results {
		modelNames = append(modelNames, r.ModelName)
	}
	header += fmt.Sprintf("%v\n\n---\n\n", modelNames)

	fullPlan := header + finalPlan

	if err := os.WriteFile(filepath.Join(outputDir, "final-plan.md"), []byte(fullPlan), 0644); err != nil {
		return nil, err
	}

	if cfg.Verbose {
		fmt.Println("[multiplan] Phase 3 complete")
	}

	return &RunResult{
		RunID:      runID,
		OutputDir:  outputDir,
		Plans:      results,
		Debate:     debate,
		FinalPlan:  fullPlan,
	}, nil
}

func generateRunID() string {
	now := time.Now()
	timestamp := now.Format("20060102T150405")
	return fmt.Sprintf("%s-%06d", timestamp, now.Nanosecond()/1000)
}
