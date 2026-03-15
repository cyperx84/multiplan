#!/usr/bin/env bash
# release.sh — publish multiplan to npm, PyPI, and tag for Homebrew
# Usage: ./scripts/release.sh [--version 0.2.0] [--npm] [--pypi] [--tag] [--all]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

VERSION=""
DO_NPM=0
DO_PYPI=0
DO_TAG=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version|-v) VERSION="$2"; shift 2 ;;
    --npm) DO_NPM=1; shift ;;
    --pypi) DO_PYPI=1; shift ;;
    --tag) DO_TAG=1; shift ;;
    --all) DO_NPM=1; DO_PYPI=1; DO_TAG=1; shift ;;
    *) echo "Unknown: $1"; exit 1 ;;
  esac
done

cd "$REPO_DIR"

# ── Version bump ───────────────────────────────────────────────────────────────
if [[ -n "$VERSION" ]]; then
  echo "Bumping version to $VERSION..."
  # package.json
  node -e "
    const fs = require('fs');
    const p = JSON.parse(fs.readFileSync('package.json','utf8'));
    p.version = '$VERSION';
    fs.writeFileSync('package.json', JSON.stringify(p, null, 2) + '\n');
    console.log('  ✓ package.json');
  "
  # pyproject.toml
  sed -i '' "s/^version = .*/version = \"$VERSION\"/" py/pyproject.toml
  echo "  ✓ py/pyproject.toml"

  # homebrew formula
  sed -i '' "s|/v[0-9.]*/|/v$VERSION/|g" homebrew/multiplan.rb
  echo "  ✓ homebrew/multiplan.rb"
fi

CURRENT_VERSION=$(node -e "const p=require('./package.json'); console.log(p.version)")
echo "Version: $CURRENT_VERSION"

# ── Build ──────────────────────────────────────────────────────────────────────
echo ""
echo "Building..."
npm run build
npm test
echo "✓ Build + tests pass"

# ── npm publish ────────────────────────────────────────────────────────────────
if [[ $DO_NPM -eq 1 ]]; then
  echo ""
  echo "Publishing to npm..."
  npm publish --access public
  echo "✓ Published to npm: multiplan@$CURRENT_VERSION"
  echo "  Install: npm install -g multiplan"
fi

# ── PyPI publish ───────────────────────────────────────────────────────────────
if [[ $DO_PYPI -eq 1 ]]; then
  echo ""
  echo "Publishing to PyPI..."
  cd py
  if command -v uv &>/dev/null; then
    uv build
    uv publish
    echo "✓ Published to PyPI: multiplan $CURRENT_VERSION"
    echo "  Install: uv tool install multiplan"
    echo "  Install: pip install multiplan"
  else
    echo "uv not found — install with: curl -LsSf https://astral.sh/uv/install.sh | sh"
    exit 1
  fi
  cd "$REPO_DIR"
fi

# ── Git tag (for Homebrew) ─────────────────────────────────────────────────────
if [[ $DO_TAG -eq 1 ]]; then
  echo ""
  echo "Tagging v$CURRENT_VERSION..."
  git add package.json py/pyproject.toml homebrew/multiplan.rb
  git -c commit.gpgsign=false commit -m "chore: release v$CURRENT_VERSION" 2>/dev/null || echo "(nothing to commit)"
  git tag -a "v$CURRENT_VERSION" -m "Release v$CURRENT_VERSION"
  git push origin main
  git push origin "v$CURRENT_VERSION"
  echo "✓ Tagged v$CURRENT_VERSION and pushed"
  echo ""
  echo "To update the Homebrew formula sha256:"
  echo "  1. Wait for GitHub to create the tarball"
  echo "  2. curl -L https://github.com/cyperx84/multiplan/archive/refs/tags/v$CURRENT_VERSION.tar.gz | shasum -a 256"
  echo "  3. Update homebrew/multiplan.rb sha256"
  echo "  4. Submit to homebrew-core or use a tap: https://github.com/cyperx84/homebrew-cyperx"
fi

echo ""
echo "════════════════════════════════"
echo " Release complete: v$CURRENT_VERSION"
echo "════════════════════════════════"
echo ""
[[ $DO_NPM -eq 1 ]]  && echo "  npm:     npm install -g multiplan"
[[ $DO_PYPI -eq 1 ]] && echo "  uv:      uv tool install multiplan"
[[ $DO_PYPI -eq 1 ]] && echo "  pip:     pip install multiplan"
[[ $DO_TAG -eq 1 ]]  && echo "  brew:    brew install cyperx84/cyperx/multiplan (after tap setup)"
echo ""
