package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type GLMProvider struct{}

func (g *GLMProvider) ID() string   { return "glm5" }
func (g *GLMProvider) Name() string { return "GLM-5 (ZhipuAI)" }

func (g *GLMProvider) Available(ctx context.Context) bool {
	_, err := g.getAPIKey()
	return err == nil
}

func (g *GLMProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	text, _, _, err := g.PlanWithTokens(ctx, prompt, timeout)
	return text, err
}

func (g *GLMProvider) PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	apiKey, err := g.getAPIKey()
	if err != nil {
		return "", 0, 0, err
	}

	client := &APIClient{
		BaseURL:      "https://api.z.ai",
		APIKey:       apiKey,
		KeyHeader:    "Authorization",
		KeyPrefix:    "Bearer ",
		ProviderName: "GLM-5",
	}

	payload := map[string]interface{}{
		"model": "glm-5",
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

	if err := client.Post(ctx, "/api/coding/paas/v4/chat/completions", payload, &result); err != nil {
		return "", 0, 0, err
	}

	if len(result.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("no content in GLM-5 response")
	}

	return result.Choices[0].Message.Content, result.Usage.PromptTokens, result.Usage.CompletionTokens, nil
}

func (g *GLMProvider) getAPIKey() (string, error) {
	if key := os.Getenv("ZAI_API_KEY"); key != "" {
		return key, nil
	}
	if key := os.Getenv("GLM_API_KEY"); key != "" {
		return key, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("GLM-5 requires ZAI_API_KEY. Get one at: https://open.bigmodel.cn")
	}

	paths := []string{
		filepath.Join(home, ".openclaw/agents/main/agent/auth-profiles.json"),
		filepath.Join(home, ".openclaw/agents/builder/agent/auth-profiles.json"),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var authProfiles struct {
			Profiles map[string]struct {
				Key string `json:"key"`
			} `json:"profiles"`
		}

		if err := json.Unmarshal(data, &authProfiles); err != nil {
			continue
		}

		if profile, ok := authProfiles.Profiles["zai:default"]; ok && profile.Key != "" {
			return profile.Key, nil
		}
	}

	return "", fmt.Errorf("GLM-5 requires ZAI_API_KEY. Get one at: https://open.bigmodel.cn")
}
