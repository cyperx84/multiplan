export function generateClaudeCodeSnippet() {
    return `## multiplan — Multimodel Planning Tool

When designing a complex feature, system, or architecture decision with real trade-offs, run multiplan to get independent plans from 4 AI models (Claude Opus, Gemini, Codex, GLM-5), cross-examine them, and converge on the best synthesis.

### Run a planning session

\`\`\`bash
multiplan plan "describe the design task" \\
  --req "requirements" \\
  --con "constraints"
\`\`\`

Output lands in \`~/.multiplan/runs/<timestamp>/\`:
- \`plan-claude.md\`, \`plan-gemini.md\`, \`plan-codex.md\`, \`plan-glm5.md\` — independent plans
- \`debate.md\` — cross-examination analysis
- \`final-plan.md\` — **convergence (start here)**

### Evaluate a plan

\`\`\`bash
# Score a plan file (structural + LLM judge)
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md

# Eval with fixture (expected topics + min score threshold)
multiplan eval ~/.multiplan/runs/LATEST --fixture /path/to/fixture.json --judge claude
\`\`\`

### Use as a module

\`\`\`typescript
import { run } from 'multiplan';
import { evalPlan } from 'multiplan/eval';

const result = await run({
  task: 'Design a rate limiting system',
  requirements: 'Per-user and per-IP limits',
  constraints: 'Redis only',
});

const report = await evalPlan(result.finalPlan, {
  task: result.config.task,
  expectedTopics: ['Redis', 'sliding window'],
  minScore: 7,
});

console.log(report.markdown);
\`\`\`

### When to use multiplan

- New feature with architectural decisions (which DB? which pattern?)
- System design with competing approaches
- Any time you want more than one perspective before committing to a plan
- Pre-implementation validation (run eval before coding)
`;
}
export function generateCodexAgentSnippet() {
    return `## multiplan — Multimodel Planning

Use multiplan when you need to design a system or feature with real architectural trade-offs.
It runs Claude, Gemini, Codex, and GLM-5 in parallel, debates the plans, and converges.

\`\`\`bash
multiplan plan "task" --req "requirements" --con "constraints"
# Final plan: ~/.multiplan/runs/<timestamp>/final-plan.md

# Eval:
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md
\`\`\`

Always read \`final-plan.md\` before starting implementation on complex tasks.
`;
}
//# sourceMappingURL=claude-code.js.map