package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	payload := map[string]interface{}{
		"model": "glm-5",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  8192,
		"temperature": 0.7,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", 0, 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.z.ai/api/coding/paas/v4/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", 0, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, 0, fmt.Errorf("GLM-5 API error: %s - %s", resp.Status, string(body))
	}

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

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, 0, err
	}

	if len(result.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("no content in response")
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
		return "", fmt.Errorf("could not get home directory")
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

	return "", fmt.Errorf("ZAI API key not found. Set ZAI_API_KEY env var or ensure ~/.openclaw/agents/main/agent/auth-profiles.json has zai:default profile")
}
