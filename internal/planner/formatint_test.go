package planner

import "testing"

func TestFormatInt(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{1234567, "1,234,567"},
	}

	for _, tt := range tests {
		got := formatInt(tt.n)
		if got != tt.want {
			t.Errorf("formatInt(%d) = %s, want %s", tt.n, got, tt.want)
		}
	}
}
