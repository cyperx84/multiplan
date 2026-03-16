package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileConfig_NotFound(t *testing.T) {
	// Change to a temp dir with no config file
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	fc, err := LoadFileConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fc != nil {
		t.Fatalf("expected nil config, got %+v", fc)
	}
}

func TestLoadFileConfig_LocalFile(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	content := `
models: [claude, gemini]
debate_model: gemini
converge_model: gemini
timeout_ms: 60000
output_dir: /tmp/runs
requirements: must be fast
constraints: no databases
`
	if err := os.WriteFile(filepath.Join(dir, ".multiplan.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fc, err := LoadFileConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fc == nil {
		t.Fatal("expected config, got nil")
	}
	if len(fc.Models) != 2 || fc.Models[0] != "claude" {
		t.Errorf("unexpected models: %v", fc.Models)
	}
	if fc.DebateModel != "gemini" {
		t.Errorf("unexpected debate_model: %s", fc.DebateModel)
	}
	if fc.TimeoutMs != 60000 {
		t.Errorf("unexpected timeout_ms: %d", fc.TimeoutMs)
	}
}

func TestApplyFileConfig_CLIOverrides(t *testing.T) {
	cfg := &Config{
		DebateModel: "claude", // explicitly set via CLI
	}
	fc := &FileConfig{
		DebateModel:   "gemini",
		ConvergeModel: "gemini",
		TimeoutMs:     60000,
	}
	ApplyFileConfig(cfg, fc)

	// CLI value should NOT be overridden
	if cfg.DebateModel != "claude" {
		t.Errorf("CLI flag overridden by config file: got %s", cfg.DebateModel)
	}
	// Unset field should be filled
	if cfg.ConvergeModel != "gemini" {
		t.Errorf("expected converge_model=gemini, got %s", cfg.ConvergeModel)
	}
	if cfg.TimeoutMs != 60000 {
		t.Errorf("expected timeout_ms=60000, got %d", cfg.TimeoutMs)
	}
}

func TestApplyFileConfig_Nil(t *testing.T) {
	cfg := &Config{DebateModel: "claude"}
	ApplyFileConfig(cfg, nil) // should not panic
	if cfg.DebateModel != "claude" {
		t.Error("config changed after nil ApplyFileConfig")
	}
}
