package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cyperx84/multiplan/internal/models"
)

type LLMJudgeScorer struct {
	judgeModel string
}

func NewLLMJudgeScorer(judgeModel string) *LLMJudgeScorer {
	if judgeModel == "" {
		judgeModel = "claude"
	}
	return &LLMJudgeScorer{judgeModel: judgeModel}
}

func (l *LLMJudgeScorer) Name() string { return "LLM Judge" }
func (l *LLMJudgeScorer) Max() float64 { return 10.0 }

func (l *LLMJudgeScorer) Score(text string, evalCase *EvalCase) (float64, error) {
	provider, ok := models.GetProvider(l.judgeModel)
	if !ok {
		return 0, fmt.Errorf("judge model not available: %s", l.judgeModel)
	}

	prompt := fmt.Sprintf(`You are evaluating a technical plan. Score it on each dimension from 0-10. Return ONLY valid JSON, no markdown, no code fences.

Plan:
%s

Return: {"completeness": N, "concreteness": N, "risk_awareness": N, "implementability": N, "overall": N, "reasoning": "one sentence"}`, text)

	ctx := context.Background()
	timeout := 60 * time.Second

	response, err := provider.Plan(ctx, prompt, timeout)
	if err != nil {
		return 0, err
	}

	var result struct {
		Completeness     float64 `json:"completeness"`
		Concreteness     float64 `json:"concreteness"`
		RiskAwareness    float64 `json:"risk_awareness"`
		Implementability float64 `json:"implementability"`
		Overall          float64 `json:"overall"`
		Reasoning        string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return 0, err
	}

	return result.Overall, nil
}
