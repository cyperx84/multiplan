package eval

import (
	"math"
	"regexp"
	"strings"
)

// Coverage scorer
type CoverageScorer struct{}

func (c *CoverageScorer) Name() string { return "Coverage" }
func (c *CoverageScorer) Max() float64 { return 1.0 }

func (c *CoverageScorer) Score(text string, evalCase *EvalCase) (float64, error) {
	requiredSections := []string{"Overview", "Architecture", "Implementation", "Trade-off", "Risk"}
	found := 0

	for _, section := range requiredSections {
		pattern := regexp.MustCompile(`(?i)##\s+` + section)
		if pattern.MatchString(text) {
			found++
		}
	}

	return math.Min(float64(found)/float64(len(requiredSections)), 1.0), nil
}

// Specificity scorer
type SpecificityScorer struct{}

func (s *SpecificityScorer) Name() string { return "Specificity" }
func (s *SpecificityScorer) Max() float64 { return 1.0 }

func (s *SpecificityScorer) Score(text string, evalCase *EvalCase) (float64, error) {
	vagueTokens := []string{
		"might", "could", "consider", "potentially", "maybe",
		"perhaps", "should", "may", "possibly",
	}

	concretePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(postgres|redis|mongodb|mysql|elasticsearch|kafka|rabbitmq|nodejs|python|go|rust|java|typescript|javascript|docker|kubernetes|aws|gcp|azure)\b`),
		regexp.MustCompile(`\d+(\.\d+)?[a-z]*\b`),
		regexp.MustCompile("`[^`]+`"),
		regexp.MustCompile(`https?://\S+`),
	}

	words := strings.Fields(strings.ToLower(text))
	vagueCount := 0

	for _, word := range words {
		for _, vague := range vagueTokens {
			if strings.Contains(word, vague) {
				vagueCount++
				break
			}
		}
	}

	concreteCount := 0
	for _, pattern := range concretePatterns {
		matches := pattern.FindAllString(text, -1)
		concreteCount += len(matches)
	}

	total := len(words)
	if total == 0 {
		return 0, nil
	}

	vagueRatio := float64(vagueCount) / float64(total)
	concreteRatio := float64(concreteCount) / math.Max(float64(total), 50)

	score := math.Max(0, concreteRatio-vagueRatio)
	return math.Min(score, 1.0), nil
}

// Actionable scorer
type ActionableScorer struct{}

func (a *ActionableScorer) Name() string { return "Actionable" }
func (a *ActionableScorer) Max() float64 { return 1.0 }

func (a *ActionableScorer) Score(text string, evalCase *EvalCase) (float64, error) {
	hasNumberedList := regexp.MustCompile(`(?m)^\d+\.`).MatchString(text)
	hasCodeFences := strings.Contains(text, "```")
	hasBulletPoints := regexp.MustCompile(`(?m)^[-*]`).MatchString(text)

	codeBlockCount := len(regexp.MustCompile("```").FindAllString(text, -1)) / 2
	numberedItems := len(regexp.MustCompile(`(?m)^\d+\.`).FindAllString(text, -1))

	score := 0.0

	if hasNumberedList {
		score += 0.3
	}
	if hasCodeFences {
		score += 0.3
	}
	if hasBulletPoints {
		score += 0.2
	}

	// Bonus for density
	if codeBlockCount >= 2 {
		score += 0.1
	}
	if numberedItems >= 5 {
		score += 0.1
	}

	return math.Min(score, 1.0), nil
}
