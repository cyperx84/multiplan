package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseClaudeJSONOutput_PlanMode(t *testing.T) {
	// permission_denials with plan text → should extract plan
	data := mustJSON(t, claudeJSONResult{
		Result: "I'll create a plan for you.",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{InputTokens: 100, OutputTokens: 200},
		PermissionDenials: []struct {
			ToolInput struct {
				Plan string `json:"plan"`
			} `json:"tool_input"`
		}{
			{ToolInput: struct {
				Plan string `json:"plan"`
			}{Plan: "## Overview\nBuild a REST API with auth"}},
		},
	})

	text, inTok, outTok, err := parseClaudeJSONOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "## Overview\nBuild a REST API with auth" {
		t.Errorf("got text %q, want plan from permission_denials", text)
	}
	if inTok != 100 || outTok != 200 {
		t.Errorf("tokens = (%d, %d), want (100, 200)", inTok, outTok)
	}
}

func TestParseClaudeJSONOutput_MultiplePermissionDenials(t *testing.T) {
	// Multiple denials — should use first non-empty plan
	data := mustJSON(t, claudeJSONResult{
		Result: "ignored",
		PermissionDenials: []struct {
			ToolInput struct {
				Plan string `json:"plan"`
			} `json:"tool_input"`
		}{
			{ToolInput: struct {
				Plan string `json:"plan"`
			}{Plan: "first plan text"}},
			{ToolInput: struct {
				Plan string `json:"plan"`
			}{Plan: "second plan text"}},
		},
	})

	text, _, _, err := parseClaudeJSONOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "first plan text" {
		t.Errorf("got %q, want first denial's plan", text)
	}
}

func TestParseClaudeJSONOutput_EmptyDenialFallsThrough(t *testing.T) {
	// permission_denials present but plan is empty → fall through to result
	data := mustJSON(t, claudeJSONResult{
		Result: "the actual result",
		PermissionDenials: []struct {
			ToolInput struct {
				Plan string `json:"plan"`
			} `json:"tool_input"`
		}{
			{ToolInput: struct {
				Plan string `json:"plan"`
			}{Plan: ""}},
		},
	})

	text, _, _, err := parseClaudeJSONOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "the actual result" {
		t.Errorf("got %q, want 'the actual result'", text)
	}
}

func TestParseClaudeJSONOutput_NormalResult(t *testing.T) {
	data := mustJSON(t, claudeJSONResult{
		Result: "Here is your plan:\n## Overview\nDo the thing",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{InputTokens: 500, OutputTokens: 1000},
	})

	text, inTok, outTok, err := parseClaudeJSONOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "## Overview") {
		t.Errorf("expected result text, got %q", text)
	}
	if inTok != 500 || outTok != 1000 {
		t.Errorf("tokens = (%d, %d), want (500, 1000)", inTok, outTok)
	}
}

func TestParseClaudeJSONOutput_EmptyResult(t *testing.T) {
	data := mustJSON(t, claudeJSONResult{
		Result: "",
	})

	_, _, _, err := parseClaudeJSONOutput(data)
	if err == nil {
		t.Fatal("expected error for empty result")
	}
	if !strings.Contains(err.Error(), "empty result") {
		t.Errorf("error = %q, want 'empty result' mention", err.Error())
	}
}

func TestParseClaudeJSONOutput_WhitespaceOnlyResult(t *testing.T) {
	data := mustJSON(t, claudeJSONResult{
		Result: "   \n\t  ",
	})

	_, _, _, err := parseClaudeJSONOutput(data)
	if err == nil {
		t.Fatal("expected error for whitespace-only result")
	}
}

func TestParseClaudeJSONOutput_NonJSONFallback(t *testing.T) {
	// Plain text output (no JSON) → return as-is
	text, inTok, outTok, err := parseClaudeJSONOutput([]byte("Just a plain text response"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Just a plain text response" {
		t.Errorf("got %q, want plain text", text)
	}
	if inTok != 0 || outTok != 0 {
		t.Errorf("tokens should be 0 for non-JSON, got (%d, %d)", inTok, outTok)
	}
}

func TestParseClaudeJSONOutput_EmptyInput(t *testing.T) {
	_, _, _, err := parseClaudeJSONOutput([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if !strings.Contains(err.Error(), "empty output") {
		t.Errorf("error = %q, want 'empty output' mention", err.Error())
	}
}

func TestParseClaudeJSONOutput_WhitespaceOnlyInput(t *testing.T) {
	_, _, _, err := parseClaudeJSONOutput([]byte("   \n\t  "))
	if err == nil {
		t.Fatal("expected error for whitespace-only input")
	}
}

func TestParseClaudeJSONOutput_LeadingWhitespace(t *testing.T) {
	inner := mustJSON(t, claudeJSONResult{
		Result: "plan after whitespace",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{InputTokens: 10, OutputTokens: 20},
	})
	// Prepend whitespace/newlines
	data := append([]byte("\n\n  \t"), inner...)

	text, inTok, outTok, err := parseClaudeJSONOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "plan after whitespace" {
		t.Errorf("got %q, want 'plan after whitespace'", text)
	}
	if inTok != 10 || outTok != 20 {
		t.Errorf("tokens = (%d, %d), want (10, 20)", inTok, outTok)
	}
}

func TestParseClaudeJSONOutput_InvalidJSON(t *testing.T) {
	// Starts with { but isn't valid JSON → fallback to raw text
	data := []byte(`{not valid json at all}`)
	text, _, _, err := parseClaudeJSONOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "{not valid json at all}" {
		t.Errorf("got %q, want raw text fallback", text)
	}
}

func TestParseClaudeJSONOutput_InvalidJSONEmptyAfterTrim(t *testing.T) {
	// Starts with { but content after trim is just whitespace around bad JSON
	// Actually the { means it tries to parse, fails, then falls back to trimmed string
	data := []byte("  {broken")
	text, _, _, err := parseClaudeJSONOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "{broken") {
		t.Errorf("got %q, want fallback containing '{broken'", text)
	}
}

// mustJSON marshals v to JSON bytes, failing the test on error.
func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	return data
}
