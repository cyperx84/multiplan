package models

import (
	"context"
	"time"
)

type Provider interface {
	ID() string
	Name() string
	Available(ctx context.Context) bool
	Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error)
}

type ModelResult struct {
	ModelID    string
	ModelName  string
	Plan       string
	DurationMs int64
	Error      string
}

var providers = map[string]Provider{
	"claude": &ClaudeProvider{},
	"gemini": &GeminiProvider{},
	"codex":  &CodexProvider{},
	"glm5":   &GLMProvider{},
}

func GetAvailableModels(ctx context.Context) []string {
	var available []string
	for id, provider := range providers {
		if provider.Available(ctx) {
			available = append(available, id)
		}
	}
	return available
}

func GetProvider(id string) (Provider, bool) {
	p, ok := providers[id]
	return p, ok
}

func GetAllProviders() map[string]Provider {
	return providers
}
