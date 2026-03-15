#!/usr/bin/env python3
"""Call GLM-5 via ZAI OpenAI-compatible API.
Reads the API key from OpenClaw's auth-profiles.json.
Usage: echo "prompt" | python3 glm.py
   or: python3 glm.py "prompt text"
"""
import sys
import json
import urllib.request
import urllib.error
import os
from pathlib import Path

BASE_URL = "https://api.z.ai/api/coding/paas/v4"
MODEL = "glm-5"

def get_key() -> str:
    # 1. Env override
    key = os.environ.get("ZAI_API_KEY") or os.environ.get("GLM_API_KEY")
    if key:
        return key

    # 2. OpenClaw auth-profiles.json (main agent — has the full key store)
    paths = [
        Path.home() / ".openclaw/agents/main/agent/auth-profiles.json",
        Path.home() / ".openclaw/agents/builder/agent/auth-profiles.json",
    ]
    for p in paths:
        if p.exists():
            data = json.loads(p.read_text())
            profile = data.get("profiles", {}).get("zai:default", {})
            k = profile.get("key", "")
            if k:
                return k

    raise RuntimeError("ZAI API key not found. Set ZAI_API_KEY env var or ensure ~/.openclaw/agents/main/agent/auth-profiles.json has zai:default profile.")

def call(prompt: str) -> str:
    key = get_key()
    payload = json.dumps({
        "model": MODEL,
        "messages": [{"role": "user", "content": prompt}],
        "max_tokens": 8192,
        "temperature": 0.7,
    }).encode()

    req = urllib.request.Request(
        f"{BASE_URL}/chat/completions",
        data=payload,
        headers={
            "Authorization": f"Bearer {key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )

    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            data = json.loads(resp.read())
            return data["choices"][0]["message"]["content"]
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        raise RuntimeError(f"GLM-5 API error {e.code}: {body[:500]}")

def main():
    if len(sys.argv) > 1:
        prompt = " ".join(sys.argv[1:])
    else:
        prompt = sys.stdin.read()

    if not prompt.strip():
        print("[GLM-5: empty prompt]", file=sys.stderr)
        sys.exit(1)

    result = call(prompt)
    print(result)

if __name__ == "__main__":
    main()
