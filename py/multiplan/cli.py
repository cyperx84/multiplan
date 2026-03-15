"""
multiplan Python CLI wrapper.

Delegates to the Node.js multiplan CLI. Installs it via npm if not found.
"""
import sys
import shutil
import subprocess
import os


def find_multiplan_bin() -> str | None:
    """Find the multiplan Node.js binary."""
    # 1. Already on PATH
    if path := shutil.which("multiplan-node"):
        return path
    # 2. npm global bin
    try:
        npm_bin = subprocess.check_output(
            ["npm", "bin", "-g"], text=True, stderr=subprocess.DEVNULL
        ).strip()
        candidate = os.path.join(npm_bin, "multiplan")
        if os.path.isfile(candidate):
            return candidate
    except Exception:
        pass
    # 3. npx fallback
    if shutil.which("npx"):
        return None  # will use npx
    return None


def ensure_installed() -> list[str]:
    """Return the command prefix to run multiplan."""
    bin_path = find_multiplan_bin()
    if bin_path:
        return [bin_path]
    # Try npx
    if shutil.which("npx"):
        return ["npx", "--yes", "multiplan"]
    # Try to install
    if shutil.which("npm"):
        print("[multiplan] Installing via npm...", file=sys.stderr)
        subprocess.run(["npm", "install", "-g", "multiplan"], check=True)
        if path := find_multiplan_bin():
            return [path]
    raise RuntimeError(
        "multiplan Node.js CLI not found. Install with: npm install -g multiplan"
    )


def main() -> None:
    """Entry point — delegate all args to the Node.js multiplan CLI."""
    try:
        cmd = ensure_installed()
    except RuntimeError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

    result = subprocess.run(cmd + sys.argv[1:])
    sys.exit(result.returncode)


if __name__ == "__main__":
    main()
