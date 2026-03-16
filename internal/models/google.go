package models

import (
	"context"
	"fmt"
	"os"
	"time"
)

type GeminiProvider struct{}

func (g *GeminiProvider) ID() string   { return "gemini" }
func (g *GeminiProvider) Name() string { return "Gemini" }

func (g *GeminiProvider) Available(ctx context.Context) bool {
	return os.Getenv("GOOGLE_AI_API_KEY") != "" || os.Getenv("GEMINI_API_KEY") != ""
}

func (g *GeminiProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	text, _, _, err := g.PlanWithTokens(ctx, prompt, timeout)
	return text, err
}

func (g *GeminiProvider) PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return "", 0, 0, fmt.Errorf("Gemini requires GOOGLE_AI_API_KEY or GEMINI_API_KEY. Get one at: https://aistudio.google.com/apikey")
	}

	// Gemini uses key as query param, not header
	path := fmt.Sprintf("/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", apiKey)

	client := &APIClient{
		BaseURL:      "https://generativelanguage.googleapis.com",
		ProviderName: "Gemini",
	}

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": 8192,
			"temperature":     0.7,
		},
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := client.Post(ctx, path, payload, &result); err != nil {
		return "", 0, 0, err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", 0, 0, fmt.Errorf("no content in Gemini response")
	}

	return result.Candidates[0].Content.Parts[0].Text,
		result.UsageMetadata.PromptTokenCount,
		result.UsageMetadata.CandidatesTokenCount,
		nil
}
