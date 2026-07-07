#!/usr/bin/env python3
"""
Jules Batch Submitter for Substrate
Sends tasks to Jules API to create async coding sessions → GitHub PRs.

Usage:
  python3 jules_submit.py               # List all available tasks
  python3 jules_submit.py --task 1      # Submit specific task by number
  python3 jules_submit.py --list        # List available tasks
  python3 jules_submit.py --status      # Check recent session status
  python3 jules_submit.py --file path   # Submit a custom prompt from a file
  python3 jules_submit.py --branch feat # Target a specific branch
"""

import json
import urllib.request
import sys
import os

# ──────────────────────────────────────────────────────────────────────────────
# Config — loads API key from .env.local or .env (never hardcode secrets)
# ──────────────────────────────────────────────────────────────────────────────

def _load_api_key():
    """Read JULES_API_KEY from environment, .env.local, or .env."""
    key = os.environ.get("JULES_API_KEY")
    if key:
        return key
    for envfile in [".env.local", ".env"]:
        path = os.path.join(os.path.dirname(os.path.abspath(__file__)), envfile)
        if os.path.exists(path):
            with open(path) as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("JULES_API_KEY="):
                        return line.split("=", 1)[1].strip()
    print("❌ JULES_API_KEY not found in environment, .env.local, or .env")
    sys.exit(1)

API_KEY = _load_api_key()
API_URL = "https://jules.googleapis.com/v1alpha/sessions"

# ── Repo root (one level up from scripts/) ────────────────────────────────────
REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# ── Update this once the GitHub repo is created ──────────────────────────────
REPO_SOURCE = "sources/github/KrushnaVardhanReddy/Substrate"

# Parse branch from args if provided
BRANCH = "main"
if "--branch" in sys.argv:
    idx = sys.argv.index("--branch")
    if idx + 1 < len(sys.argv):
        BRANCH = sys.argv[idx + 1]

# ──────────────────────────────────────────────────────────────────────────────
# Mandatory safety rules (prepended to every prompt)
# ──────────────────────────────────────────────────────────────────────────────

SAFETY_RULES = """
MANDATORY RULES — VIOLATION = REJECTED PR:
1. NEVER stub, mock, or TODO existing implementation code.
2. Every file you modify MUST still build — run lint/typecheck/go vet before committing.
3. If a test fails, FIX the code or test — do NOT delete or skip tests.
4. Do NOT alter any file in docs/specs/ — those are the source of truth. Implement from them, never rewrite them.
5. All new Go packages MUST have at least one table-driven unit test.
6. Commit message must start with "jules: " prefix.
7. Follow the tech stack exactly as listed below.
8. 100% SPEC-FIRST RULE: If your implementation deviates from the spec in docs/specs/, STOP and flag it — do not silently change behaviour.

Project: Substrate — CI/CD-integrated data contract and dependency intelligence platform.
Tech stack:
- Diff Engine CLI: Go (modules, standard library preferred)
- GitHub App: TypeScript (Node.js, Octokit SDK)
- API Server: Go + PostgreSQL
- Frontend: SvelteKit + TypeScript + Tailwind CSS
- Tests: Go table-driven tests (engine), Vitest (GitHub App / API), Playwright (e2e)

Repo layout (to be established):
  engine/         — Go diff engine (core library + CLI binary)
  github-app/     — TypeScript GitHub App (webhook handler)
  api/            — Go REST API server
  dashboard/      — SvelteKit frontend
  docs/specs/     — Source-of-truth specification files (READ ONLY for Jules)
  prompts/        — Jules/Stitch task prompts
  scripts/        — Automation scripts (jules_submit.py, stitch_submit.py)
""".strip()

# ──────────────────────────────────────────────────────────────────────────────
# Task definitions — populate as implementation progresses
# ──────────────────────────────────────────────────────────────────────────────

TASKS = {
    # ── Phase 0: Specs (handled by Antigravity, not Jules) ────────────────────

    # ── Phase 1: Go Diff Engine ───────────────────────────────────────────────
    1: {
        "name": "P1-T01 — Go Module Scaffold",
        "phase": "phase-1-diff-engine",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-1-diff-engine/t01_go_scaffold.txt")
        ).read()
    },
    2: {
        "name": "P1-T02 — OpenAPI Parser + IR",
        "phase": "phase-1-diff-engine",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-1-diff-engine/t02_openapi_parser.txt")
        ).read()
    },
    3: {
        "name": "P1-T03 — Diff Comparator",
        "phase": "phase-1-diff-engine",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-1-diff-engine/t03_diff_comparator.txt")
        ).read()
    },
    4: {
        "name": "P1-T04 — Rule Engine",
        "phase": "phase-1-diff-engine",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-1-diff-engine/t04_rule_engine.txt")
        ).read()
    },
    5: {
        "name": "P1-T05 — CLI + JSON Output",
        "phase": "phase-1-diff-engine",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-1-diff-engine/t05_cli_output.txt")
        ).read()
    },
    6: {
        "name": "P1-T06 — Unit Tests",
        "phase": "phase-1-diff-engine",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-1-diff-engine/t06_unit_tests.txt")
        ).read()
    },

    # ── Phase 2: GitHub App ───────────────────────────────────────────────────
    10: {
        "name": "P2-T01 — GitHub App Scaffold + Webhook",
        "phase": "phase-2-github-app",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-2-github-app/t01_github_app_scaffold.txt")
        ).read()
    },

    # ── Phase 3: Dashboard ────────────────────────────────────────────────────
    20: {
        "name": "P3-T01 — Go API Server Scaffold",
        "phase": "phase-3-dashboard",
        "prompt": open(
            os.path.join(REPO_ROOT,
            "prompts/phase-3-dashboard/t01_api_server.txt")
        ).read()
    },
}

# ──────────────────────────────────────────────────────────────────────────────
# Submission logic
# ──────────────────────────────────────────────────────────────────────────────

def submit_task(task_num):
    if task_num not in TASKS:
        print(f"❌ Task {task_num} not found. Use --list to see available tasks.")
        sys.exit(1)

    task = TASKS[task_num]
    full_prompt = SAFETY_RULES + "\n\n---\n\n" + task["prompt"]

    payload = json.dumps({
        "agent_settings": {
            "model_id": "GEMINI_2_5_PRO"
        },
        "context": {
            "repository": {
                "resource_name": REPO_SOURCE,
                "branch": BRANCH
            }
        },
        "prompt": full_prompt
    }).encode()

    req = urllib.request.Request(
        API_URL,
        data=payload,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {API_KEY}"
        },
        method="POST"
    )

    print(f"🚀 Submitting: [{task_num}] {task['name']} → branch: {BRANCH}")
    try:
        with urllib.request.urlopen(req) as resp:
            result = json.loads(resp.read())
            session_id = result.get("name", "unknown").split("/")[-1]
            print(f"✅ Session created: {session_id}")
            print(f"   View at: https://jules.google.com/")
    except urllib.error.HTTPError as e:
        print(f"❌ HTTP {e.code}: {e.read().decode()}")
        sys.exit(1)


def submit_file(filepath):
    if not os.path.exists(filepath):
        print(f"❌ File not found: {filepath}")
        sys.exit(1)
    with open(filepath) as f:
        prompt_content = f.read()

    full_prompt = SAFETY_RULES + "\n\n---\n\n" + prompt_content
    payload = json.dumps({
        "agent_settings": {"model_id": "GEMINI_2_5_PRO"},
        "context": {
            "repository": {
                "resource_name": REPO_SOURCE,
                "branch": BRANCH
            }
        },
        "prompt": full_prompt
    }).encode()

    req = urllib.request.Request(
        API_URL,
        data=payload,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {API_KEY}"
        },
        method="POST"
    )

    print(f"🚀 Submitting custom prompt from: {filepath} → branch: {BRANCH}")
    try:
        with urllib.request.urlopen(req) as resp:
            result = json.loads(resp.read())
            session_id = result.get("name", "unknown").split("/")[-1]
            print(f"✅ Session created: {session_id}")
    except urllib.error.HTTPError as e:
        print(f"❌ HTTP {e.code}: {e.read().decode()}")
        sys.exit(1)


def list_tasks():
    print("\n📋 Available Substrate Jules Tasks:\n")
    for num, task in sorted(TASKS.items()):
        print(f"  [{num:>3}] {task['name']}  ({task['phase']})")
    print()


def main():
    args = sys.argv[1:]

    if not args or "--help" in args or "-h" in args:
        print(__doc__)
        sys.exit(0)

    if "--list" in args:
        list_tasks()
        sys.exit(0)

    if "--file" in args:
        idx = args.index("--file")
        if idx + 1 >= len(args):
            print("❌ Please specify a file path after --file.")
            sys.exit(1)
        submit_file(args[idx + 1])
        sys.exit(0)

    if "--task" in args:
        idx = args.index("--task")
        if idx + 1 >= len(args):
            print("❌ Please specify a task number after --task.")
            sys.exit(1)
        submit_task(int(args[idx + 1]))
        sys.exit(0)

    # Default: list tasks
    list_tasks()


if __name__ == "__main__":
    main()
