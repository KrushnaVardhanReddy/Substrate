import os
import re

phases = [
    (8, "Readiness, Authz & Jobs", "phase8-readiness.md", "phase-8-readiness/t10_e2e_tests.txt"),
    (9, "Compliance & Risk Scoring", "phase9-compliance.md", "phase-9-compliance/t17_e2e_validation.txt"),
    (10, "Ecosystem & Zombie Pruning", "phase10-ecosystem.md", "phase-10/t20_e2e_validation.txt"),
    (11, "Graph UI & Visual Studio", "phase11-ui.md", "phase-11/t18_e2e_validation.txt"),
    (12, "SSE Boundaries & Scaling", "phase12-qa.md", "phase-12/t12_system_e2e.txt"),
    (13, "God Mode & FinOps", "phase13-god-mode.md", "phase-13/t01_finops_e2e.txt"),
    (14, "Predictive Intelligence", "phase14-predictive.md", "phase-14/t07_e2e_validation.txt"),
    (15, "Monetization & Insurance", "phase15-monetization.md", "phase-15/t13_e2e_validation.txt"),
    (99, "MCP Parity & Cross-Cutting", "cross-cutting-mcp.md", "cross_cutting/p_mcp_01_full_parity.txt")
]

spec_template = """# Phase {num} E2E Spec: {title}

## Objective
Validate the Phase {num} {title} features. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/core-repo`
3. Phase-specific mock data.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase{num}_api_test.go`)
- Assert standard API behaviors and validation logic.

### 2. Playwright UI Tests (`dashboard/playwright/phase{num}.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/`.
- Assert the UI renders the correct state.
"""

prompt_template = """TASK: P{num} — Phase {num} {title} Full-Stack E2E Validation

═══════════════════════════════════════════════════════════════
OBJECTIVE
═══════════════════════════════════════════════════════════════
Using the new Full-Stack PGlite Test Harness (CC-T02), implement the complete E2E 
test suite for Phase {num} ({title}). 

You must write BOTH the Go API tests AND the Playwright UI tests.

═══════════════════════════════════════════════════════════════
CRITICAL READING — DO THIS BEFORE WRITING A SINGLE LINE OF CODE
═══════════════════════════════════════════════════════════════
1. docs/specs/e2e/{spec_file}
   → Read the exact API assertions and UI actions required.

2. seed_mcp.sql
   → Ensure the data is seeded properly for this phase.

═══════════════════════════════════════════════════════════════
DELIVERABLES
═══════════════════════════════════════════════════════════════
5. FILE: scripts/e2e/phase{num}_api_test.go (NEW)
   - API assertions.

6. FILE: dashboard/tests/e2e/phase{num}.spec.ts (NEW)
   - UI assertions.

═══════════════════════════════════════════════════════════════
🚨 CRITICAL RULES FOR THIS RUN 🚨
═══════════════════════════════════════════════════════════════
1. NO INTERNAL PACKAGE MOCKING:
   - Your Go tests in `scripts/e2e/` are in a separate module. You CANNOT import `github.com/KrushnaVardhanReddy/Substrate/api/internal/...` to directly call handlers.
   - You MUST make external HTTP requests (`http.NewRequest`) to the live API running at `http://localhost:8090`.

2. DATABASE & PGLITE:
   - PGlite runs on port `54320`.
   - The connection string MUST be: `postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true`
   - You MUST clean up and seed your own data in your test file by instantiating `pgxpool.New` with the above URL.

3. CYTOSCAPE PLAYWRIGHT TESTS:
   - Cytoscape is an HTML canvas. You CANNOT use standard DOM locators for graph nodes.
   - You MUST interact with the graph using `page.evaluate(() => window.cyInstance...)` and wait for it to load using `waitForFunction`.

4. VENDOR PRESERVATION:
   - Do NOT modify or delete anything inside `scripts/e2e/vendor/pglite/`.
   - If you need to add Go dependencies to `scripts/e2e`, run `go mod vendor` inside `scripts/e2e/`.

5. RIVER QUEUE CRASHES:
   - PGlite is incompatible with River Queue. The server runs with `SKIP_RIVER="true"`. Do not attempt to start or test River background jobs natively on PGlite.

Commit message: "test: P{num} add Phase {num} full-stack E2E tests"
Target branch: feature/dev
"""

for num, title, spec_file, prompt_file in phases:
    # Create Spec
    spec_path = os.path.join("docs/specs/e2e", spec_file)
    with open(spec_path, "w") as f:
        f.write(spec_template.format(num=num, title=title))
    
    # Create Prompt
    prompt_path = os.path.join("prompts", prompt_file)
    os.makedirs(os.path.dirname(prompt_path), exist_ok=True)
    with open(prompt_path, "w") as f:
        f.write(prompt_template.format(num=num, title=title, spec_file=spec_file))

print("Generated all missing Specs and Prompts.")
