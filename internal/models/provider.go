package models

import (
	"context"
	"time"
)

// Provider is the interface for all LLM backends.
type Provider interface {
	ID() string
	Name() string
	Available(ctx context.Context) bool
	Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error)
}

// ProviderWithTokens extends Provider with token-count reporting.
type ProviderWithTokens interface {
	Provider
	PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (text string, inputTokens int, outputTokens int, err error)
}

// ModelResult holds the output from a single model run.
type ModelResult struct {
	ModelID      string
	ModelName    string
	Plan         string
	DurationMs   int64
	Error        string
	InputTokens  int
	OutputTokens int
}

// TokenCost holds per-model pricing in USD per 1M tokens.
type TokenCost struct {
	InputPer1M  float64
	OutputPer1M float64
}

// ModelPricing maps model ID → pricing.
var ModelPricing = map[string]TokenCost{
	"claude": {InputPer1M: 15.0, OutputPer1M: 75.0},
	"gemini": {InputPer1M: 1.25, OutputPer1M: 5.0},
	"codex":  {InputPer1M: 2.50, OutputPer1M: 10.0},
	"glm5":   {InputPer1M: 1.0, OutputPer1M: 2.0},
}

// EstimateCost returns the estimated cost in USD for the given token counts.
func EstimateCost(modelID string, inputTokens, outputTokens int) float64 {
	pricing, ok := ModelPricing[modelID]
	if !ok {
		return 0
	}
	return float64(inputTokens)/1_000_000*pricing.InputPer1M +
		float64(outputTokens)/1_000_000*pricing.OutputPer1M
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
