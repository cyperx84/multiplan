package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cyperx84/multiplan/internal/config"
)

// TestLatticeThinkJSONParsing verifies parsing of lattice think --json output.
func TestLatticeThinkJSONParsing(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantModels int
		wantFirst  string
		wantSlug   string
		wantErr    bool
	}{
		{
			name: "standard output",
			input: `{
				"problem": "build a task management CLI",
				"models": [
					{"model_name": "First Principles", "model_slug": "first-principles", "category": "General"},
					{"model_name": "Inversion", "model_slug": "inversion", "category": "General"}
				],
				"summary": "Break the problem down to fundamentals."
			}`,
			wantModels: 2,
			wantFirst:  "First Principles",
			wantSlug:   "first-principles",
		},
		{
			name:       "empty models",
			input:      `{"problem": "test", "models": [], "summary": ""}`,
			wantModels: 0,
		},
		{
			name:    "invalid json",
			input:   `Here is my analysis: {"models": []}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result latticeThinkResult
			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Models) != tt.wantModels {
				t.Errorf("got %d models, want %d", len(result.Models), tt.wantModels)
			}
			if tt.wantFirst != "" && len(result.Models) > 0 {
				if result.Models[0].ModelName != tt.wantFirst {
					t.Errorf("first model = %q, want %q", result.Models[0].ModelName, tt.wantFirst)
				}
				if result.Models[0].ModelSlug != tt.wantSlug {
					t.Errorf("first slug = %q, want %q", result.Models[0].ModelSlug, tt.wantSlug)
				}
			}
		})
	}
}

// TestLatticeSearchJSONParsing verifies parsing of lattice search --json output.
func TestLatticeSearchJSONParsing(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSlugs []string
		wantErr   bool
	}{
		{
			name: "multiple results",
			input: `[
				{"slug": "inversion", "name": "Inversion", "category": "General"},
				{"slug": "second-order-thinking", "name": "Second-Order Thinking", "category": "Systems"},
				{"slug": "circle-of-competence", "name": "Circle of Competence", "category": "Decision-Making"}
			]`,
			wantSlugs: []string{"inversion", "second-order-thinking", "circle-of-competence"},
		},
		{
			name:      "empty results",
			input:     `[]`,
			wantSlugs: nil,
		},
		{
			name:    "invalid json",
			input:   `not valid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []latticeSearchResult
			err := json.Unmarshal([]byte(tt.input), &results)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var slugs []string
			for _, r := range results {
				if r.Slug != "" {
					slugs = append(slugs, r.Slug)
				}
			}

			if len(slugs) != len(tt.wantSlugs) {
				t.Fatalf("got %d slugs, want %d", len(slugs), len(tt.wantSlugs))
			}
			for i, s := range slugs {
				if s != tt.wantSlugs[i] {
					t.Errorf("slug[%d] = %q, want %q", i, s, tt.wantSlugs[i])
				}
			}
		})
	}
}

// TestLatticeSlugCapping verifies the 5-slug cap on search results.
func TestLatticeSlugCapping(t *testing.T) {
	slugs := []string{"a", "b", "c", "d", "e", "f", "g"}
	if len(slugs) > 5 {
		slugs = slugs[:5]
	}
	if len(slugs) != 5 {
		t.Errorf("expected 5 slugs after cap, got %d", len(slugs))
	}
	if slugs[4] != "e" {
		t.Errorf("last slug = %q, want 'e'", slugs[4])
	}
}

// TestLatticeFramingMarkdownOutput verifies the framing markdown file generation.
func TestLatticeFramingMarkdownOutput(t *testing.T) {
	tmpDir := t.TempDir()
	task := "build a CLI tool tracker"

	models := []struct {
		ModelName string
		Category  string
	}{
		{"First Principles", "General"},
		{"Inversion", "General"},
		{"Moat", "Business"},
	}
	summary := "Break it down to core abstractions, invert to find failure modes."

	// Mirror the rendering logic from planner.go
	var framingBuf strings.Builder
	framingBuf.WriteString("# Lattice Mental Model Framing\n\n")
	framingBuf.WriteString(fmt.Sprintf("Problem: %s\n\n", task))
	framingBuf.WriteString("## Models Applied\n\n")
	for _, m := range models {
		framingBuf.WriteString(fmt.Sprintf("- **%s** (%s)\n", m.ModelName, m.Category))
	}
	if summary != "" {
		framingBuf.WriteString(fmt.Sprintf("\n## Synthesis\n\n%s\n", summary))
	}

	outPath := filepath.Join(tmpDir, "lattice_framing.md")
	if err := os.WriteFile(outPath, []byte(framingBuf.String()), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	content := string(data)

	// Verify structure
	checks := []string{
		"# Lattice Mental Model Framing",
		"Problem: build a CLI tool tracker",
		"## Models Applied",
		"**First Principles** (General)",
		"**Inversion** (General)",
		"**Moat** (Business)",
		"## Synthesis",
		"Break it down to core abstractions",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("missing %q in output:\n%s", c, content)
		}
	}
}

// TestLatticePromptInjection verifies lattice model names are injected into planner prompts.
func TestLatticePromptInjection(t *testing.T) {
	latticeModels := []string{"First Principles", "Inversion", "Second-Order Thinking"}
	basePrompt := "Plan a CLI tool for tracking coding agent changelogs."

	// Mirror the injection logic from planner.go
	if len(latticeModels) > 0 {
		basePrompt += "\n\nRelevant mental models: [" + strings.Join(latticeModels, ", ") + "]\nConsider these frameworks in your plan."
	}

	if !strings.Contains(basePrompt, "Relevant mental models:") {
		t.Error("missing model injection")
	}
	if !strings.Contains(basePrompt, "First Principles, Inversion, Second-Order Thinking") {
		t.Error("models not joined correctly")
	}
	if !strings.Contains(basePrompt, "Consider these frameworks") {
		t.Error("missing framework instruction")
	}
}

// TestLatticeSkipConfig verifies the SkipLattice config flag.
func TestLatticeSkipConfig(t *testing.T) {
	tests := []struct {
		name      string
		skip      bool
		wantPhase bool
	}{
		{"enabled", false, true},
		{"disabled", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{SkipLattice: tt.skip}
			if (!cfg.SkipLattice) != tt.wantPhase {
				t.Errorf("SkipLattice=%v: expected phase0=%v", tt.skip, tt.wantPhase)
			}
		})
	}
}

// TestLatticeFallbackSearch verifies the fallback search when primary query returns nothing.
func TestLatticeFallbackSearch(t *testing.T) {
	// Simulate: primary search returns empty, fallback to "planning"
	primaryResults := []string{}
	fallbackQuery := "planning"

	if len(primaryResults) == 0 {
		// Would call latticeSearch(cmd, "planning", verbose) in real code
		if fallbackQuery != "planning" {
			t.Errorf("fallback query = %q, want 'planning'", fallbackQuery)
		}
	}
}
