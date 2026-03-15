import { writeFileSync, mkdirSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
export function generateSkillMd() {
    return `# multiplan — Multimodel Planning Workflow

Run a task through **Claude (Opus), Gemini, Codex (GPT-5.4), and GLM-5** in parallel, cross-examine the plans, and converge on a single definitive technical plan.

## When to use

- Designing a new feature or system with real architectural trade-offs
- Any decision where a single model's blind spots matter
- When you want a plan stress-tested before a single line of code is written
- When asked to "plan X with all models", "run multimodel planning on X", or "multiplan X"

## How it works

**Phase 1 — Independent planning (parallel)**
All four models receive the same task spec and produce independent plans with no cross-contamination.

**Phase 2 — Cross-examination (debate)**
One model reviews all four plans: what each gets right, what each misses, where they agree/disagree.

**Phase 3 — Convergence**
Final synthesis: best ideas from all four, disagreements resolved, gaps filled. One actionable plan.

---

## Invocation

\`\`\`bash
multiplan plan "task description" [--req "requirements"] [--con "constraints"] [--out /path]
# or with the legacy shell wrapper:
multiplan "task" [--req "..."] [--con "..."]
\`\`\`

### CLI Options

| Flag | Description |
|------|-------------|
| \`--req\` | Requirements |
| \`--con\` | Constraints |
| \`--out\` | Output dir (default: \`~/.multiplan/runs/<timestamp>\`) |
| \`--models\` | Subset: \`claude,gemini,codex,glm5\` |
| \`--verbose\` | Extra logging |

### Examples

\`\`\`bash
multiplan plan "Design a rate limiting system for the API"

multiplan plan "Build a real-time notification system" \\
  --req "Must support 10k concurrent users, WebSocket-based" \\
  --con "No new infrastructure — use existing Redis + Postgres"
\`\`\`

---

## Eval

\`\`\`bash
# Evaluate a plan file
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md

# Evaluate with a fixture (expected topics + min score)
multiplan eval ~/.multiplan/runs/LATEST --fixture eval/fixtures/rate-limiter.json --judge claude

# Output as JSON
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md --json
\`\`\`

---

## Output files (per run)

| File | Contents |
|------|----------|
| \`plan-claude.md\` | Claude Opus independent plan |
| \`plan-gemini.md\` | Gemini independent plan |
| \`plan-codex.md\` | Codex/GPT independent plan |
| \`plan-glm5.md\` | GLM-5 independent plan |
| \`debate.md\` | Cross-examination analysis |
| \`final-plan.md\` | ✅ The convergence — start here |

---

## Integration: OpenClaw

When a user asks to plan something with all models:

\`\`\`bash
multiplan plan "<task>" --req "<requirements>" --con "<constraints>"
\`\`\`

Then present the contents of \`final-plan.md\` as the response.

For longer tasks, spawn a subagent:
\`\`\`
sessions_spawn: task = "run multiplan plan '<task>' and return the final plan"
\`\`\`

---

## Integration: Claude Code

Add to \`CLAUDE.md\` in any project:

\`\`\`markdown
## Multimodel Planning
When designing a complex feature or architecture decision, run:
\`\`\`bash
multiplan plan "describe the design task" --req "requirements" --con "constraints"
\`\`\`
\`\`\`

---

## Prompt templates

Located at \`src/prompts/\` in the multiplan repo (\`~/github/multiplan\`):
- \`plan.md\` — planning prompt (vars: \`{{TASK}}\`, \`{{REQUIREMENTS}}\`, \`{{CONSTRAINTS}}\`)
- \`debate.md\` — cross-exam prompt (adds: \`{{PLAN_A}}\`, \`{{PLAN_B}}\`, \`{{PLAN_C}}\`, \`{{PLAN_D}}\`)
- \`converge.md\` — convergence prompt (all of the above + \`{{DEBATE}}\`)

---

## Auth / Keys

| Model | Auth |
|-------|------|
| Claude | \`claude\` CLI (authenticated via Claude Code) |
| Gemini | \`gemini\` CLI (authenticated via Google) |
| Codex | \`codex\` CLI (authenticated via OpenAI) |
| GLM-5 | Read from \`~/.openclaw/agents/main/agent/auth-profiles.json\` → \`profiles["zai:default"].key\`, or \`ZAI_API_KEY\` env var |
`;
}
export function writeSkill() {
    const skillDir = join(homedir(), 'openclaw/skills/multiplan');
    mkdirSync(skillDir, { recursive: true });
    const skillPath = join(skillDir, 'SKILL.md');
    writeFileSync(skillPath, generateSkillMd(), 'utf-8');
    console.log(`✓ SKILL.md written to ${skillPath}`);
}
//# sourceMappingURL=openclaw.js.map