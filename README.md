# multiplan

**4-model parallel planning workflow with eval framework.**

Run a task through Claude (Opus), Gemini, Codex (GPT-5.4), and GLM-5 simultaneously. Each produces an independent plan. Then cross-examine them. Then converge on the best synthesis.

Kill model selection anxiety. Get plans that have been stress-tested before a single line of code is written.

---

## Install

```bash
npm install -g multiplan
```

Or run from source:
```bash
git clone https://github.com/cyperx84/multiplan
cd multiplan && npm install && npm run build
npm link
```

---

## Quick start

```bash
# Run planning on any task
multiplan "Design a rate limiting system for the API"

# With requirements and constraints
multiplan "Build a real-time notification system" \
  --req "Must support 10k concurrent users, WebSocket-based" \
  --con "No new infrastructure — use existing Redis + Postgres"

# Evaluate the output
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md
```

---

## How it works

**Phase 1 — Independent planning (parallel)**
All four models receive the same task spec. Each produces a complete plan with no cross-contamination. They run concurrently — total time ≈ slowest model.

**Phase 2 — Cross-examination**
One model (default: Claude) reviews all four plans: what each gets right, what each misses, where they agree, where they disagree.

**Phase 3 — Convergence**
Final synthesis: best ideas from all four, disagreements resolved, gaps filled. One actionable, validated plan.

Output directory: `~/.multiplan/runs/<timestamp>/`

| File | Contents |
|------|----------|
| `plan-claude.md` | Claude Opus plan |
| `plan-gemini.md` | Gemini plan |
| `plan-codex.md` | Codex/GPT plan |
| `plan-glm5.md` | GLM-5 plan |
| `debate.md` | Cross-examination |
| `final-plan.md` | ✅ Start here |

---

## CLI reference

```
multiplan <task> [options]
multiplan plan <task> [options]
multiplan eval <file-or-dir> [options]
multiplan skill
multiplan integrations [--claude-code] [--codex]
```

### Options

| Flag | Description |
|------|-------------|
| `--req <text>` | Requirements |
| `--con <text>` | Constraints |
| `--out <dir>` | Output directory |
| `--models <list>` | Comma-separated: `claude,gemini,codex,glm5` |
| `--debate-model` | Model for cross-examination phase |
| `--converge-model` | Model for convergence phase |
| `--timeout <ms>` | Per-model timeout (default: 120000) |
| `--verbose` | Extra logging |

### Eval options

```bash
# Eval a single plan file
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md

# Eval all plans in a run directory
multiplan eval ~/.multiplan/runs/LATEST/

# Use a fixture (expected topics + min score threshold)
multiplan eval ~/.multiplan/runs/LATEST/ --fixture eval/fixtures/rate-limiter.json

# Choose LLM judge model
multiplan eval final-plan.md --judge gemini

# Skip LLM judge (fast, structural only)
multiplan eval final-plan.md --no-judge

# JSON output
multiplan eval final-plan.md --json
```

---

## Eval framework

The eval system scores plans on two axes:

### Structural scorers (fast, no model calls)

| Scorer | What it measures |
|--------|-----------------|
| **Coverage** | Required sections present (Overview, Architecture, Implementation, Trade-offs) |
| **Specificity** | Ratio of concrete terms (tech names, numbers, commands) vs vague language ("might", "could", "perhaps") |
| **Actionable** | Numbered steps, code blocks, concrete commands |

### LLM Judge (calls a model)

Grades the plan 0-10 on: completeness, concreteness, risk awareness, implementability. Returns `overall` as the judge score.

### Fixtures

Pre-built eval cases in `eval/fixtures/*.json`:

```json
{
  "task": "Design a rate limiting system for a REST API",
  "requirements": "Per-user and per-IP limits, sliding window",
  "constraints": "Redis only",
  "expectedTopics": ["Redis", "sliding window", "middleware"],
  "minScore": 6
}
```

### Use as a module

```typescript
import { run } from 'multiplan';
import { evalPlan, evalRun } from 'multiplan/eval';

// Run planning
const result = await run({
  task: 'Design a rate limiting system',
  requirements: 'Per-user limits',
  constraints: 'Redis only',
});

// Eval a single plan
const report = await evalPlan(result.finalPlan, {
  task: result.task,
  expectedTopics: ['Redis', 'sliding window'],
  minScore: 7,
}, { judge: 'claude' });

console.log(report.markdown);
console.log(report.pass); // true/false

// Eval all plans in a run directory
const reports = await evalRun(result.outputDir, {
  task: result.task,
  minScore: 6,
});
```

---

## Model auth

| Model | Auth |
|-------|------|
| **Claude** | `claude` CLI — authenticated via Claude Code / `claude auth login` |
| **Gemini** | `gemini` CLI — authenticated via `gemini auth login` |
| **Codex** | `codex` CLI — authenticated via `codex auth login` |
| **GLM-5** | Reads from `~/.openclaw/agents/main/agent/auth-profiles.json` (OpenClaw), or set `ZAI_API_KEY` env var |

Models are auto-discovered at runtime. If a model is missing or fails, it's marked unavailable and skipped — planning continues with the remaining models.

---

## Config

Create `~/.multiplanrc` or `.multiplanrc` in your project root:

```json
{
  "defaultModels": ["claude", "gemini", "glm5"],
  "timeoutMs": 180000,
  "debateModel": "claude",
  "convergeModel": "claude"
}
```

---

## Integration

### OpenClaw skill

```bash
multiplan skill
```

Generates/updates `~/openclaw/skills/multiplan/SKILL.md` so OpenClaw's AI can invoke multiplan directly.

### Claude Code (CLAUDE.md)

```bash
multiplan integrations --claude-code
```

Paste the output into your project's `CLAUDE.md`.

### Codex (.codex/agent.md)

```bash
multiplan integrations --codex
```

Paste into `.codex/agent.md` in your project.

---

## Development

```bash
npm run build       # TypeScript compile
npm run dev         # Watch mode
npm test            # Build + run unit tests (no model calls)
npm run test:unit   # Unit tests only (after build)
```

---

## License

MIT — [CyperX](https://github.com/cyperx84)
