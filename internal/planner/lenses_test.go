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
