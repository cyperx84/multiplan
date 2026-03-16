package config

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
