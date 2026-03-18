package models

import (
	"context"
	"fmt"
	"os"
	"time"
)

type CodexProvider struct{}

func (c *CodexProvider) ID() string   { return "codex" }
func (c *CodexProvider) Name() string { return "Codex (GPT)" }

func (c *CodexProvider) Available(ctx context.Context) bool {
	return os.Getenv("OPENAI_API_KEY") != ""
}

func (c *CodexProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	text, _, _, err := c.PlanWithTokens(ctx, prompt, timeout)
	return text, err
}

func (c *CodexProvider) PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", 0, 0, fmt.Errorf("OpenAI requires OPENAI_API_KEY (skipped — set key to enable)")
	}

	client := &APIClient{
		BaseURL:      "https://api.openai.com",
		APIKey:       apiKey,
		KeyHeader:    "Authorization",
		KeyPrefix:    "Bearer ",
		ProviderName: "OpenAI",
	}

	payload := map[string]interface{}{
		"model": "gpt-4o",
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
