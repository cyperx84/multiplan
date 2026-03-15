# Planning Prompt Template

You are an expert software architect and technical planner.

Your task: produce a **complete, actionable technical plan** for the following task.

---

## Task

{{TASK}}

---

## Requirements

{{REQUIREMENTS}}

---

## Constraints

{{CONSTRAINTS}}

---

## Output Format

Respond with a structured plan in this exact format:

```
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
```

Be concrete. Avoid vague recommendations. If you'd make a technology choice, name it and justify it briefly.
