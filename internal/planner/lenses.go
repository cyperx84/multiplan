package planner

import (
	"fmt"
	"strings"
)

// GetLensPrompt returns a lens-based planning prompt for the given model
func GetLensPrompt(modelID, task, requirements, constraints string) string {
	lens := getLensForModel(modelID)
	
	template := `# Planning Prompt Template

You are an expert software architect and technical planner.

%s

---

## Task

%s

---

## Requirements

%s

---

## Constraints

%s

---

## Output Format

Respond with a structured plan in this exact format:

` + "```" + `
## Overview
[2-3 sentence summary of your approach]

## Architecture
[Key architectural decisions and why]

## Components
[List each major component with its responsibility]

## Implementation Steps
[Ordered list of concrete implementation steps]

## Trade-offs & Risks
[What you're giving up, what could go wrong]

## Open Questions
[Things that need clarification or more context]
` + "```" + `

Be concrete. Avoid vague recommendations. If you'd make a technology choice, name it and justify it briefly.
`

	return fmt.Sprintf(template, lens, task, requirements, constraints)
}

func getLensForModel(modelID string) string {
	lenses := map[string]string{
		"claude": "Your lens: **Plan this prioritising correctness and edge cases**. Focus on what could go wrong, defensive programming, data validation, and robust error handling. Think like a principal engineer reviewing a critical production system.",
		"gemini": "Your lens: **Plan this prioritising scale and operational simplicity**. Focus on horizontal scaling, observability, deployment simplicity, and operational overhead. Think like a platform engineer building infrastructure that needs to handle 10x growth.",
		"codex":  "Your lens: **Plan this from a pure implementation perspective — what ships fastest**. Focus on the shortest path to a working prototype, minimal viable architecture, and iterative improvement. Think like a startup engineer shipping v1.",
		"glm5":   "Your lens: **Plan this as a critic — what are all the ways this could fail**. Focus on failure modes, security vulnerabilities, performance bottlenecks, and hidden complexity. Think like a senior architect doing a design review.",
	}

	lens, ok := lenses[modelID]
	if !ok {
		return "Your task: produce a **complete, actionable technical plan** for the following task."
	}
	return lens
}

func GetDebatePrompt(task string, plans map[string]string) string {
	template := `# Cross-Examination Prompt

You are a critical technical reviewer tasked with cross-examining four independent architectural plans.

---

## Original Task

%s

---

%s

---

## Your Job

Analyse all four plans critically. For each, identify:
1. What it gets right
2. What it misses or gets wrong
3. What assumptions it makes that may not hold

Then identify:
- Where the plans **agree** (high-confidence areas)
- Where the plans **disagree** (contested decisions — these need the most scrutiny)
- The **single best idea** from each plan that the others missed

## Output Format

` + "```" + `
## Agreements
[What all four plans converge on — these are safe bets]

## Disagreements
[Where they diverge — include what each says and why it matters]

## Best Unique Ideas
- Plan A: [best idea only it has]
- Plan B: [best idea only it has]
- Plan C: [best idea only it has]
- Plan D: [best idea only it has]

## Critical Gaps
[Important things ALL FOUR plans missed]

## Recommendation Summary
[2-3 sentences on what the convergence plan should prioritise]
` + "```" + `
`

	// Build plan sections
	planSections := []string{}
	labels := []string{"A", "B", "C", "D"}
	modelNames := []string{"Claude / Opus", "Gemini", "Codex / GPT", "GLM-5 / ZhipuAI"}
	
	i := 0
	for _, plan := range plans {
		if i < len(labels) {
			section := fmt.Sprintf("## Plan %s (%s)\n\n%s", labels[i], modelNames[i], plan)
			planSections = append(planSections, section)
			i++
		}
	}

	return fmt.Sprintf(template, task, strings.Join(planSections, "\n\n---\n\n"))
}

func GetConvergePrompt(task string, plans map[string]string, debate string, scores map[string]float64) string {
	template := `# Convergence Prompt

You are a senior architect producing a final, unified technical plan.

You have four independent plans, eval scores for each, and a cross-examination analysis. Your job: synthesise the best ideas from all four into one definitive, actionable plan.

---

## Original Task

%s

---

## Eval Scores (0-10 scale, higher = better)

%s

---

%s

---

## Cross-Examination Analysis

%s

---

## Instructions

- Take the **best ideas from each plan** as identified in the debate
- **Weight higher-scoring plans more heavily** — they demonstrated better structure and specificity
- Resolve all disagreements — pick a side and justify it briefly
- Fill in any critical gaps identified in the cross-examination
- The output must be **immediately actionable** — someone should be able to hand this to a developer and start building

## Output Format

` + "```" + `
## Final Architecture Decision
[The definitive architectural approach — clear, no hedging]

## Component Breakdown
[Each component, its responsibility, technology choice]

## Implementation Plan
[Phase-by-phase, ordered steps. Be specific.]

## Why These Choices
[Brief justification for the key decisions, especially where plans disagreed]

## What We're NOT Doing
[Explicitly list approaches considered and rejected, and why]

## First 3 Actions
[The very first three concrete things to do — unambiguous]
` + "```" + `
`

	// Build score summary
	scoreLines := []string{}
	labels := []string{"A", "B", "C", "D"}
	modelNames := []string{"Claude", "Gemini", "Codex", "GLM-5"}
	
	i := 0
	for modelID, score := range scores {
		if i < len(labels) {
			scoreLines = append(scoreLines, fmt.Sprintf("- Plan %s (%s / %s): %.1f/10", labels[i], modelNames[i], modelID, score*10))
			i++
		}
	}

	// Build plan sections
	planSections := []string{}
	i = 0
	for _, plan := range plans {
		if i < len(labels) {
			section := fmt.Sprintf("## Plan %s (%s)\n\n%s", labels[i], modelNames[i], plan)
			planSections = append(planSections, section)
			i++
		}
	}

	return fmt.Sprintf(template, task, strings.Join(scoreLines, "\n"), strings.Join(planSections, "\n\n---\n\n"), debate)
}
