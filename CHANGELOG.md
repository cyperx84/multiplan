# Changelog

## v0.4.0 (2026-03-16)

### ✨ New Features

#### Shared HTTP client with retry logic
Extracted a shared `APIClient` in `internal/models/client.go` used by all 4 providers (Claude, Gemini, Codex/GPT, GLM-5). Eliminates ~80% HTTP code duplication across providers.

#### Exponential backoff retry
All provider HTTP calls now retry automatically on rate limits (429) and server errors (500, 502, 503, 504) with exponential backoff: 1s, 2s, 4s. Max 3 retries. Client errors (400, 401, 403, 404) are not retried.

Verbose mode prints retry progress:
```
[retry] Claude: rate limited, retrying in 2s (attempt 2/3)
```

#### Better error messages
API key missing and authentication errors now include actionable guidance:
```
Claude requires ANTHROPIC_API_KEY. Set it with: export ANTHROPIC_API_KEY=sk-...
Gemini requires GOOGLE_AI_API_KEY or GEMINI_API_KEY. Get one at: https://aistudio.google.com/apikey
```

#### Planner unit tests
Added `internal/planner/planner_test.go` with 6 test cases covering phase ordering, score weighting, output file creation, model failure handling, all-models-fail recovery, and debate failure recovery. Tests use mock providers — no real API calls required.

#### New eval fixtures
Added 4 new evaluation fixtures in `eval/fixtures/`: `auth-system.json`, `db-migration.json`, `realtime-chat.json`, `cicd-pipeline.json`.

#### Clean `--version` output
`multiplan --version` now outputs `multiplan v0.4.0` (previously used cobra's default verbose format).

---

## v0.3.0 (2026-03-16)

### ✨ New Features

#### Streaming progress output
Default (non-verbose) mode now prints a progress line as each model completes:
```
⏳ Claude (Opus)... done (4.2s)
⏳ Gemini... done (3.1s)
```

#### `--quiet` flag
New global flag to suppress all progress output. Only errors and the final result are printed.

#### `--json` flag for `plan` command
Output a structured JSON object including run_id, output_dir, model excerpts, durations, debate excerpt, and the final plan.

#### Config file support
Load settings from `.multiplan.yml` (current directory) or `$HOME/.config/multiplan/config.yml`. CLI flags always override config file values. Supported fields: `models`, `debate_model`, `converge_model`, `timeout_ms`, `output_dir`, `requirements`, `constraints`.

#### Token cost tracking
Each model call now extracts token counts from API responses. At the end of a run, a cost summary is printed:
```
📊 Token usage: 45,230 input / 12,450 output (~$0.85 estimated)
```
Pricing used: Claude $15/$75 per 1M, Gemini $1.25/$5 per 1M, GPT-4o $2.50/$10 per 1M, GLM-5 $1/$2 per 1M.

#### LLM judge: any model
The `--judge` flag on `eval` now accepts `claude`, `gemini`, `codex`, or `glm5`.

### 🔧 Technical Changes

- `ModelResult` extended with `InputTokens` and `OutputTokens` fields
- New `ProviderWithTokens` interface implemented by all 4 providers
- New `internal/config/loader.go` for YAML config file loading
- CI updated to Go (was accidentally using Node.js steps)

---

## v0.2.0 (2026-03-15)

### 🎉 Complete rewrite in Go

- **Single binary** — No more Node.js runtime or npm dependencies
- **Direct API calls** — HTTP requests to Claude, Gemini, Codex, GLM-5 (no CLI shelling)
- **Faster startup** — Go binary vs Node.js interpreter

### ✨ New Features

#### Lens-based prompts
Each model gets a **different planning angle** to maximize diversity:

| Model | Lens |
|-------|------|
| Claude | Correctness & edge cases |
| Gemini | Scale & operational simplicity |
| Codex | Implementation speed |
| GLM-5 | Failure analysis & critique |

#### Eval → Convergence
Plans are now **scored before final synthesis**:
- Structural scorers run on all plans (coverage, specificity, actionability)
- Scores are injected into the convergence prompt
- Higher-scoring plans are weighted more heavily in the final synthesis

#### Streaming progress
See which models finish first via goroutines + channels (parallel execution with real-time feedback).

### 🔧 Technical Changes

- **Language**: TypeScript → Go
- **HTTP client**: Direct `net/http` calls (no child processes)
- **CLI framework**: `commander` → `cobra`
- **Tests**: Node.js test runner → Go testing
- **Build**: `npm run build` → `go build`

### 📦 Installation

```bash
# Go install
go install github.com/cyperx84/multiplan@latest

# Build from source
git clone https://github.com/cyperx84/multiplan
cd multiplan
go build -o multiplan .
```

### ⚠️ Breaking Changes

- **API keys via environment variables** (not CLI tools):
  - `ANTHROPIC_API_KEY` for Claude
  - `GOOGLE_AI_API_KEY` or `GEMINI_API_KEY` for Gemini
  - `OPENAI_API_KEY` for Codex
  - `ZAI_API_KEY` for GLM-5
- No more `claude`, `gemini`, `codex` CLI dependencies
- Homebrew formula changed from npm to Go binary

### 🧪 Testing

```bash
go test ./...
```

All structural eval scorers, config helpers, and lens prompt generation are covered.

---

## v0.1.0 (2024-03-15)

Initial TypeScript release with 4-model parallel planning, cross-examination, and structural eval framework.
