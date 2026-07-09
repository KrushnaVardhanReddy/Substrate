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
        # Look in the repo root, not the scripts/ directory
        repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        path = os.path.join(repo_root, envfile)
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
BRANCH = "feature/dev"
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
5. TESTING IS MANDATORY — every task MUST include both:
   a. Unit tests: table-driven `_test.go` files (Go) or `*.test.ts` (TypeScript). Min 80% coverage on new code.
   b. E2E tests: for any CLI command, invoke the compiled binary with real input files and assert stdout/exit code.
   A PR with no tests will be rejected, no exceptions.
6. Commit message must start with "jules: " prefix.
7. 100% SPEC-FIRST RULE: If your implementation deviates from the spec in docs/specs/, STOP and flag it.

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

def _load_prompt(relative_path):
    """Load a prompt file lazily — returns its content or an error string."""
    full_path = os.path.join(REPO_ROOT, relative_path)
    if not os.path.exists(full_path):
        return f"ERROR: Prompt file not found: {full_path}"
    with open(full_path) as f:
        return f.read()


TASKS = {
    # ── Phase 0: Specs (handled by Antigravity, not Jules) ────────────────────
    # (no Jules tasks for Phase 0)

    # ── Phase 1a: OpenAPI diff engine (oasdiff-powered) ───────────────────────
    # Architecture: oasdiff Go library wraps OpenAPI diffing. We build the
    # adapter, override config parser, CLI, and tests on top.
    1: {
        "name": "P1-T01 — Go Module Scaffold + oasdiff Dependency",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t01_go_scaffold.txt"),
    },
    2: {
        "name": "P1-T02 — oasdiff Adapter (oasdiff output → DiffReport)",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t02_oasdiff_adapter.txt"),
    },
    3: {
        "name": "P1-T03 — Override Config Parser + CLI",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t03_cli_override.txt"),
    },
    4: {
        "name": "P1-T04 — Unit Tests + E2E Tests",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t04_tests.txt"),
    },
    6: {
        "name": "P1-T06 — oasdiff Checker Adapter (semantic 37-rule engine)",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t06_checker_adapter.txt"),
    },
    7: {
        "name": "P1-T07 — substrate init command",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t07_init_command.txt"),
    },

    # ── Phase 1b: SQL Migrations ─────────────────────────────────────────────
    101: {
        "name": "P1b-T02 — SQL DDL Parser + Diff Engine + Rule Engine",
        "phase": "phase-1b-sql",
        "prompt": _load_prompt("prompts/phase-1b-sql/t02_sql_parser.txt"),
    },
    102: {
        "name": "P1b-T03 — SQL Rule Engine: Expand Coverage + Edge-Case Tests",
        "phase": "phase-1b-sql",
        "prompt": _load_prompt("prompts/phase-1b-sql/t03_sql_rule_engine_tests.txt"),
    },

    # ── Phase 1d: Protobuf & gRPC (Wave 2 — no deps, submit now) ─────────────
    # Independent of Phase 3. Just adds proto support to the engine.
    # buf is invoked as a binary via exec.Command — no new Go dependencies.
    103: {
        "name": "P1d-T01 — Protobuf & gRPC Breaking Change Adapter (buf checker)",
        "phase": "phase-1d-protobuf",
        "prompt": _load_prompt("prompts/phase-1d-protobuf/t01_buf_adapter.txt"),
    },

    # ── Phase 2: GitHub App ───────────────────────────────────────────────────
    # Dependency map:
    #   T01, T02a, T03 → NO mutual dependencies → submit all 3 in parallel
    #   T02b (wire together) → needs T01 + T02a merged first
    #   T04 (status check)  → needs T02b
    #   T05 (deploy)        → needs T01+T02a+T03+T04 all merged
    10: {
        "name": "P2-T01 — GitHub App Scaffold + Cloudflare Worker Webhook Receiver",
        "phase": "phase-2-github-app",
        "prompt": _load_prompt("prompts/phase-2-github-app/t01_github_app_scaffold.txt"),
    },
    11: {
        "name": "P2-T02a — substrate serve HTTP Mode (Container Service Binary)",
        "phase": "phase-2-github-app",
        "prompt": _load_prompt("prompts/phase-2-github-app/t02a_binary_serve_mode.txt"),
    },
    12: {
        "name": "P2-T03 — PR Comment Formatter (Standalone TypeScript Module)",
        "phase": "phase-2-github-app",
        "prompt": _load_prompt("prompts/phase-2-github-app/t03_pr_comment_formatter.txt"),
    },
    13: {
        "name": "P2-T02b — Wire Worker → Container Service → GitHub APIs (Real Diff Results)",
        "phase": "phase-2-github-app",
        "prompt": _load_prompt("prompts/phase-2-github-app/t02b_wire_worker_container.txt"),
    },

    # ── Phase 3: Contract Registry (P1 tasks — no dependencies, submit in parallel) ──
    # P1 (Core Registry) tasks — P3-T01 and P3-T02 have NO dependencies → submit together
    # P3-T02b and P3-T02c require T01+T02 merged first → submit in second wave
    20: {
        "name": "P3-T01 — Go API Server Scaffold + PostgreSQL Schema",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t01_api_server.txt"),
    },
    21: {
        "name": "P3-T02 — substrate.yaml Consumer Declaration Parser",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t02_consumer_parser.txt"),
    },
    22: {
        "name": "P3-T02b — Contract Registry Sync (push-to-main webhook handler)",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t02b_contract_sync.txt"),
    },
    23: {
        "name": "P1e-T01 — AsyncAPI breaking change adapter",
        "phase": "phase-1e-asyncapi-avro",
        "prompt": _load_prompt("prompts/phase-1e-asyncapi-avro/t01_asyncapi_adapter.txt"),
    },
    24: {
        "name": "P1e-T02 — Apache Avro Schema Registry compatibility adapter",
        "phase": "phase-1e-asyncapi-avro",
        "prompt": _load_prompt("prompts/phase-1e-asyncapi-avro/t02_avro_adapter.txt"),
    },
    25: {
        "name": "P1h-T01 — Terraform Plan JSON adapter",
        "phase": "phase-1h-iac",
        "prompt": _load_prompt("prompts/phase-1h-iac/t01_terraform_adapter.txt"),
    },
    26: {
        "name": "P1c-T01 — GraphQL Schema Diff Adapter",
        "phase": "phase-1c-graphql",
        "prompt": _load_prompt("prompts/phase-1c-graphql/t01_parser.txt"),
    },
    27: {
        "name": "P3-T02c — Cross-Repo Compatibility Check on PR",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t02c_cross_repo_check.txt"),
    },
    28: {
        "name": "P1f-T01 — AI/ML Model Contract Diff Adapter",
        "phase": "phase-1f-aiml",
        "prompt": _load_prompt("prompts/phase-1f-aiml/t01_aiml_adapter.txt"),
    },
    29: {
        "name": "P3-T03 — SvelteKit Dashboard & Design System",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t03_dashboard.txt"),
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
        "prompt": full_prompt,
        "sourceContext": {
            "source": REPO_SOURCE,
            "githubRepoContext": {
                "startingBranch": BRANCH
            }
        }
    }).encode()

    req = urllib.request.Request(
        API_URL,
        data=payload,
        headers={
            "Content-Type": "application/json",
            "x-goog-api-key": API_KEY
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
        "prompt": full_prompt,
        "sourceContext": {
            "source": REPO_SOURCE,
            "githubRepoContext": {
                "startingBranch": BRANCH
            }
        }
    }).encode()

    req = urllib.request.Request(
        API_URL,
        data=payload,
        headers={
            "Content-Type": "application/json",
            "x-goog-api-key": API_KEY
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
