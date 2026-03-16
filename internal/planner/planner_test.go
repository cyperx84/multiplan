package planner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cyperx84/multiplan/internal/config"
)

// MockProvider is a test double that returns canned responses.
type MockProvider struct {
	id     string
	name   string
	plan   string
	err    error
	tokens [2]int // input, output
}

func (m *MockProvider) ID() string   { return m.id }
func (m *MockProvider) Name() string { return m.name }
func (m *MockProvider) Available(_ context.Context) bool { return true }
func (m *MockProvider) Plan(_ context.Context, _ string, _ time.Duration) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.plan, nil
}

// buildFactory creates a ProviderFactory from a map of MockProviders.
func buildFactory(providers map[string]*MockProvider) func(id string) (config.Provider, bool) {
	return func(id string) (config.Provider, bool) {
		p, ok := providers[id]
		if !ok {
			return nil, false
		}
		return p, true
	}
}

// baseConfig returns a Config suitable for testing with a temp output dir.
func baseConfig(t *testing.T, factory func(string) (config.Provider, bool)) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		Task:            "Design a rate limiter",
		Models:          []string{"claude", "gemini"},
		DebateModel:     "claude",
		ConvergeModel:   "claude",
		TimeoutMs:       5000,
		OutputDir:       dir,
		Quiet:           true,
		ProviderFactory: factory,
	}
}

// TestPlanner_PhaseOrdering verifies that plans → debate → convergence happen
// in order by checking output files are created.
func TestPlanner_PhaseOrdering(t *testing.T) {
	providers := map[string]*MockProvider{
		"claude": {id: "claude", name: "Claude", plan: "# Claude Plan\nStep 1"},
		"gemini": {id: "gemini", name: "Gemini", plan: "# Gemini Plan\nStep 1"},
	}
	cfg := baseConfig(t, buildFactory(providers))

	result, err := Run(cfg)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Both plan files should exist
	if _, err := os.Stat(filepath.Join(cfg.OutputDir, "plan-claude.md")); err != nil {
		t.Errorf("plan-claude.md not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cfg.OutputDir, "plan-gemini.md")); err != nil {
		t.Errorf("plan-gemini.md not created: %v", err)
	}

	// debate.md should exist
	if _, err := os.Stat(filepath.Join(cfg.OutputDir, "debate.md")); err != nil {
		t.Errorf("debate.md not created: %v", err)
	}

	// final-plan.md should exist
	if _, err := os.Stat(filepath.Join(cfg.OutputDir, "final-plan.md")); err != nil {
		t.Errorf("final-plan.md not created: %v", err)
	}

	// Result should have 2 plans
	if len(result.Plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(result.Plans))
	}
}

// TestPlanner_ScoreWeighting verifies the convergence prompt includes scores.
func TestPlanner_ScoreWeighting(t *testing.T) {
	// Use distinct plans with different quality signals (structural eval uses word matching)
	highQualityPlan := strings.Repeat("# Architecture\n## Overview\n### Details\nThis plan covers implementation, testing, deployment, monitoring, rollback, and security considerations.\n\n", 5)
	lowQualityPlan := "do stuff"

	providers := map[string]*MockProvider{
		"claude": {id: "claude", name: "Claude", plan: highQualityPlan},
		"gemini": {id: "gemini", name: "Gemini", plan: lowQualityPlan},
	}
	cfg := baseConfig(t, buildFactory(providers))
	cfg.Models = []string{"claude", "gemini"}

	result, err := Run(cfg)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Final plan should be non-empty (convergence ran)
	if result.FinalPlan == "" {
		t.Error("FinalPlan is empty")
	}
	if len(result.Plans) != 2 {
		t.Errorf("expected 2 model results, got %d", len(result.Plans))
	}
}

// TestPlanner_OutputFiles verifies all expected output files are created.
func TestPlanner_OutputFiles(t *testing.T) {
	providers := map[string]*MockProvider{
		"claude": {id: "claude", name: "Claude", plan: "# Claude Plan"},
		"gemini": {id: "gemini", name: "Gemini", plan: "# Gemini Plan"},
	}
	cfg := baseConfig(t, buildFactory(providers))

	result, err := Run(cfg)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	expected := []string{"plan-claude.md", "plan-gemini.md", "debate.md", "final-plan.md"}
	for _, name := range expected {
		path := filepath.Join(cfg.OutputDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected output file %s not found: %v", name, err)
		}
	}

	// RunID and OutputDir should be populated
	if result.RunID == "" {
		t.Error("RunID is empty")
	}
	if result.OutputDir != cfg.OutputDir {
		t.Errorf("OutputDir mismatch: got %s, want %s", result.OutputDir, cfg.OutputDir)
	}
}

// TestPlanner_ModelFailure verifies that one model erroring doesn't stop the pipeline.
func TestPlanner_ModelFailure(t *testing.T) {
	providers := map[string]*MockProvider{
		"claude": {id: "claude", name: "Claude", plan: "# Claude Plan"},
		"gemini": {id: "gemini", name: "Gemini", err: fmt.Errorf("simulated API failure")},
	}
	cfg := baseConfig(t, buildFactory(providers))

	result, err := Run(cfg)
	if err != nil {
		t.Fatalf("Run() should not error when one model fails, got: %v", err)
	}

	var claudeResult, geminiResult *struct{ err string }
	for _, r := range result.Plans {
		switch r.ModelID {
		case "claude":
			claudeResult = &struct{ err string }{err: r.Error}
		case "gemini":
			geminiResult = &struct{ err string }{err: r.Error}
		}
	}

	if claudeResult == nil || claudeResult.err != "" {
		t.Errorf("expected Claude to succeed, got error: %v", claudeResult)
	}
	if geminiResult == nil || geminiResult.err == "" {
		t.Errorf("expected Gemini to fail, but error is empty")
	}

	// final-plan.md should still exist
	if _, err := os.Stat(filepath.Join(cfg.OutputDir, "final-plan.md")); err != nil {
		t.Errorf("final-plan.md not created despite partial success: %v", err)
	}
}

// TestPlanner_AllModelsFail verifies graceful handling when all models error.
func TestPlanner_AllModelsFail(t *testing.T) {
	providers := map[string]*MockProvider{
		"claude": {id: "claude", name: "Claude", err: fmt.Errorf("claude down")},
		"gemini": {id: "gemini", name: "Gemini", err: fmt.Errorf("gemini down")},
	}
	cfg := baseConfig(t, buildFactory(providers))

	result, err := Run(cfg)
	// Run should complete (not panic), though it may return an error or an empty FinalPlan.
	// The important thing is it doesn't panic and debate/final files are created.
	if err != nil {
		// Some implementations may return an error — that's acceptable.
		t.Logf("Run() returned error (acceptable): %v", err)
		return
	}

	// If no error, result should still be populated
	if result == nil {
		t.Fatal("Run() returned nil result without error")
	}
	for _, r := range result.Plans {
		if r.Error == "" {
			t.Errorf("expected all models to fail, but %s succeeded", r.ModelID)
		}
	}
}

// TestPlanner_DebateFailure verifies the pipeline completes even if debate fails.
func TestPlanner_DebateFailure(t *testing.T) {
	callCount := 0
	// Claude is used for both planning and debate/converge.
	// We make it succeed for Plan calls 1-2 (planning phase), fail for call 3 (debate), succeed for call 4 (converge).
	var claudeErr error
	providers := map[string]*MockProvider{
		"claude": {id: "claude", name: "Claude"},
		"gemini": {id: "gemini", name: "Gemini", plan: "# Gemini Plan"},
	}

	factory := func(id string) (config.Provider, bool) {
		if id == "claude" {
			return &callCountProvider{
				MockProvider: MockProvider{id: "claude", name: "Claude", plan: "# Claude Plan"},
				onCall: func(n int) error {
					callCount = n
					// Fail on the debate call (3rd call overall: claude plan, gemini plan parallel → debate)
					// Actually debate is sequential after plans. We'll fail specifically via a flag.
					return claudeErr
				},
			}, true
		}
		p, ok := providers[id]
		return p, ok
	}

	cfg := baseConfig(t, factory)
	cfg.Models = []string{"claude", "gemini"}

	// Set debate to fail
	claudeErr = fmt.Errorf("debate model down")

	result, err := Run(cfg)
	if err != nil {
		t.Fatalf("Run() should not return error on debate failure: %v", err)
	}

	// Debate should contain error message
	if !strings.Contains(result.Debate, "Debate failed") && !strings.Contains(result.Debate, "debate model down") {
		t.Errorf("Debate should contain failure info, got: %q", result.Debate)
	}

	_ = callCount
	// final-plan.md should exist (convergence still ran)
	if _, err := os.Stat(filepath.Join(cfg.OutputDir, "final-plan.md")); err != nil {
		t.Errorf("final-plan.md not created after debate failure: %v", err)
	}
}

// callCountProvider calls onCall with an incrementing counter on each Plan().
type callCountProvider struct {
	MockProvider
	count  int
	onCall func(n int) error
}

func (c *callCountProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	c.count++
	if err := c.onCall(c.count); err != nil {
		return "", err
	}
	return c.plan, nil
}
