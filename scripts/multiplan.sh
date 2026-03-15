#!/usr/bin/env bash
# multiplan — 3-model parallel planning workflow
# Usage: multiplan "task description" [--req "requirements"] [--con "constraints"] [--out /path/to/output]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
PROMPTS_DIR="$SCRIPT_DIR/../prompts"

# ── Defaults ──────────────────────────────────────────────────────────────────
TASK=""
REQUIREMENTS="None specified."
CONSTRAINTS="None specified."
OUT_DIR=""
VERBOSE=0

# ── Arg parsing ───────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --req|-r) REQUIREMENTS="$2"; shift 2 ;;
    --con|-c) CONSTRAINTS="$2"; shift 2 ;;
    --out|-o) OUT_DIR="$2"; shift 2 ;;
    --verbose|-v) VERBOSE=1; shift ;;
    --help|-h)
      echo "Usage: multiplan <task> [--req <requirements>] [--con <constraints>] [--out <dir>]"
      exit 0 ;;
    *)
      if [[ -z "$TASK" ]]; then TASK="$1"; shift
      else echo "Unknown arg: $1" >&2; exit 1; fi ;;
  esac
done

if [[ -z "$TASK" ]]; then
  echo "Error: task description required." >&2
  echo "Usage: multiplan <task> [--req ...] [--con ...]" >&2
  exit 1
fi

# ── Output dir ────────────────────────────────────────────────────────────────
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
if [[ -z "$OUT_DIR" ]]; then
  OUT_DIR="$HOME/.multiplan/runs/$TIMESTAMP"
fi
mkdir -p "$OUT_DIR"

log() { echo "[multiplan] $*"; }
verbose() { [[ $VERBOSE -eq 1 ]] && echo "[multiplan:verbose] $*" || true; }

# ── Template rendering via Python (handles multiline content) ─────────────────
RENDER_PY="$SCRIPT_DIR/render.py"

render_template() {
  python3 "$RENDER_PY" "$PROMPTS_DIR/plan.md" \
    --var TASK "$TASK" \
    --var REQUIREMENTS "$REQUIREMENTS" \
    --var CONSTRAINTS "$CONSTRAINTS"
}

render_debate_template() {
  local plan_a="$1"
  local plan_b="$2"
  local plan_c="$3"
  local plan_d="$4"
  python3 "$RENDER_PY" "$PROMPTS_DIR/debate.md" \
    --var TASK "$TASK" \
    --file PLAN_A "$plan_a" \
    --file PLAN_B "$plan_b" \
    --file PLAN_C "$plan_c" \
    --file PLAN_D "$plan_d"
}

render_converge_template() {
  local plan_a="$1"
  local plan_b="$2"
  local plan_c="$3"
  local plan_d="$4"
  local debate="$5"
  python3 "$RENDER_PY" "$PROMPTS_DIR/converge.md" \
    --var TASK "$TASK" \
    --file PLAN_A "$plan_a" \
    --file PLAN_B "$plan_b" \
    --file PLAN_C "$plan_c" \
    --file PLAN_D "$plan_d" \
    --file DEBATE "$debate"
}

# ── Check dependencies ────────────────────────────────────────────────────────
for cmd in claude gemini codex; do
  if ! command -v "$cmd" &>/dev/null; then
    log "WARNING: '$cmd' not found in PATH — that model will be skipped"
  fi
done

# ── Phase 1: Independent planning (parallel) ──────────────────────────────────
log "Phase 1 — Running 4 models in parallel..."

PLAN_PROMPT=$(render_template)

PLAN_A="$OUT_DIR/plan-claude.md"
PLAN_B="$OUT_DIR/plan-gemini.md"
PLAN_C="$OUT_DIR/plan-codex.md"
PLAN_D="$OUT_DIR/plan-glm5.md"
GLM_PY="$SCRIPT_DIR/glm.py"

run_claude() {
  log "  → Claude (Opus) thinking..."
  if command -v claude &>/dev/null; then
    printf '%s' "$PLAN_PROMPT" | claude --print --permission-mode bypassPermissions > "$PLAN_A" 2>&1
    log "  ✓ Claude done"
  else
    echo "[Claude not available]" > "$PLAN_A"
    log "  ✗ Claude skipped (not installed)"
  fi
}

run_gemini() {
  log "  → Gemini thinking..."
  if command -v gemini &>/dev/null; then
    gemini -p "$PLAN_PROMPT" > "$PLAN_B" 2>&1
    log "  ✓ Gemini done"
  else
    echo "[Gemini not available]" > "$PLAN_B"
    log "  ✗ Gemini skipped (not installed)"
  fi
}

run_codex() {
  log "  → Codex (GPT) thinking..."
  if command -v codex &>/dev/null; then
    # Codex needs PTY for interactive mode; use exec with a simple query approach
    # Write prompt to temp file and use codex in headless fashion
    local tmp_prompt
    tmp_prompt=$(mktemp /tmp/multiplan-codex-XXXXXX.md)
    echo "$PLAN_PROMPT" > "$tmp_prompt"
    codex exec --full-auto "Read the planning prompt at $tmp_prompt and output your technical plan to stdout. The task: $TASK" \
      > "$PLAN_C" 2>&1 || echo "[Codex output error — check $PLAN_C]" >> "$PLAN_C"
    rm -f "$tmp_prompt"
    log "  ✓ Codex done"
  else
    echo "[Codex not available]" > "$PLAN_C"
    log "  ✗ Codex skipped (not installed)"
  fi
}

run_glm() {
  log "  → GLM-5 (ZhipuAI) thinking..."
  if [[ -f "$GLM_PY" ]]; then
    printf '%s' "$PLAN_PROMPT" | python3 "$GLM_PY" > "$PLAN_D" 2>&1
    log "  ✓ GLM-5 done"
  else
    echo "[GLM-5 not available — glm.py not found]" > "$PLAN_D"
    log "  ✗ GLM-5 skipped"
  fi
}

# Run all four in parallel
run_claude &
PID_A=$!
run_gemini &
PID_B=$!
run_codex &
PID_C=$!
run_glm &
PID_D=$!

wait $PID_A $PID_B $PID_C $PID_D
log "Phase 1 complete. Plans at $OUT_DIR/"

# ── Phase 2: Cross-examination (debate) ───────────────────────────────────────
log "Phase 2 — Cross-examination..."

DEBATE_PROMPT=$(render_debate_template "$PLAN_A" "$PLAN_B" "$PLAN_C" "$PLAN_D")
DEBATE="$OUT_DIR/debate.md"

if command -v claude &>/dev/null; then
  printf '%s' "$DEBATE_PROMPT" | claude --print --permission-mode bypassPermissions > "$DEBATE" 2>&1
  log "  ✓ Debate complete (via Claude)"
elif command -v gemini &>/dev/null; then
  gemini -p "$DEBATE_PROMPT" > "$DEBATE" 2>&1
  log "  ✓ Debate complete (via Gemini)"
else
  echo "[No model available for debate phase]" > "$DEBATE"
  log "  ✗ Debate skipped"
fi

# ── Phase 3: Convergence ──────────────────────────────────────────────────────
log "Phase 3 — Convergence..."

CONVERGE_PROMPT=$(render_converge_template "$PLAN_A" "$PLAN_B" "$PLAN_C" "$PLAN_D" "$DEBATE")
FINAL="$OUT_DIR/final-plan.md"

# Add metadata header
{
  echo "# Multimodel Plan: $TASK"
  echo ""
  echo "> Generated: $(date)"
  echo "> Models: Claude Opus, Gemini, Codex/GPT, GLM-5"
  echo ""
  echo "---"
  echo ""
} > "$FINAL"

if command -v claude &>/dev/null; then
  printf '%s' "$CONVERGE_PROMPT" | claude --print --permission-mode bypassPermissions >> "$FINAL" 2>&1
  log "  ✓ Convergence complete (via Claude)"
elif command -v gemini &>/dev/null; then
  gemini -p "$CONVERGE_PROMPT" >> "$FINAL" 2>&1
  log "  ✓ Convergence complete (via Gemini)"
else
  echo "[No model available for convergence phase]" >> "$FINAL"
  log "  ✗ Convergence skipped"
fi

# ── Output summary ────────────────────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════"
echo " Multimodel Planning Complete"
echo "════════════════════════════════════════"
echo ""
echo " Task: $TASK"
echo ""
echo " Outputs:"
echo "   Plan A (Claude): $PLAN_A"
echo "   Plan B (Gemini): $PLAN_B"
echo "   Plan C (Codex):  $PLAN_C"
echo "   Plan D (GLM-5):  $PLAN_D"
echo "   Debate:          $DEBATE"
echo "   Final Plan:      $FINAL"
echo ""
echo " → Final plan: $FINAL"
echo ""

# Print final plan to stdout
cat "$FINAL"
