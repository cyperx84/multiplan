# Changelog

## v0.5.0 (2026-03-20)

### New Features

#### Claude CLI mode with plan mode JSON extraction
Claude provider now uses the `claude` CLI as its primary mode. Prompts are sent via `--print` with `--output-format json`, and plans are extracted from plan mode's `permission_denials[].tool_input.plan` field. Falls back to `ANTHROPIC_API_KEY` if the CLI is not available.

#### Codex CLI dual-mode provider
Codex provider supports both `codex exec` CLI mode and direct OpenAI API calls. CLI mode writes output to a temp file and reads it back. Falls back to `OPENAI_API_KEY`.

#### Dynamic debate/convergence prompts
Debate and convergence phases now work with any number of models (1–N), not just a fixed set of 4. Prompt templates dynamically adapt to the actual models that produced plans.

#### Lattice Phase 0 integration
Optional mental model framing via the `lattice` binary. When available, multiplan queries lattice for relevant mental models and injects them into Phase 1 prompts. Controlled via `--skip-lattice` and `--lattice-cmd` flags.

#### Edge case test coverage improvements
Added tests for Claude JSON parser edge cases (malformed JSON, empty responses, nested plan extraction) and provider path selection (CLI vs API mode).

---

## v0.4.0 (2026-03-16)

### New Features

#### Shared HTTP client with retry logic
Extracted a shared `APIClient` in `internal/models/client.go` used by all providers. Eliminates ~80% HTTP code duplication across providers.

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
```

#### Planner unit tests
Added `internal/planner/planner_test.go` with 6 test cases covering phase ordering, score weighting, output file creation, model failure handling, all-models-fail recovery, and debate failure recovery.

#### New eval fixtures
Added 4 new evaluation fixtures in `eval/fixtures/`: `auth-system.json`, `db-migration.json`, `realtime-chat.json`, `cicd-pipeline.json`.

#### Clean `--version` output
`multiplan --version` now outputs `multiplan v0.4.0` (previously used cobra's default verbose format).

---

## v0.3.0 (2026-03-16)

### New Features

#### Streaming progress output
Default (non-verbose) mode now prints a progress line as each model completes.

#### `--quiet` flag
New global flag to suppress all progress output. Only errors and the final result are printed.

#### `--json` flag for `plan` command
Output a structured JSON object including run_id, output_dir, model excerpts, durations, debate excerpt, and the final plan.

#### Config file support
Load settings from `.multiplan.yml` (current directory) or `$HOME/.config/multiplan/config.yml`. CLI flags always override config file values.

#### Token cost tracking
Each model call now extracts token counts from API responses. At the end of a run, a cost summary is printed.

#### LLM judge: any model
The `--judge` flag on `eval` now accepts any configured model.

### Technical Changes

- `ModelResult` extended with `InputTokens` and `OutputTokens` fields
- New `ProviderWithTokens` interface implemented by all providers
- New `internal/config/loader.go` for YAML config file loading
- CI updated to Go (was accidentally using Node.js steps)

---

## v0.2.0 (2026-03-15)

### Complete rewrite in Go

- **Single binary** — No more Node.js runtime or npm dependencies
- **Direct API calls** — HTTP requests to Claude, Codex, GLM-5 (no CLI shelling)
- **Faster startup** — Go binary vs Node.js interpreter

### New Features

#### Lens-based prompts
Each model gets a different planning angle to maximize diversity.

#### Eval-weighted convergence
Plans are scored before final synthesis. Higher-scoring plans are weighted more heavily.

#### Streaming progress
See which models finish first via goroutines + channels.

### Technical Changes

- **Language**: TypeScript -> Go
- **HTTP client**: Direct `net/http` calls (no child processes)
- **CLI framework**: `commander` -> `cobra`
- **Tests**: Node.js test runner -> Go testing
- **Build**: `npm run build` -> `go build`

### Breaking Changes

- API keys via environment variables (not CLI tools)
- Homebrew formula changed from npm to Go binary

---

## v0.1.0 (2024-03-15)

Initial TypeScript release with multi-model parallel planning, cross-examination, and structural eval framework.
