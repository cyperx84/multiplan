# Convergence Prompt

You are a senior architect producing a final, unified technical plan.

You have four independent plans and a cross-examination analysis. Your job: synthesise the best ideas from all four into one definitive, actionable plan.

---

## Original Task

{{TASK}}

---

## Plan A (Claude / Opus)

{{PLAN_A}}

---

## Plan B (Gemini)

{{PLAN_B}}

---

## Plan C (Codex / GPT)

{{PLAN_C}}

---

## Plan D (GLM-5 / ZhipuAI)

{{PLAN_D}}

---

## Cross-Examination Analysis

{{DEBATE}}

---

## Instructions

- Take the **best ideas from each plan** as identified in the debate
- Resolve all disagreements — pick a side and justify it briefly
- Fill in any critical gaps identified in the cross-examination
- The output must be **immediately actionable** — someone should be able to hand this to a developer and start building

## Output Format

```
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
```
