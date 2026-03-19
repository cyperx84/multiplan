package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ClaudeProvider struct {
	// ClaudeCmd overrides the CLI binary path (default: "claude").
	ClaudeCmd string
	// ClaudeModel overrides the model for CLI mode (e.g. "claude-opus-4-20250514").
	ClaudeModel string
}

func (c *ClaudeProvider) ID() string   { return "claude" }
func (c *ClaudeProvider) Name() string { return "Claude (Opus)" }

func (c *ClaudeProvider) Available(ctx context.Context) bool {
	// Mode A: API key
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return true
	}
	// Mode B: CLI binary on PATH
	cmd := c.cliCmd()
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (c *ClaudeProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	text, _, _, err := c.PlanWithTokens(ctx, prompt, timeout)
	return text, err
}

func (c *ClaudeProvider) PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		return c.planViaAPI(ctx, apiKey, prompt, timeout)
	}
	return c.planViaCLI(ctx, prompt, timeout)
}

// planViaAPI uses the Anthropic HTTP API (Mode A).
func (c *ClaudeProvider) planViaAPI(ctx context.Context, apiKey, prompt string, timeout time.Duration) (string, int, int, error) {
	client := &APIClient{
		BaseURL:      "https://api.anthropic.com",
		APIKey:       apiKey,
		KeyHeader:    "X-API-Key",
		KeyPrefix:    "",
		ExtraHeaders: map[string]string{"anthropic-version": "2023-06-01"},
		ProviderName: "Claude",
	}

	payload := map[string]interface{}{
		"model":      "claude-opus-4-20250514",
		"max_tokens": 8192,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := client.Post(ctx, "/v1/messages", payload, &result); err != nil {
		return "", 0, 0, err
	}

	if len(result.Content) == 0 {
		return "", 0, 0, fmt.Errorf("no content in Claude response")
	}

	return result.Content[0].Text, result.Usage.InputTokens, result.Usage.OutputTokens, nil
}

// planViaCLI uses `claude --print` subprocess (Mode B).
// Uses --output-format json to reliably extract the plan text, because Claude CLI
// may auto-enter "plan mode" for coding tasks, trapping the actual plan in tool_use
// calls and returning an empty or unhelpful text result.
func (c *ClaudeProvider) planViaCLI(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	cmd := c.cliCmd()
	if _, err := exec.LookPath(cmd); err != nil {
		return "", 0, 0, fmt.Errorf("Claude requires ANTHROPIC_API_KEY or the claude CLI on PATH")
	}

	// CLI mode is slower than direct API — give it 2x the configured timeout.
	cliTimeout := timeout * 2
	if cliTimeout < 10*time.Minute {
		cliTimeout = 10 * time.Minute
	}

	// Use JSON output for reliable parsing. Flags before prompt arg.
	args := []string{"--output-format", "json", "--print", prompt}
	if c.ClaudeModel != "" {
		args = append(args, "--model", c.ClaudeModel)
	}

	ctx, cancel := context.WithTimeout(ctx, cliTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", 0, 0, fmt.Errorf("claude CLI failed: %s", errMsg)
	}

	return parseClaudeJSONOutput(stdout.Bytes())
}

// claudeJSONResult represents the JSON output from `claude --print --output-format json`.
type claudeJSONResult struct {
	Result string `json:"result"`
	Usage  struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	PermissionDenials []struct {
		ToolInput struct {
			Plan string `json:"plan"`
		} `json:"tool_input"`
	} `json:"permission_denials"`
}

// parseClaudeJSONOutput extracts the plan text from Claude CLI JSON output.
// Claude may auto-enter plan mode — the actual plan lives in permission_denials[].tool_input.plan.
func parseClaudeJSONOutput(data []byte) (string, int, int, error) {
	// Find JSON start (output may have leading whitespace/newlines)
	start := bytes.IndexByte(data, '{')
	if start < 0 {
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", 0, 0, fmt.Errorf("claude CLI returned empty output (no JSON)")
		}
		// Fallback: non-JSON output, return as-is
		return text, 0, 0, nil
	}

	var result claudeJSONResult
	if err := json.Unmarshal(data[start:], &result); err != nil {
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", 0, 0, fmt.Errorf("claude CLI returned unparseable JSON: %s", err)
		}
		return text, 0, 0, nil
	}

	// Priority 1: check permission_denials for plan mode plan text
	for _, denial := range result.PermissionDenials {
		if plan := strings.TrimSpace(denial.ToolInput.Plan); plan != "" {
			return plan, result.Usage.InputTokens, result.Usage.OutputTokens, nil
		}
	}

	// Priority 2: use the result field
	text := strings.TrimSpace(result.Result)
	if text == "" {
		return "", 0, 0, fmt.Errorf("claude CLI returned empty result")
	}

	return text, result.Usage.InputTokens, result.Usage.OutputTokens, nil
}

// cliCmd returns the CLI binary name/path.
func (c *ClaudeProvider) cliCmd() string {
	if c.ClaudeCmd != "" {
		return c.ClaudeCmd
	}
	return "claude"
}
