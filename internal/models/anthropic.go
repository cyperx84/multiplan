package models

import (
	"context"
	"fmt"
	"os"
	"time"
)

type ClaudeProvider struct{}

func (c *ClaudeProvider) ID() string   { return "claude" }
func (c *ClaudeProvider) Name() string { return "Claude (Opus)" }

func (c *ClaudeProvider) Available(ctx context.Context) bool {
	return os.Getenv("ANTHROPIC_API_KEY") != ""
}

func (c *ClaudeProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	text, _, _, err := c.PlanWithTokens(ctx, prompt, timeout)
	return text, err
}

func (c *ClaudeProvider) PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", 0, 0, fmt.Errorf("Claude requires ANTHROPIC_API_KEY. Set it with: export ANTHROPIC_API_KEY=sk-...")
	}

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
