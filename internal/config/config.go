package config

import (
	"context"
	"time"
)

// Provider mirrors models.Provider to avoid import cycles.
type Provider interface {
	ID() string
	Name() string
	Available(ctx context.Context) bool
	Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error)
}

// Config holds all runtime configuration for a multiplan run.
type Config struct {
	Task          string
	Requirements  string
	Constraints   string
	Models        []string
	OutputDir     string
	DebateModel   string
	ConvergeModel string
	TimeoutMs     int
	Verbose       bool
	Quiet         bool
	JSON          bool
	SkipLattice   bool
	LatticeCmd    string
	// ProviderFactory is optional. If set, the planner calls it to get providers
	// instead of the global registry. Used in tests to inject mock providers.
	ProviderFactory func(id string) (Provider, bool)
}

func (c *Config) GetRequirements() string {
	if c.Requirements == "" {
		return "None specified."
	}
	return c.Requirements
}

func (c *Config) GetConstraints() string {
	if c.Constraints == "" {
		return "None specified."
	}
	return c.Constraints
}

func (c *Config) GetTimeoutMs() int {
	if c.TimeoutMs == 0 {
		return 120000
	}
	return c.TimeoutMs
}
