package config

import "testing"

func TestConfig_GetRequirements(t *testing.T) {
	tests := []struct {
		name string
		req  string
		want string
	}{
		{"with requirements", "Must support 10k users", "Must support 10k users"},
		{"empty requirements", "", "None specified."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Requirements: tt.req}
			if got := c.GetRequirements(); got != tt.want {
				t.Errorf("GetRequirements() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetConstraints(t *testing.T) {
	tests := []struct {
		name string
		con  string
		want string
	}{
		{"with constraints", "Redis only", "Redis only"},
		{"empty constraints", "", "None specified."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Constraints: tt.con}
			if got := c.GetConstraints(); got != tt.want {
				t.Errorf("GetConstraints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetTimeoutMs(t *testing.T) {
	tests := []struct {
		name      string
		timeoutMs int
		want      int
	}{
		{"with timeout", 60000, 60000},
		{"zero timeout defaults to 120000", 0, 120000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{TimeoutMs: tt.timeoutMs}
			if got := c.GetTimeoutMs(); got != tt.want {
				t.Errorf("GetTimeoutMs() = %v, want %v", got, tt.want)
			}
		})
	}
}
