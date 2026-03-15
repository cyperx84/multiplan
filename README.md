# multiplan

**4-model parallel planning workflow with eval framework.**

Run a task through Claude (Opus), Gemini, Codex (GPT), and GLM-5 simultaneously. Each produces an independent plan with a **lens-based prompt** (correctness, scale, speed, or failure analysis). Then cross-examine them. Then converge on the best synthesis **weighted by eval scores**.

Kill model selection anxiety. Get plans that have been stress-tested before a single line of code is written.

---

## Install

### Go install

```bash
go install github.com/cyperx84/multiplan@latest
```

### Download binary

Grab the latest release from [Releases](https://github.com/cyperx84/multiplan/releases).

### Build from source

```bash
git clone https://github.com/cyperx84/multiplan
cd multiplan
go build -o multiplan .
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

## How it works (v0.2.0 — lens-based prompts + eval scores)

**Phase 1 — Independent planning (parallel, lens-based)**

All four models receive the same task spec, but with **different lenses**:

| Model | Lens |
|-------|------|
| **Claude** | "Plan this prioritising correctness and edge cases" |
| **Gemini** | "Plan this prioritising scale and operational simplicity" |
| **Codex/GPT** | "Plan this from a pure implementation perspective — what ships fastest" |
| **GLM-5** | "Plan this as a critic — what are all the ways this could fail" |

Each produces a complete plan with no cross-contamination. They run concurrently — total time ≈ slowest model.

**Phase 1.5 — Eval scores (NEW in v0.2.0)**

Each plan is scored on:
- **Coverage** — Required sections present
- **Specificity** — Concrete tech/numbers vs vague language
- **Actionable** — Numbered steps, code blocks

Scores (0-10) are injected into the convergence phase so the final plan can **weight higher-scoring plans more heavily**.

**Phase 2 — Cross-examination**

One model (default: Claude) reviews all four plans: what each gets right, what each misses, where they agree, where they disagree.

**Phase 3 — Convergence (weighted by eval scores)**

Final synthesis: best ideas from all four, **prioritising higher-scoring plans**, disagreements resolved, gaps filled. One actionable, validated plan.

Output directory: `~/.multiplan/runs/<timestamp>/`

| File | Contents |
|------|----------|
| `plan-claude.md` | Claude Opus plan (correctness lens) |
| `plan-gemini.md` | Gemini plan (scale lens) |
| `plan-codex.md` | Codex/GPT plan (speed lens) |
| `plan-glm5.md` | GLM-5 plan (critic lens) |
| `debate.md` | Cross-examination |
| `final-plan.md` | ✅ Start here |

---

## CLI reference

```
multiplan <task> [options]
multiplan plan <task> [options]
multiplan eval <file-or-dir> [options]
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

---

## Model auth

| Model | Auth |
|-------|------|
| **Claude** | `ANTHROPIC_API_KEY` environment variable |
| **Gemini** | `GOOGLE_AI_API_KEY` or `GEMINI_API_KEY` environment variable |
| **Codex** | `OPENAI_API_KEY` environment variable |
| **GLM-5** | `ZAI_API_KEY` env var, or reads from `~/.openclaw/agents/main/agent/auth-profiles.json` (OpenClaw) |

Models are auto-discovered at runtime. If a model is missing or fails, it's marked unavailable and skipped — planning continues with the remaining models.

---

## What's new in v0.2.0

1. **Lens-based prompts** — Each model gets a different planning angle (correctness, scale, speed, critique)
2. **Eval → convergence** — Plans are scored before convergence, and scores are injected into the final synthesis prompt
3. **Streaming progress** — See which models finish first (goroutines + channels)
4. **Single binary** — No runtime dependencies, just Go

---

## Development

```bash
go build -o multiplan .   # Build
go test ./...             # Run tests
./multiplan --help        # Verify CLI
```

---

## License

MIT — [CyperX](https://github.com/cyperx84)
