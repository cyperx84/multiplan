package planner

import (
	"strings"
	"testing"
)

func TestGetLensPrompt(t *testing.T) {
	task := "Design a rate limiting system"
	requirements := "Must support 10k users"
	constraints := "Redis only"

	tests := []struct {
		name      string
		modelID   string
		wantLens  string
	}{
		{"claude lens", "claude", "correctness and edge cases"},
		{"gemini lens", "gemini", "scale and operational simplicity"},
		{"codex lens", "codex", "what ships fastest"},
		{"glm5 lens", "glm5", "what are all the ways this could fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := GetLensPrompt(tt.modelID, task, requirements, constraints)

			if !strings.Contains(prompt, tt.wantLens) {
				t.Errorf("expected prompt to contain lens '%s'", tt.wantLens)
			}

			if !strings.Contains(prompt, task) {
				t.Error("expected prompt to contain task")
			}

			if !strings.Contains(prompt, requirements) {
				t.Error("expected prompt to contain requirements")
			}

			if !strings.Contains(prompt, constraints) {
				t.Error("expected prompt to contain constraints")
			}
		})
	}
}

func TestGetDebatePrompt(t *testing.T) {
	task := "Design a rate limiting system"
	plans := map[string]string{
		"claude": "Use Redis sorted sets",
		"gemini": "Use token bucket",
		"codex":  "Use middleware",
		"glm5":   "Consider failure modes",
	}

	prompt := GetDebatePrompt(task, plans)

	if !strings.Contains(prompt, task) {
		t.Error("expected prompt to contain task")
	}

	if !strings.Contains(prompt, "Plan A") {
		t.Error("expected prompt to contain Plan A")
	}

	if !strings.Contains(prompt, "Agreements") {
		t.Error("expected prompt to contain Agreements section")
	}
}

func TestGetLensPrompt_UnknownModel(t *testing.T) {
	prompt := GetLensPrompt("nonexistent", "task", "reqs", "cons")
	if !strings.Contains(prompt, "complete, actionable technical plan") {
		t.Error("unknown model should get default lens")
	}
	if !strings.Contains(prompt, "task") {
		t.Error("prompt should still contain the task")
	}
}

func TestGetLensPrompt_EmptyFields(t *testing.T) {
	prompt := GetLensPrompt("claude", "", "", "")
	if !strings.Contains(prompt, "correctness and edge cases") {
		t.Error("lens should still be present with empty fields")
	}
	if !strings.Contains(prompt, "## Task") {
		t.Error("template structure should be preserved")
	}
}

func TestGetDebatePrompt_SinglePlan(t *testing.T) {
	plans := map[string]string{
		"claude": "Use Redis",
	}
	prompt := GetDebatePrompt("rate limiter", plans)
	if !strings.Contains(prompt, "1 independent") {
		t.Error("should mention 1 plan")
	}
	if !strings.Contains(prompt, "Plan A") {
		t.Error("single plan should be labeled A")
	}
	if strings.Contains(prompt, "Plan B") {
		t.Error("should not contain Plan B with one plan")
	}
}

func TestGetDebatePrompt_ManyPlans(t *testing.T) {
	plans := map[string]string{
		"claude": "plan1",
		"gemini": "plan2",
		"codex":  "plan3",
		"glm5":   "plan4",
	}
	prompt := GetDebatePrompt("task", plans)
	if !strings.Contains(prompt, "4 independent") {
		t.Error("should mention 4 plans")
	}
	// All four plan labels should appear
	for _, label := range []string{"Plan A", "Plan B", "Plan C", "Plan D"} {
		if !strings.Contains(prompt, label) {
			t.Errorf("expected %s in debate prompt", label)
		}
	}
}

func TestGetConvergePrompt_MissingScores(t *testing.T) {
	plans := map[string]string{
		"claude": "plan A",
		"gemini": "plan B",
	}
	// Only one score provided for two plans
	scores := map[string]float64{
		"claude": 0.9,
	}
	prompt := GetConvergePrompt("task", plans, "debate text", scores)
	if !strings.Contains(prompt, "9.0/10") {
		t.Error("should contain the score that exists")
	}
	if !strings.Contains(prompt, "debate text") {
		t.Error("should contain debate")
	}
}

func TestGetConvergePrompt_EmptyScores(t *testing.T) {
	plans := map[string]string{
		"claude": "plan A",
	}
	scores := map[string]float64{}
	prompt := GetConvergePrompt("task", plans, "debate", scores)
	// Should still produce a valid prompt without scores
	if !strings.Contains(prompt, "Final Architecture Decision") {
		t.Error("output format should still be present")
	}
}

func TestGetConvergePrompt(t *testing.T) {
	task := "Design a rate limiting system"
	plans := map[string]string{
		"claude": "Use Redis sorted sets",
		"gemini": "Use token bucket",
	}
	debate := "Plans agree on Redis"
	scores := map[string]float64{
		"claude": 0.85,
		"gemini": 0.78,
	}

	prompt := GetConvergePrompt(task, plans, debate, scores)

	if !strings.Contains(prompt, task) {
		t.Error("expected prompt to contain task")
	}

	if !strings.Contains(prompt, debate) {
		t.Error("expected prompt to contain debate")
	}

	if !strings.Contains(prompt, "8.5/10") {
		t.Error("expected prompt to contain claude score")
	}

	if !strings.Contains(prompt, "Final Architecture Decision") {
		t.Error("expected prompt to contain convergence output format")
	}
}
