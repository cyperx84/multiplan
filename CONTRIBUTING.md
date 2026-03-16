# Contributing to multiplan

## Development Setup

```bash
git clone https://github.com/cyperx84/multiplan.git
cd multiplan
go build -o multiplan .
```

Requires Go 1.22+.

## Building

```bash
# Build binary
go build -o multiplan .

# Install to $GOPATH/bin
go install .
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific package
go test ./internal/config/...
```

## Linting

```bash
go vet ./...
```

## Project Structure

```
main.go                    → Entrypoint
cmd/
  root.go                  → Root Cobra command + persistent flags
  plan.go                  → plan subcommand (--json, --quiet, etc.)
  eval.go                  → eval subcommand
internal/
  config/
    config.go              → Config struct + helper methods
    loader.go              → Load .multiplan.yml config file
  planner/
    planner.go             → Core 3-phase orchestration (parallel → debate → converge)
    lenses.go              → Lens-based prompts for each model
  models/
    provider.go            → Provider + ProviderWithTokens interfaces, ModelResult, cost helpers
    anthropic.go           → Claude (Opus) adapter
    google.go              → Gemini adapter
    openai.go              → Codex (GPT-4o) adapter
    glm.go                 → GLM-5 (ZhipuAI) adapter
  eval/
    types.go               → EvalCase, EvalReport, Scorer interface
    structural.go          → Structural scorers (length, headers, etc.)
    judge.go               → LLM judge scorer (any provider)
    eval.go                → EvalPlan runner + report generation
eval/
  fixtures/                → Sample eval fixture JSON files
homebrew/
  multiplan.rb             → Homebrew formula
```

## Adding a New Model

1. Create `internal/models/myprovider.go` and implement the `Provider` interface:

```go
package models

import (
    "context"
    "time"
)

type MyProvider struct{}

func (m *MyProvider) ID() string   { return "myprovider" }
func (m *MyProvider) Name() string { return "My Provider" }

func (m *MyProvider) Available(ctx context.Context) bool {
    return os.Getenv("MY_API_KEY") != ""
}

func (m *MyProvider) Plan(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
    text, _, _, err := m.PlanWithTokens(ctx, prompt, timeout)
    return text, err
}

// PlanWithTokens implements ProviderWithTokens for cost tracking.
func (m *MyProvider) PlanWithTokens(ctx context.Context, prompt string, timeout time.Duration) (string, int, int, error) {
    // ... make HTTP call, return (text, inputTokens, outputTokens, err)
}
```

2. Register it in `provider.go`:

```go
var providers = map[string]Provider{
    // ... existing providers ...
    "myprovider": &MyProvider{},
}
```

3. Add pricing in `provider.go`:

```go
var ModelPricing = map[string]TokenCost{
    // ... existing ...
    "myprovider": {InputPer1M: 1.0, OutputPer1M: 3.0},
}
```

4. Add a lens-based prompt for it in `internal/planner/lenses.go`.

5. Write tests in `internal/models/myprovider_test.go`.

## Adding a New Scorer

Scorers live in `internal/eval/structural.go`. Each scorer implements the `Scorer` interface:

```go
type Scorer interface {
    Name() string
    Max() float64
    Score(text string, evalCase *EvalCase) (float64, error)
}
```

Example:

```go
type MyScorer struct{}

func (s *MyScorer) Name() string { return "My Scorer" }
func (s *MyScorer) Max() float64 { return 10.0 }

func (s *MyScorer) Score(text string, evalCase *EvalCase) (float64, error) {
    // Analyze text and return score 0..10
    return 7.5, nil
}
```

Register it in `eval.go` inside `EvalPlan()` by appending to the `scorers` slice.

## Config File

multiplan supports `.multiplan.yml` in the current directory or `$HOME/.config/multiplan/config.yml`:

```yaml
models: [claude, gemini, codex, glm5]
debate_model: claude
converge_model: claude
timeout_ms: 120000
output_dir: ~/.multiplan/runs
requirements: ""
constraints: ""
```

CLI flags always override config file values.

## Release Process

1. Update `Version` in `cmd/root.go`:
   ```go
   Version: "0.4.0",
   ```

2. Update `CHANGELOG.md` with a new version section.

3. Update `homebrew/multiplan.rb` with the new version and SHA256 (available after tagging).

4. Commit, tag, and push:
   ```bash
   git add -A
   git commit -m "Release v0.4.0"
   git tag v0.4.0
   git push origin main --tags
   ```

5. The GitHub Actions CI will verify the build. After the tag is pushed, update the Homebrew formula SHA256 with the tarball hash from the release.
