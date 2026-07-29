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
import subprocess
def get_current_branch():
    try:
        return subprocess.check_output(['git', 'rev-parse', '--abbrev-ref', 'HEAD'], text=True).strip()
    except Exception:
        return "feature/dev"

BRANCH = get_current_branch()
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
   c. Frontend UI: Any changes to Svelte components MUST include accompanying unit tests (`*.test.ts`) using vitest and `@testing-library/svelte`.
   A PR with no tests will be rejected, no exceptions.
6. Commit message must start with "jules: " prefix.
7. 100% SPEC-FIRST RULE: If your implementation deviates from the spec in docs/specs/, STOP and flag it.
8. LLM-WIKI MANDATE: You MUST read `CLAUDE.md` and `docs/wiki/index.md` before writing code to establish project-wide architectural context.

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
    1401: {
        "name": "P14-T01 — AI Contract Negotiation",
        "phase": "phase-14",
        "prompt": _load_prompt("prompts/phase-14/t01_contract_negotiation.txt"),
    },
    1014: {
        "name": "P10-T14 — Custom Governance (Goja)",
        "phase": "phase-10-ecosystem",
        "prompt": _load_prompt("prompts/phase-10/t14_custom_governance.txt"),
    },
    916: {
        "name": "P9-T16 — WASM Git Pre-Commit Hooks",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t16_wasm_hooks.txt"),
    },
    1018: {
        "name": "P10-T18 — GraphQL Supergraph Federation",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t18_graphql_federation.txt"),
    },
    1010: {
        "name": "P10-T10 — Consumer-Driven Contract Manifests",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t10_consumer_contracts.txt"),
    },
    1502: {
        "name": "P15-T02 — Granular GitHub Check Suite",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t02_granular_check_suite.txt"),
    },
    1001: {
        "name": "P10-T01 — Automated Schema Robustness Validation (QA Fuzzing)",
        "phase": "phase-10-ecosystem",
        "prompt": _load_prompt("prompts/phase-10/t01_security_fuzzing.txt"),
    },
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
    999: {
        "name": "P1-T09 — Phase 1f (AI/ML) and 1g (Salesforce) E2E Validation",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t09_phase1f_1g_e2e_validation.txt"),
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
    30: {
        "name": "P3-T02d — Cross-Repo E2E Fixture Tests",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t02d_cross_repo_e2e.txt"),
    },
    31: {
        "name": "P1g-T01 — Enterprise Metadata Diff Adapter (Salesforce & SOAP)",
        "phase": "phase-1g-enterprise",
        "prompt": _load_prompt("prompts/phase-1g-enterprise/t01_enterprise_adapter.txt"),
    },
    32: {
        "name": "P3-T09 — MCP Server Implementation",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t09_mcp_implementation.txt"),
    },
    33: {
        "name": "P3-T09b — Wire MCP Server to live Registry API",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t09b_wire_mcp_registry.txt"),
    },
    34: {
        "name": "P3-T14 — Interactive Diff Viewer URL in GitHub PR Comments",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t14_interactive_diff_url.txt"),
    },
    35: {
        "name": "P3-T06 — GitHub OAuth and Org ACL Middleware",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t06_auth_org.txt"),
    },
    36: {
        "name": "P3-T04 — Connected Repos List & Schema Browser",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t04_dashboard_repos.txt"),
    },
    37: {
        "name": "P3-T07 — Free Tier Limits + Production Deployment",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t07_rate_limits.txt"),
    },
    38: {
        "name": "P3-T10 — MCP Deployment + IDE Integration Docs",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t10_mcp_docs.txt"),
    },

    # ── Phase 4: AI Intelligence Layer ───────────────────────────────────────────
    40: {
        "name": "P4-T02 — Streaming SSE AI Analyze Handler (Go API)",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t02_streaming_sse_ai_handler.txt"),
    },
    41: {
        "name": "P4-T03 — Safe Schema Patch Generator + PR Comment Upgrade",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t03_safe_schema_patch_generator.txt"),
    },
    42: {
        "name": "P4-T04 — GitHub PR Comment Upgrade (Wire AI Autofix)",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t04_github_pr_comment_upgrade.txt"),
    },
    46: {
        "name": "P4-T06 — Shift-Left IDE Extension (VSCode)",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t06_vscode_extension.txt"),
    },
    47: {
        "name": "P4-T07 — Traffic-Aware Diffing (Zero False Positives)",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t07_traffic_aware_diffing.txt"),
    },
    48: {
        "name": "P4-T08 — PII & Compliance Auditing",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t08_compliance_auditing.txt"),
    },
    9: {
        "name": "P4-T09 — Breaking Change History (MCP)",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t09_breaking_change_history.txt"),
    },

    # ── Phase 5: Automated Dependency Discovery (Enterprise) ────────────────────
    51: {
        "name": "P5-T01 — Env Var & URL Registry Scanner",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t01_env_scanner.txt"),
    },
    52: {
        "name": "P5-T02 — Package & Generator Scanners",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t02_package_scanner.txt"),
    },
    53: {
        "name": "P5-T03 — Terraform & UI Confidence Scoring",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t03_terraform_and_ui.txt"),
    },
    61: {
        "name": "P6-T05 — Phase 6 E2E Tests",
        "phase": "phase-6-qa",
        "prompt": _load_prompt("prompts/phase-6-qa/t05_e2e_tests.txt"),
    },
    104: {
        "name": "UI-T01 — Dynamic Cytoscape Graph Rendering",
        "phase": "ui",
        "prompt": _load_prompt("prompts/ui/t01_dynamic_graph.txt"),
    },
    105: {
        "name": "E2E-T01 — 1-Hour Chaos Endurance Mode & Live UI Polling",
        "phase": "ui",
        "prompt": _load_prompt("prompts/e2e/t01_endurance_mode.txt"),
    },
    106: {
        "name": "UI-T03 — Graph Filtering & Navigation",
        "phase": "ui",
        "prompt": _load_prompt("prompts/ui/t03_graph_filtering.txt"),
    },
    56: {
        "name": "P5-T06 — Phase 5 E2E Tests",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t06_e2e_tests.txt"),
    },
    57: {
        "name": "P5-T07 — 100-Repo Scale & Noise Simulation",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t07_scale_simulation.txt"),
    },
    58: {
        "name": "P5-T08 — Chi Router Migration & Panic Recovery",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t08_chi_router.txt"),
    },
    54: {
        "name": "P5-T04 — Event-Driven Discovery",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t04_event_discovery.txt"),
    },
    55: {
        "name": "P5-T05 — Runtime Confirmation",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t05_runtime_confirmation.txt"),
    },

    # ── Phase 6: QA & Automation Layer ──────────────────────────────────────────
    61: {
        "name": "P6-T01 — Auto-Updating Postman Collections",
        "phase": "phase-6-qa",
        "prompt": _load_prompt("prompts/phase-6-qa/t01_postman_sync.txt"),
    },
    62: {
        "name": "P6-T02 — Shadow API Test Coverage",
        "phase": "phase-6-qa",
        "prompt": _load_prompt("prompts/phase-6-qa/t02_shadow_coverage.txt"),
    },
    63: {
        "name": "P6-T03 — Auto-Generating Test Code",
        "phase": "phase-6-qa",
        "prompt": _load_prompt("prompts/phase-6-qa/t03_test_code_generation.txt"),
    },
    64: {
        "name": "P6-T04 — Mock Server Time Machine",
        "phase": "phase-6-qa",
        "prompt": _load_prompt("prompts/phase-6-qa/t04_mock_server.txt"),
    },
    71: {
        "name": "V1-T01 — GitHub App Auto-Discovery",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t01_github_app_auto_discovery.txt"),
    },
    72: {
        "name": "V1-T02 — CLI AI Architect",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t02_cli_ai_architect.txt"),
    },
    73: {
        "name": "V1-T03 — Interactive Diff Viewer UI",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t03_interactive_diff_viewer.txt"),
    },
    74: {
        "name": "V1-T04 — Deployment Safety Gate",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t04_check_deploy.txt"),
    },
    75: {
        "name": "V1-T05 — Local Validate CLI",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t05_local_validate_cli.txt"),
    },
    76: {
        "name": "V1-T06 — Legal Audit",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t06_legal_audit.txt"),
    },
    77: {
        "name": "V1-T07 — V1.0 System E2E Tests",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t07_system_e2e.txt"),
    },
    78: {
        "name": "V1-T08 — Custom Discovery Rules via YAML",
        "phase": "v1-preflight",
        "prompt": _load_prompt("prompts/v1-preflight/t08_custom_discovery_rules.txt"),
    },

    # ── Phase 7: Enterprise Integrations & ITSM ──────────────────────────────
    80: {
        "name": "P7-T00 — Zero-Config Org Rollout",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t00_zero_config_org_rollout.txt"),
    },
    81: {
        "name": "P7-T01 — Enterprise Webhook & Event Egress",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t01_webhook_egress.txt"),
    },
    83: {
        "name": "P7-T03 — Custom Rules Engine (CEL)",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t03_custom_rules_engine.txt"),
    },
    84: {
        "name": "P7-T04 — Cross-Repo Auto-Fix PRs",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t04_cross_repo_autofix.txt"),
    },
    85: {
        "name": "P7-T05 — Runtime Drift Detection (Sidecar)",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t05_runtime_drift_detection.txt"),
    },
    86: {
        "name": "P7-T06 — Audit Mode (Shadow Mode)",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t06_audit_mode_rollout.txt"),
    },
    87: {
        "name": "P7-T07 — Phase 7 E2E Testing",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t07_e2e_tests.txt"),
    },

    # ── Phase 8: Enterprise Readiness & Scale ────────────────────────────────
    91: {
        "name": "P8-T01 — PostgreSQL Job Queue (River)",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t01_postgres_job_queue.txt"),
    },
    92: {
        "name": "P8-T02 — Enterprise Authz (Casbin/OpenFGA) [Backend]",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t02_enterprise_authz.txt"),
    },
    96: {
        "name": "P8-T06 — Management ROI Dashboard [Backend]",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t06_roi_dashboard.txt"),
    },
    97: {
        "name": "P8-T07 — Billing & Subscription Engine (Paywall Pause)",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t07_billing_engine.txt"),
    },
    93: {
        "name": "P8-T03 — CI/CD Cascading Rollback Gate",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t03_cascading_rollback.txt"),
    },
    94: {
        "name": "P8-T04 — Spotify Backstage Plugin",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t04_backstage_plugin.txt"),
    },
    95: {
        "name": "P8-T05 — Distributed Tracing (OpenTelemetry)",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t05_opentelemetry.txt"),
    },
    98: {
        "name": "P8-T08 — Single Binary VPC Deployment",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t08_single_binary.txt"),
    },
    99: {
        "name": "P8-T09 — Docker & Helm Enterprise Delivery",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t09_docker_helm.txt"),
    },
    100: {
        "name": "P8-T10 — Phase 8 E2E Testing",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t10_e2e_tests.txt"),
    },

    # ── Phase 11: Advanced Graph Visualization (V2.0 UX) ─────────────────────
    1101: {
        "name": "P11-T01 — Cascading Blast Radius (Nth-Degree)",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t01_blast_radius.txt"),
    },
    1102: {
        "name": "P11-T02 — Team Neighborhoods",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t02_team_neighborhoods.txt"),
    },
    1103: {
        "name": "P11-T03 — Interactive Edge Tooltips",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t03_edge_tooltips.txt"),
    },
    1104: {
        "name": "P11-T04 — Historical Volatility Heatmap",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t04_volatility_heatmap.txt"),
    },
    1107: {
        "name": "P11-T07 — Visual API Design Studio",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t07_visual_api_studio.txt"),
    },
    1108: {
        "name": "P11-T08 — Substrate WASM Engine (In-Browser Diffing)",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t08_wasm_engine.txt"),
    },
    1109: {
        "name": "P11-T09 — Server-Sent Events (SSE) Real-Time UI",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t09_sse_ui.txt"),
    },
    110: {
        "name": "P11-T10 — Svelte Flow Migration",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t10_svelte_flow.txt"),
    },
    111: {
        "name": "P11-T11 — Global Command Palette (Cmd+K)",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t11_command_palette.txt"),
    },
    112: {
        "name": "P11-T12 — Rich Side-by-Side Diff Viewer & Sign Out",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t12_diff_viewer.txt"),
    },
    113: {
        "name": "P11-T13 — Time-Travel Graph Replay",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t13_time_travel.txt"),
    },
    114: {
        "name": "P11-T14 — Premium Aesthetics System",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t14_premium_aesthetics.txt"),
    },
    1115: {
        "name": "P11-T15 — Zero-to-One Onboarding Wizard",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t15_onboarding_wizard.txt"),
    },
    1116: {
        "name": "P11-T16 — Taxonomy & Metadata Tagging",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t16_taxonomy_metadata.txt"),
    },
    1117: {
        "name": "P11-T17 — Graph Image Export",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t17_graph_export.txt"),
    },

    # ── Phase 12: V2.0 Public Launch & Quality Assurance ─────────────────
    1201: {
        "name": "P12-T01 — Zero-to-One Onboarding E2E",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t01_onboarding_e2e.txt"),
    },
    1202: {
        "name": "P12-T02 — Svelte Flow Interaction E2E",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t02_svelte_flow_e2e.txt"),
    },
    1203: {
        "name": "P12-T03 — Visual Studio Resilience Test",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t03_studio_resilience_e2e.txt"),
    },
    1204: {
        "name": "P12-T04 — SSE Connection Resilience Test (Go E2E)",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t04_sse_resilience_e2e.txt"),
    },
    1205: {
        "name": "P12-T05 — WASM Engine Boundary Tests (Go E2E)",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t05_wasm_boundary_e2e.txt"),
    },
    1206: {
        "name": "P12-T06 — 1,000-Node Scale Generator (Go)",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t06_ui_stress_test_e2e.txt"),
    },
    # NOTE: P12-T07 (Telemetry) and P12-T01/T02/T03 (Playwright) are Stitch tasks → see stitch_submit.py
    1208: {
        "name": "P12-T08 — V2.0 Production Cutover (Docker + CI)",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t08_production_cutover.txt"),
    },
    1210: {
        "name": "P12-T10 — VCS-Agnostic Webhook & API Adapter",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t10_vcs_agnostic_adapter.txt"),
    },
    1211: {
        "name": "P12-T11 — Multi-VCS Onboarding UI",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t11_multi_vcs_onboarding_ui.txt"),
    },
    1212: {
        "name": "P12-T12 — System Matrix E2E Test",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t12_system_matrix_e2e.txt"),
    },
    1213: {
        "name": "P12-T13 — Zero-Config Developer Portal",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t13_zero_config_catalog.txt"),
    },
    1209: {
        "name": "P12-T09 — AI Support Copilot",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t09_ai_copilot.txt"),
    },
    1012: {
        "name": "P10-T12 — Tree-sitter Deterministic Impact Analysis",
        "phase": "phase-10-ecosystem",
        "prompt": _load_prompt("prompts/phase-10/t12_tree_sitter.txt"),
    },
    912: {
        "name": "P9-T12 — Embedded SQLite (LibSQL) Local Caching",
        "phase": "phase-9-compliance",
        "prompt": _load_prompt("prompts/phase-9/t12_embedded_sqlite.txt"),
    },
    1301: {
        "name": "P13-T01 — FinOps Cost Prediction",
        "phase": "phase-13-god-mode",
        "prompt": _load_prompt("prompts/phase-13/t01_finops.txt"),
    },
    1503: {
        "name": "P15-T03 — Dependency SLA Tracking",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t03_dependency_sla.txt"),
    },
    1504: {
        "name": "P15-T04 — Schema Smell Detector",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t04_schema_smell.txt"),
    },
    1505: {
        "name": "P15-T05 — AI Incident Post-Mortem Generator",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t05_postmortem_generator.txt"),
    },
    1506: {
        "name": "P15-T06 — Natural Language Governance Rules",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t06_nl_governance.txt"),
    },
    1508: {
        "name": "P15-T08 — Substrate Marketplace",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t08_marketplace.txt"),
    },
    1510: {
        "name": "P15-T10 — Schema Insurance",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t10_schema_insurance.txt"),
    },
    1404: {
        "name": "P14-T04 — Contract Score Badge",
        "phase": "phase-14",
        "prompt": _load_prompt("prompts/phase-14/t04_contract_score_badge.txt"),
    },
    1405: {
        "name": "P14-T05 — Retroactive Dependency Archaeology",
        "phase": "phase-14",
        "prompt": _load_prompt("prompts/phase-14/t05_archaeology.txt"),
    },
    1407: {
        "name": "P14-T07 — Phase 14 E2E Validation",
        "phase": "phase-14",
        "prompt": _load_prompt("prompts/phase-14/t07_e2e_validation.txt"),
    },
    902: {
        "name": "P9-T02 — Continuous AI Sync (watch)",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t02_continuous_ai_sync.txt"),
    },
    903: {
        "name": "P9-T03 — Compliance Mapping",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t03_compliance_mapping.txt"),
    },
    904: {
        "name": "P9-T04 — Quality Gates",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t04_quality_gates.txt"),
    },
    905: {
        "name": "P9-T05 — Hexagonal Architecture & sqlc Refactor",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t05_hexagonal_architecture.txt"),
    },
    906: {
        "name": "P9-T06 — Configuration Management (Viper)",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t06_configuration_management.txt"),
    },
    908: {
        "name": "P9-T08 — Deployment Risk Scoring",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t08_deployment_risk_scoring.txt"),
    },
    909: {
        "name": "P9-T09 — AI Impact Analysis Summaries",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t09_ai_impact_analysis.txt"),
    },
    911: {
        "name": "P9-T11 — Automated Deprecation Campaigns",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t11_automated_deprecation.txt"),
    },
    913: {
        "name": "P9-T13 — Embedded Mermaid Blast Radius",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t13_mermaid_blast_radius.txt"),
    },
    914: {
        "name": "P9-T14 — Ephemeral API Preview URLs",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t14_preview_urls.txt"),
    },
    917: {
        "name": "P9-T17 — Phase 9 E2E Validation",
        "phase": "phase-9",
        "prompt": _load_prompt("prompts/phase-9/t17_e2e_validation.txt"),
    },
    1003: {
        "name": "P10-T03 — AI Mock Data Generator",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t03_ai_mock_data_generator.txt"),
    },
    1004: {
        "name": "P10-T04 — AI Spectral Linter",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t04_ai_spectral_linter.txt"),
    },
    1005: {
        "name": "P10-T05 — Auto-SDK Generator PRs",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t05_auto_sdk_generator.txt"),
    },
    1006: {
        "name": "P10-T06 — Traffic-Aware Pruning (Zombies)",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t06_traffic_aware_pruning.txt"),
    },
    1013: {
        "name": "P10-T13 — Integrated API Documentation Catalog",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t13_api_documentation_catalog.txt"),
    },
    1019: {
        "name": "P10-T19 — CRM/Billing Blast Radius",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t19_crm_billing_blast_radius.txt"),
    },
    1020: {
        "name": "P10-T20 — Phase 10 E2E Validation",
        "phase": "phase-10",
        "prompt": _load_prompt("prompts/phase-10/t20_e2e_validation.txt"),
    },
    1509: {
        "name": "P15-T09 — Substrate Certified Partner Program",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t09_partner_program.txt"),
    },
    1511: {
        "name": "P15-T11 — Substrate for Startups Free Tier",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t11_startups_free_tier.txt"),
    },
    1513: {
        "name": "P15-T13 — Phase 15 E2E Validation",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t13_e2e_validation.txt"),
    },

    1601: {
        "name": "CC-T01 — PASETO Security Migration",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/t01_paseto_migration.txt"),
    },
    1515: {
        "name": "P15-T15 — Air-Gapped License Validator",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t15_license_validator.txt"),
    },

    # ── Missing E2E Coverage (audited 2026-07-23) ─────────────────────────────
    1118: {
        "name": "P11-T18 — Phase 11 Backend E2E Validation",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t18_e2e_validation.txt"),
    },
    1701: {
        "name": "P17-T01 — API Maturity Scorecards (Manager Dashboard)",
        "phase": "phase-17-management",
        "prompt": _load_prompt("prompts/phase-17/t01_maturity_scorecards.txt"),
    },
    1702: {
        "name": "P17-T02 — Legacy API FinOps Translation",
        "phase": "phase-17-management",
        "prompt": _load_prompt("prompts/phase-17/t02_finops_translation.txt"),
    },
    1704: {
        "name": "P17-T04 — PagerDuty Blast Radius Injection",
        "phase": "phase-17-management",
        "prompt": _load_prompt("prompts/phase-17/t04_pagerduty_injection.txt"),
    },
    1705: {
        "name": "P17-T05 — Auto-Rollback via ArgoCD/Flux",
        "phase": "phase-17-management",
        "prompt": _load_prompt("prompts/phase-17/t05_argo_rollback.txt"),
    },
    1799: {
        "name": "P17-T99 — Phase 17 E2E Validation",
        "phase": "phase-17-management",
        "prompt": _load_prompt("prompts/phase-17/t99_e2e_validation.txt"),
    },
    0: {
        "name": "CC-T02 — E2E Overhaul: Full-Stack PGlite Harness",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/e2e/pglite_infrastructure.txt"),
    },
    333: {
        "name": "CC-T03 — Live VCS E2E Integration (Forgejo)",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/t03_live_vcs_forgejo.txt"),
    },
    1: {
        "name": "P3-T12 — Phase 3 Contract Registry E2E Validation (Pipeline Phase 1)",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t12_e2e_validation.txt"),
    },
    2: {
        "name": "P4-T10 — Phase 4 AI Diff Engine E2E Validation (Pipeline Phase 2)",
        "phase": "phase-4-ai",
        "prompt": _load_prompt("prompts/phase-4/t10_e2e_validation.txt"),
    },
    3: {
        "name": "P5-T06 — Phase 5 Discovery Scanners E2E Validation (Pipeline Phase 3)",
        "phase": "phase-5-discovery",
        "prompt": _load_prompt("prompts/phase-5-discovery/t06_e2e_tests.txt"),
    },
    10: {
        "name": "P1-T08 — Phase 1 Core Diff Engine E2E Validation",
        "phase": "phase-1-diff-engine",
        "prompt": _load_prompt("prompts/phase-1-diff-engine/t08_e2e_validation.txt"),
    },
    20: {
        "name": "P2-T05 — Phase 2 GitHub App E2E Validation",
        "phase": "phase-2-github-app",
        "prompt": _load_prompt("prompts/phase-2-github-app/t05_e2e_validation.txt"),
    },
    60: {
        "name": "P6-T10 — Phase 6 QA & Shadow API E2E Validation",
        "phase": "phase-6-qa",
        "prompt": _load_prompt("prompts/phase-6-qa/t10_e2e_tests.txt"),
    },
    70: {
        "name": "P7-T07 — Phase 7 Enterprise E2E Validation",
        "phase": "phase-7-enterprise",
        "prompt": _load_prompt("prompts/phase-7-enterprise/t07_e2e_tests.txt"),
    },
    80: {
        "name": "P8-T10 — Phase 8 Readiness, Authz & Jobs E2E Validation",
        "phase": "phase-8-readiness",
        "prompt": _load_prompt("prompts/phase-8-readiness/t10_e2e_tests.txt"),
    },
    90: {
        "name": "P9-T17 — Phase 9 Compliance E2E Validation",
        "phase": "phase-9-compliance",
        "prompt": _load_prompt("prompts/phase-9-compliance/t17_e2e_validation.txt"),
    },
    100: {
        "name": "P10-T20 — Phase 10 Ecosystem E2E Validation",
        "phase": "phase-10-ecosystem",
        "prompt": _load_prompt("prompts/phase-10/t20_e2e_validation.txt"),
    },
    110: {
        "name": "P11-T18 — Phase 11 Graph UI E2E Validation",
        "phase": "phase-11-ui",
        "prompt": _load_prompt("prompts/phase-11/t18_e2e_validation.txt"),
    },
    120: {
        "name": "P12-T12 — Phase 12 SSE & Scaling E2E Validation",
        "phase": "phase-12-qa",
        "prompt": _load_prompt("prompts/phase-12/t12_system_e2e.txt"),
    },
    130: {
        "name": "P13-T01 — Phase 13 God Mode E2E Validation",
        "phase": "phase-13-god-mode",
        "prompt": _load_prompt("prompts/phase-13/t01_finops_e2e.txt"),
    },
    140: {
        "name": "P14-T07 — Phase 14 Predictive Intelligence E2E Validation",
        "phase": "phase-14",
        "prompt": _load_prompt("prompts/phase-14/t07_e2e_validation.txt"),
    },
    150: {
        "name": "P15-T13 — Phase 15 Monetization & Insurance E2E Validation",
        "phase": "phase-15",
        "prompt": _load_prompt("prompts/phase-15/t13_e2e_validation.txt"),
    },
    990: {
        "name": "P-MCP-01 — Cross-Cutting MCP Parity E2E Validation",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/p_mcp_01_full_parity.txt"),
    },
    312: {
        "name": "P3-T12 — Phase 3 Contract Registry E2E Validation",
        "phase": "phase-3-registry",
        "prompt": _load_prompt("prompts/phase-3-registry/t12_e2e_validation.txt"),
    },
    901: {
        "name": "P-MCP-01 — Full MCP Server Parity + SSE Transport",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/p_mcp_01_full_parity.txt"),
    },
    2000: {
        "name": "E2E-BACKFILL — Backfill Missing Full-Stack E2E Validation Tests",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/e2e_backfill_prompt.txt"),
    },
    2001: {
        "name": "CC-T04 — UI Org Context & API Keys Refactor",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/t04_ui_org_context.txt"),
    },
    2002: {
        "name": "CC-T05 — E2E Validation: UI Org Context",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/t05_ui_org_context_e2e.txt"),
    },
    2003: {
        "name": "CC-T06 — API Keys Backend Integration",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/t06_api_keys_backend.txt"),
    },
    2004: {
        "name": "CC-T07 — E2E Validation: API Keys Backend",
        "phase": "cross-cutting",
        "prompt": _load_prompt("prompts/cross_cutting/t07_api_keys_e2e.txt"),
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
