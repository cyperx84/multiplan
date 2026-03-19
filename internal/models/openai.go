package models

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CodexProvider supports two modes:
//   A) OPENAI_API_KEY env var — direct HTTP to OpenAI API
//   B) codex CLI on PATH — uses your ChatGPT/OpenAI subscription
//
// Mode B takes priority when PREFER_CLI=true or no API key is set.
// Mode A is used when OPENAI_API_KEY is present (bring-your-own-key users).
type CodexProvider struct {
	// CodexCmd overrides the CLI binary (default: "codex").
	CodexCmd string
	// CodexModel overrides the model for CLI mode (default: gpt-5.4 / codex default).
	CodexModel string
}

func (c *CodexProvider) ID() string   { return "codex" }
func (c *CodexProvider) Name() string { return "Codex (GPT)" }

func (c *CodexProvider) Available(ctx context.Context) bool {
	// API key is available
	if os.Getenv("OPENAI_API_KEY") != "" {
		return true
	}
	// CLI binary on PATH
	cmd := c.cliCmd()
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (c *CodexProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	text, _, _, err := c.PlanWithTokens(ctx, prompt, timeout)
	return text, err
}

func (c *CodexProvider) PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		return c.planViaAPI(ctx, apiKey, prompt, timeout)
	}
	return c.planViaCLI(ctx, prompt, timeout)
}

// planViaAPI uses the OpenAI HTTP API (bring-your-own-key).
func (c *CodexProvider) planViaAPI(ctx context.Context, apiKey, prompt string, timeout time.Duration) (string, int, int, error) {
	client := &APIClient{
		BaseURL:      "https://api.openai.com",
		APIKey:       apiKey,
		KeyHeader:    "Authorization",
		KeyPrefix:    "Bearer ",
		ProviderName: "OpenAI",
	}

	model := "gpt-4o"
	if c.CodexModel != "" {
		model = c.CodexModel
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  8192,
		"temperature": 0.7,
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := client.Post(ctx, "/v1/chat/completions", payload, &result); err != nil {
		return "", 0, 0, err
	}

	if len(result.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("no content in OpenAI response")
	}

	return result.Choices[0].Message.Content, result.Usage.PromptTokens, result.Usage.CompletionTokens, nil
}

// planViaCLI uses `codex exec` with your OpenAI subscription (no API key needed).
func (c *CodexProvider) planViaCLI(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	cmd := c.cliCmd()
	if _, err := exec.LookPath(cmd); err != nil {
		return "", 0, 0, fmt.Errorf("OpenAI requires OPENAI_API_KEY or the codex CLI on PATH")
	}

	// CLI mode is slower — give it 2x the configured timeout.
	cliTimeout := timeout * 2
	if cliTimeout < 10*time.Minute {
		cliTimeout = 10 * time.Minute
	}

	// Write output to a temp file; codex exec --output-last-message captures final reply.
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("multiplan-codex-%d.txt", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	args := []string{
		"exec",
		"--full-auto",
		"--skip-git-repo-check",
		"-o", tmpFile,
	}
	if c.CodexModel != "" {
		args = append(args, "--model", c.CodexModel)
	}

	ctx, cancel := context.WithTimeout(ctx, cliTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)
	command.Stdin = strings.NewReader(prompt)
	var stderr bytes.Buffer
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", 0, 0, fmt.Errorf("codex CLI failed: %s", errMsg)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", 0, 0, fmt.Errorf("codex CLI: could not read output file: %s", err)
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", 0, 0, fmt.Errorf("codex CLI returned empty output")
	}

	// CLI does not expose token usage
	return text, 0, 0, nil
}

func (c *CodexProvider) cliCmd() string {
	if c.CodexCmd != "" {
		return c.CodexCmd
	}
	return "codex"
}
