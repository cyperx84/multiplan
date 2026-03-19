package models

import "testing"

func TestCodexProvider_IDAndName(t *testing.T) {
	p := &CodexProvider{}
	if p.ID() != "codex" {
		t.Errorf("ID() = %q, want 'codex'", p.ID())
	}
	if p.Name() != "Codex (GPT)" {
		t.Errorf("Name() = %q, want 'Codex (GPT)'", p.Name())
	}
}

func TestCodexProvider_CliCmd(t *testing.T) {
	// Default
	p := &CodexProvider{}
	if cmd := p.cliCmd(); cmd != "codex" {
		t.Errorf("default cliCmd() = %q, want 'codex'", cmd)
	}

	// Override
	p = &CodexProvider{CodexCmd: "/usr/local/bin/my-codex"}
	if cmd := p.cliCmd(); cmd != "/usr/local/bin/my-codex" {
		t.Errorf("overridden cliCmd() = %q, want '/usr/local/bin/my-codex'", cmd)
	}
}

func TestClaudeProvider_CliCmd(t *testing.T) {
	// Default
	p := &ClaudeProvider{}
	if cmd := p.cliCmd(); cmd != "claude" {
		t.Errorf("default cliCmd() = %q, want 'claude'", cmd)
	}

	// Override
	p = &ClaudeProvider{ClaudeCmd: "/opt/claude"}
	if cmd := p.cliCmd(); cmd != "/opt/claude" {
		t.Errorf("overridden cliCmd() = %q, want '/opt/claude'", cmd)
	}
}
