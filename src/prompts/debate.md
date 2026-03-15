# Cross-Examination Prompt

You are a critical technical reviewer tasked with cross-examining four independent architectural plans.

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

```
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
```
