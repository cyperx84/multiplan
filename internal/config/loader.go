package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the YAML config file structure.
type FileConfig struct {
	Models        []string `yaml:"models"`
	DebateModel   string   `yaml:"debate_model"`
	ConvergeModel string   `yaml:"converge_model"`
	TimeoutMs     int      `yaml:"timeout_ms"`
	OutputDir     string   `yaml:"output_dir"`
	Requirements  string   `yaml:"requirements"`
	Constraints   string   `yaml:"constraints"`
	ClaudeCmd     string   `yaml:"claude_cmd"`
	ClaudeModel   string   `yaml:"claude_model"`
}

// LoadFileConfig tries to load .multiplan.yml from:
//  1. Current directory
//  2. $HOME/.config/multiplan/config.yml
//
// Returns nil (no error) if no config file is found — defaults apply.
func LoadFileConfig() (*FileConfig, error) {
	candidates := []string{
		".multiplan.yml",
	}

	home, err := os.UserHomeDir()
	if err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "multiplan", "config.yml"))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		var fc FileConfig
		if err := yaml.Unmarshal(data, &fc); err != nil {
			return nil, err
		}

		// Expand ~ in output_dir
		if strings.HasPrefix(fc.OutputDir, "~/") && home != "" {
			fc.OutputDir = filepath.Join(home, fc.OutputDir[2:])
		}

		return &fc, nil
	}

	return nil, nil
}

// ApplyFileConfig applies file config values to a Config, but only for
// fields that haven't been explicitly set via CLI flags (zero values).
func ApplyFileConfig(cfg *Config, fc *FileConfig) {
	if fc == nil {
		return
	}
	if len(cfg.Models) == 0 && len(fc.Models) > 0 {
		cfg.Models = fc.Models
	}
	if cfg.DebateModel == "" && fc.DebateModel != "" {
		cfg.DebateModel = fc.DebateModel
	}
	if cfg.ConvergeModel == "" && fc.ConvergeModel != "" {
		cfg.ConvergeModel = fc.ConvergeModel
	}
	if cfg.TimeoutMs == 0 && fc.TimeoutMs > 0 {
		cfg.TimeoutMs = fc.TimeoutMs
	}
	if cfg.OutputDir == "" && fc.OutputDir != "" {
		cfg.OutputDir = fc.OutputDir
	}
	if cfg.Requirements == "" && fc.Requirements != "" {
		cfg.Requirements = fc.Requirements
	}
	if cfg.Constraints == "" && fc.Constraints != "" {
		cfg.Constraints = fc.Constraints
	}
	if cfg.ClaudeCmd == "" && fc.ClaudeCmd != "" {
		cfg.ClaudeCmd = fc.ClaudeCmd
	}
	if cfg.ClaudeModel == "" && fc.ClaudeModel != "" {
		cfg.ClaudeModel = fc.ClaudeModel
	}
}
