package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EvalPlan(planText string, evalCase *EvalCase, opts *EvalOptions) (*EvalReport, error) {
	scorers := []Scorer{
		&CoverageScorer{},
		&SpecificityScorer{},
		&ActionableScorer{},
	}

	if opts != nil && opts.Judge != "" {
		scorers = append(scorers, NewLLMJudgeScorer(opts.Judge))
	}

	scores := []EvalScore{}
	for _, scorer := range scorers {
		score, err := scorer.Score(planText, evalCase)
		if err != nil {
			continue
		}

		normalized := score / scorer.Max()
		pass := normalized >= 0.5

		scores = append(scores, EvalScore{
			Name:  scorer.Name(),
			Score: normalized,
			Max:   1.0,
			Pass:  pass,
		})
	}

	overallScore := computeOverallScore(scores)
	minScore := evalCase.MinScore
	if minScore == 0 {
		minScore = 6.0
	}
	pass := overallScore*10 >= minScore

	summary := fmt.Sprintf("Model: plan | Overall: %.1f/10 | %s", overallScore*10, passFailStr(pass))

	markdown := buildMarkdown("plan", evalCase.Task, scores, overallScore, pass)

	return &EvalReport{
		Task:         evalCase.Task,
		Model:        "plan",
		Scores:       scores,
		OverallScore: overallScore,
		Pass:         pass,
		Summary:      summary,
		Markdown:     markdown,
	}, nil
}

func EvalRun(runDir string, evalCase *EvalCase, opts *EvalOptions) (map[string]*EvalReport, error) {
	reports := make(map[string]*EvalReport)

	files, err := os.ReadDir(runDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "plan-") && strings.HasSuffix(file.Name(), ".md") {
			modelID := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "plan-"), ".md")
			filePath := filepath.Join(runDir, file.Name())

			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			report, err := EvalPlan(string(content), evalCase, opts)
			if err != nil {
				continue
			}

			report.Model = modelID
			reports[modelID] = report
		}
	}

	// Eval final-plan.md if it exists
	finalPath := filepath.Join(runDir, "final-plan.md")
	if content, err := os.ReadFile(finalPath); err == nil {
		report, err := EvalPlan(string(content), evalCase, opts)
		if err == nil {
			report.Model = "final"
			reports["final"] = report
		}
	}

	return reports, nil
}

func computeOverallScore(scores []EvalScore) float64 {
	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, s := range scores {
		sum += s.Score
	}
	return sum / float64(len(scores))
}

func buildMarkdown(model, task string, scores []EvalScore, overallScore float64, pass bool) string {
	var sb strings.Builder

	sb.WriteString("# Evaluation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Task:** %s\n", task))
	sb.WriteString(fmt.Sprintf("**Model:** %s\n", model))
	sb.WriteString(fmt.Sprintf("**Overall Score:** %.1f/10 %s\n\n", overallScore*10, passFailIcon(pass)))
	sb.WriteString("## Scores\n\n")

	for _, score := range scores {
		icon := passFailIcon(score.Pass)
		sb.WriteString(fmt.Sprintf("- **%s:** %.1f/10 %s\n", score.Name, score.Score*10, icon))
		if score.Details != "" {
			sb.WriteString(fmt.Sprintf("  - %s\n", score.Details))
		}
	}

	return sb.String()
}

func passFailStr(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}

func passFailIcon(pass bool) string {
	if pass {
		return "✓"
	}
	return "✗"
}
