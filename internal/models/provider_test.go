package models

import "testing"

func TestEstimateCost(t *testing.T) {
	tests := []struct {
		modelID      string
		inputTokens  int
		outputTokens int
		wantMin      float64
		wantMax      float64
	}{
		{"claude", 1_000_000, 1_000_000, 89.9, 90.1},  // 15 + 75
		{"gemini", 1_000_000, 1_000_000, 6.2, 6.3},    // 1.25 + 5.0
		{"codex", 1_000_000, 1_000_000, 12.4, 12.6},   // 2.50 + 10.0
		{"glm5", 1_000_000, 1_000_000, 2.9, 3.1},      // 1.0 + 2.0
		{"unknown", 1_000_000, 1_000_000, 0, 0},
	}

	for _, tt := range tests {
		got := EstimateCost(tt.modelID, tt.inputTokens, tt.outputTokens)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("EstimateCost(%s, %d, %d) = %.4f, want [%.4f, %.4f]",
				tt.modelID, tt.inputTokens, tt.outputTokens, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestGetProvider(t *testing.T) {
	for _, id := range []string{"claude", "gemini", "codex", "glm5"} {
		p, ok := GetProvider(id)
		if !ok {
			t.Errorf("expected provider for %s", id)
		}
		if p.ID() != id {
			t.Errorf("provider.ID() = %s, want %s", p.ID(), id)
		}
	}

	_, ok := GetProvider("nonexistent")
	if ok {
		t.Error("expected no provider for 'nonexistent'")
	}
}

func TestProviderImplementsTokens(t *testing.T) {
	for _, id := range []string{"claude", "gemini", "codex", "glm5"} {
		p, _ := GetProvider(id)
		if _, ok := p.(ProviderWithTokens); !ok {
			t.Errorf("provider %s does not implement ProviderWithTokens", id)
		}
	}
}
