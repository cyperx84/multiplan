![Go Version](https://img.shields.io/badge/go-1.22+-blue)
[![CI](https://github.com/cyperx84/multiplan/actions/workflows/ci.yml/badge.svg)](https://github.com/cyperx84/multiplan/actions/workflows/ci.yml)
![License](https://img.shields.io/badge/license-MIT-green)

# multiplan

**Multi-model parallel planning workflow with eval framework.**

Run a task through Claude (Opus), Codex (GPT), and GLM-5 simultaneously. Each produces an independent plan with a **lens-based prompt** (correctness, speed, or failure analysis). Then cross-examine them. Then converge on the best synthesis **weighted by eval scores**.

Kill model selection anxiety. Get plans that have been stress-tested before a single line of code is written.

---

## Install

### Go install (recommended)

```bash
go install github.com/cyperx84/multiplan@latest
```

### GitHub releases

Grab the latest binary from [Releases](https://github.com/cyperx84/multiplan/releases).

### Homebrew

```bash
brew install cyperx84/tap/multiplan
```

### Build from source

```bash
git clone https://github.com/cyperx84/multiplan
cd multiplan
make build
# or: go build -o multiplan .
```

---

## Quick start

```bash
# Run planning on any task (bare args or plan subcommand both work)
multiplan "Design a rate limiting system for the API"
multiplan plan "Design a rate limiting system for the API"

# With requirements and constraints
multiplan "Build a real-time notification system" \
  --req "Must support 10k concurrent users, WebSocket-based" \
  --con "No new infrastructure — use existing Redis + Postgres"

# Evaluate the output
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md
```

---

## How it works

**Phase 0 — Lattice mental model framing (optional)**

If a `lattice` binary is on your PATH, multiplan queries it for relevant mental models and injects them into Phase 1 prompts. Skip with `--skip-lattice`.

**Phase 1 — Independent planning (parallel, lens-based)**

All models receive the same task spec, but with **different lenses**:

| Model | Lens |
|-------|------|
| **Claude** | "Plan this prioritising correctness and edge cases" |
| **Codex/GPT** | "Plan this from a pure implementation perspective — what ships fastest" |
| **GLM-5** | "Plan this as a critic — what are all the ways this could fail" |

Each produces a complete plan with no cross-contamination. They run concurrently — total time ≈ slowest model.

**Phase 1.5 — Eval scores**

Each plan is scored on Coverage, Specificity, and Actionability. Scores are injected into the convergence phase.

**Phase 2 — Cross-examination**

One model (default: Claude) reviews all plans: what each gets right, what each misses, where they agree, where they disagree.

**Phase 3 — Convergence (weighted by eval scores)**

Final synthesis: best ideas from all plans, prioritising higher-scoring plans. One actionable, validated plan.

Output directory: `~/.multiplan/runs/<timestamp>/`

| File | Contents |
|------|----------|
| `plan-claude.md` | Claude Opus plan (correctness lens) |
| `plan-codex.md` | Codex/GPT plan (speed lens) |
| `plan-glm5.md` | GLM-5 plan (critic lens) |
| `debate.md` | Cross-examination |
| `final-plan.md` | Start here |

---

## CLI reference

```
multiplan <task> [options]
multiplan plan <task> [options]
multiplan eval <file-or-dir> [options]
```

### Global flags

| Flag | Description |
|------|-------------|
| `--req <text>` | Requirements |
| `--con <text>` | Constraints |
| `--out <dir>` | Output directory |
| `--models <list>` | Comma-separated: `claude,codex,glm5` |
| `--debate-model` | Model for cross-examination phase (default: claude) |
| `--converge-model` | Model for convergence phase (default: claude) |
| `--timeout <ms>` | Per-model timeout (default: 300000 / 5m) |
| `--verbose` | Extra logging |
| `--quiet` | Suppress all progress output (errors + final result only) |
| `--skip-lattice` | Skip lattice mental model framing (Phase 0) |
| `--lattice-cmd` | Lattice binary command (default: `lattice`) |
| `--claude-model` | Model override for Claude CLI mode (e.g. `claude-opus-4-20250514`) |
| `--claude-cmd` | Claude CLI binary path (default: `claude`) |

### Plan flags

| Flag | Description |
|------|-------------|
| `--json` | Output structured JSON result |

### Eval flags

```bash
# Eval a single plan file
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md

# Use a fixture
multiplan eval ~/.multiplan/runs/LATEST/ --fixture eval/fixtures/rate-limiter.json

# Choose LLM judge model (claude, codex, glm5)
multiplan eval final-plan.md --judge claude

# Skip LLM judge
multiplan eval final-plan.md --no-judge

# JSON output
multiplan eval final-plan.md --json
```

---

## Config file

multiplan loads from `.multiplan.yml` in the current directory, or `$HOME/.config/multiplan/config.yml`.

```yaml
models: [claude, codex, glm5]
debate_model: claude
converge_model: claude
timeout_ms: 300000
output_dir: ~/.multiplan/runs
requirements: ""
constraints: ""
```

CLI flags always override config file values. The config file is optional — defaults work without it.

---

## JSON output

Add `--json` to the `plan` command for machine-readable output:

```bash
multiplan plan "Design a caching layer" --json
```

Output:

```json
{
  "run_id": "20260316T120000-123456",
  "output_dir": "/home/user/.multiplan/runs/...",
  "models": [
    {
      "model_id": "claude",
      "model_name": "Claude (Opus)",
      "plan_excerpt": "...",
      "duration_ms": 4200
    }
  ],
  "debate_excerpt": "...",
  "final_plan": "# Multimodel Plan: ..."
}
```

---

## Quiet mode

Suppress all progress output — useful in scripts:

```bash
multiplan plan "Design a cache layer" --quiet --json > result.json
```

---

## Token & cost tracking

After each run, multiplan prints a cost estimate:

```
Token usage: 45,230 input / 12,450 output (~$0.85 estimated)
```

Pricing used (rough estimates):

| Model | Input / 1M | Output / 1M |
|-------|-----------|------------|
| Claude Opus | $15 | $75 |
| GPT-4o | $2.50 | $10 |
| GLM-5 | $1 | $2 |

---

## Eval framework

### Structural scorers (fast, no model calls)

| Scorer | What it measures |
|--------|-----------------|
| **Coverage** | Required sections present |
| **Specificity** | Concrete tech/numbers vs vague language |
| **Actionable** | Numbered steps, code blocks, commands |

### LLM Judge

```bash
multiplan eval final-plan.md --judge claude   # or codex, glm5
```

### Fixtures

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

Claude and Codex support **CLI mode** as the primary auth method — no API keys needed if you have the CLI tools installed.

| Model | Primary (CLI mode) | Fallback (API key) |
|-------|-------------------|-------------------|
| **Claude** | `claude` CLI (uses plan mode with JSON extraction) | `ANTHROPIC_API_KEY` |
| **Codex** | `codex` CLI (dual-mode: exec + interactive) | `OPENAI_API_KEY` |
| **GLM-5** | — | `ZAI_API_KEY` or `GLM_API_KEY` or `~/.openclaw/` auth profiles |

Models are auto-discovered at runtime. Missing/failed models are skipped — multiplan works with any subset of models.

---

## Development

```bash
make build     # Build binary
make test      # Run tests with race detection
make lint      # go vet
make install   # go install
make clean     # Remove build artifacts
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## License

MIT — [CyperX](https://github.com/cyperx84)
