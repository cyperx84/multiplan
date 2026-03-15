# multiplan

A 4-model parallel planning workflow with evaluation framework. Get superior technical plans by running Claude, Gemini, Codex, and GLM-5 in parallel, then cross-examining and synthesizing the results.

## What it is

**multiplan** orchestrates three phases:
1. **Parallel Planning** — All 4 models generate independent technical plans
2. **Cross-Examination** — A model analyzes agreements, disagreements, and unique insights
3. **Convergence** — A model synthesizes the best ideas into one unified final plan

Each plan is immediately actionable — someone can hand it to a developer and start building.

## Installation

```bash
npm install -g multiplan
```

Or install from source:
```bash
git clone https://github.com/cyperx84/multiplan.git
cd multiplan
npm install
npm run build
npm install -g .
```

## Quick start

```bash
# Generate a plan for a rate limiting system
multiplan plan "Design a rate limiting system for a REST API" \
  --req "Support per-user and per-IP limits" \
  --con "Use Redis for state"

# Evaluate the final plan against a fixture
multiplan eval ~/.multiplan/runs/LATEST/final-plan.md \
  --fixture eval/fixtures/rate-limiter.json
```

## How it works

### Phase 1: Parallel Planning
All available models receive the same planning prompt and generate independent technical plans. Plans are written to disk immediately.

**Models:**
- **Claude (Opus)** — Best architectural thinking, clear reasoning
- **Gemini** — Fast, practical implementations, good cost/quality
- **Codex (GPT)** — Production experience, familiar patterns
- **GLM-5 (ZhipuAI)** — Unique perspective, sometimes brilliant at trade-offs

### Phase 2: Cross-Examination (Debate)
One model (default: Claude) analyzes all 4 plans side-by-side:
- What each plan gets right
- What it misses or gets wrong
- Where all plans agree (safe bets)
- Where they disagree (contested territory)
- The single best unique idea from each plan
- Critical gaps all 4 missed

### Phase 3: Convergence (Synthesis)
Another model (default: Claude) produces the final unified plan:
- Takes the best ideas from each plan
- Resolves disagreements with clear justification
- Fills gaps identified in debate
- Output is immediately buildable

## Evaluation Framework

**multiplan** includes a scoring framework for evaluating plan quality:

```bash
# Score a single plan
multiplan eval my-plan.md --fixture eval/fixtures/rate-limiter.json

# Score all plans in a run
multiplan eval ~/.multiplan/runs/LATEST --fixture eval/fixtures/rate-limiter.json

# Include LLM judge (calls model to grade 0-10)
multiplan eval my-plan.md --fixture eval/fixtures/rate-limiter.json --judge claude

# Output as JSON
multiplan eval my-plan.md --fixture eval/fixtures/rate-limiter.json --json
```

**Scorers:**
- **Coverage** (0-1) — Presence of required sections: Overview, Architecture, Implementation, Trade-offs
- **Specificity** (0-1) — Ratio of concrete (tech names, numbers) to vague tokens ("might", "could")
- **Actionable** (0-1) — Presence of numbered steps, code blocks, commands
- **LLM Judge** (0-10) — Model grades on completeness, concreteness, risk awareness, implementability

**Eval Fixtures:**
```json
{
  "task": "Design a rate limiting system",
  "requirements": "Per-user and per-IP limits",
  "constraints": "Use Redis for state",
  "expectedTopics": ["Redis", "sliding window", "middleware"],
  "minScore": 6
}
```

## Configuration

### Models

Each model requires specific setup:

| Model | Setup | Key/Auth |
|-------|-------|----------|
| Claude | CLI: `claude` binary | Requires Claude Code or shell access |
| Gemini | CLI: `gemini` binary | Requires Gemini CLI |
| Codex | CLI: `codex` binary | Requires Codex shell |
| GLM-5 | HTTP API | `~/.openclaw/agents/main/agent/auth-profiles.json` or `$ZAI_API_KEY` env var |

### Environment

```bash
# GLM-5 API key (optional if in auth-profiles.json)
export ZAI_API_KEY="your-key"

# Disable specific models
multiplan plan "..." --models claude,gemini

# Set per-model timeout (default 120s)
multiplan plan "..." --timeout 180000

# Use different models for debate/convergence
multiplan plan "..." --debate-model gemini --converge-model claude

# Output to custom directory
multiplan plan "..." --out /tmp/my-plan

# Verbose logging
multiplan plan "..." --verbose
```

### .multiplanrc (future)

```json
{
  "models": ["claude", "gemini", "codex", "glm5"],
  "debateModel": "claude",
  "convergeModel": "claude",
  "outputDir": "~/.multiplan/runs",
  "timeoutMs": 120000
}
```

## CLI Commands

```bash
# Plan
multiplan plan <task> [--req <requirements>] [--con <constraints>] [--out <dir>] [--models <list>] [--verbose]

# Evaluate
multiplan eval <file-or-dir> [--fixture <path>] [--judge <model>] [--json]

# Integrations
multiplan skill         # Generate OpenClaw SKILL.md
multiplan integrations  # Print integration snippets
```

## Integration

### OpenClaw

```bash
multiplan skill > ~/openclaw/skills/multiplan/SKILL.md
```

### Claude Code

Add to `CLAUDE.md`:
```
## Multiplan Integration
- Skill: `/multiplan "design rate limiter" --req "Redis" --con "..."`
- Eval: `/multiplan-eval path/to/plan.md`
```

### Codex Agent

Add to `.codex/agent.md`:
```
Multiplan is available. Use for:
- Technical architecture planning
- Evaluating plan quality
- Cross-model synthesis
```

## Development

```bash
# Build
npm run build

# Watch
npm run dev

# Test (structural scorers)
npm test
```

## License

MIT — Copyright © CyperX
