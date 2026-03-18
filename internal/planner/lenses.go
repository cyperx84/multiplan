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
	// Build plan sections dynamically from however many plans actually exist
	planSections := []string{}
	bestUniqueLines := []string{}
	labels := planLabels(len(plans))

	i := 0
	for modelID, plan := range plans {
		label := labels[i]
		section := fmt.Sprintf("## Plan %s (%s)\n\n%s", label, modelID, plan)
		planSections = append(planSections, section)
		bestUniqueLines = append(bestUniqueLines, fmt.Sprintf("- Plan %s (%s): [best idea only it has]", label, modelID))
		i++
	}

	n := len(plans)
	header := fmt.Sprintf("You are a critical technical reviewer tasked with cross-examining %d independent architectural plan(s).", n)

	prompt := fmt.Sprintf(`# Cross-Examination Prompt

%s

---

## Original Task

%s

---

%s

---

## Your Job

Analyse all plans critically. For each plan, identify:
1. What it gets right
2. What it misses or gets wrong
3. What assumptions it makes that may not hold

Then identify:
- Where the plans **agree** (high-confidence areas)
- Where the plans **disagree** (contested decisions — these need the most scrutiny)
- The **single best idea** from each plan that the others missed

## IMPORTANT

Do NOT ask clarifying questions. Do NOT wait for more input. Execute the cross-examination now based on the plans provided above.

## Output Format

`+"```"+`
## Agreements
[What all plans converge on — these are safe bets]

## Disagreements
[Where they diverge — include what each says and why it matters]

## Best Unique Ideas
%s

## Critical Gaps
[Important things ALL plans missed]

## Recommendation Summary
[2-3 sentences on what the convergence plan should prioritise]
`+"```",
		header,
		task,
		strings.Join(planSections, "\n\n---\n\n"),
		strings.Join(bestUniqueLines, "\n"),
	)

	return prompt
}

func GetConvergePrompt(task string, plans map[string]string, debate string, scores map[string]float64) string {
	n := len(plans)
	labels := planLabels(n)

	// Build score summary
	scoreLines := []string{}
	i := 0
	for modelID, score := range scores {
		if i < len(labels) {
			scoreLines = append(scoreLines, fmt.Sprintf("- Plan %s (%s): %.1f/10", labels[i], modelID, score*10))
			i++
		}
	}

	// Build plan sections
	planSections := []string{}
	i = 0
	for modelID, plan := range plans {
		if i < len(labels) {
			section := fmt.Sprintf("## Plan %s (%s)\n\n%s", labels[i], modelID, plan)
			planSections = append(planSections, section)
			i++
		}
	}

	prompt := fmt.Sprintf(`# Convergence Prompt

You are a senior architect producing a final, unified technical plan.

You have %d independent plan(s), eval scores for each, and a cross-examination analysis. Your job: synthesise the best ideas into one definitive, actionable plan.

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

## IMPORTANT

Do NOT ask clarifying questions. Do NOT wait for more input. Produce the final converged plan NOW based on the plans and analysis above.

## Instructions

- Take the **best ideas from each plan** as identified in the debate
- **Weight higher-scoring plans more heavily** — they demonstrated better structure and specificity
- Resolve all disagreements — pick a side and justify it briefly
- Fill in any critical gaps identified in the cross-examination
- The output must be **immediately actionable** — someone should be able to hand this to a developer and start building

## Output Format

`+"```"+`
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
`+"```",
		n,
		task,
		strings.Join(scoreLines, "\n"),
		strings.Join(planSections, "\n\n---\n\n"),
		debate,
	)

	return prompt
}

// planLabels returns labels A, B, C, ... for the given count.
func planLabels(n int) []string {
	labels := make([]string, n)
	for i := 0; i < n; i++ {
		labels[i] = string(rune('A' + i))
	}
	return labels
}
