package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		return "", 0, 0, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	payload := map[string]interface{}{
		"model":      "claude-opus-4-20250514",
		"max_tokens": 8192,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", 0, 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", 0, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, 0, fmt.Errorf("anthropic API error: %s - %s", resp.Status, string(body))
	}

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

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, 0, err
	}

	if len(result.Content) == 0 {
		return "", 0, 0, fmt.Errorf("no content in response")
	}

	return result.Content[0].Text, result.Usage.InputTokens, result.Usage.OutputTokens, nil
}
