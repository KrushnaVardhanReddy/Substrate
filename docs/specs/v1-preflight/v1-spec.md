# V1.0 Pre-Flight Checklist Specification

## Overview
Before pushing Substrate to the GitHub Marketplace for general enterprise availability (V1.0.0), we must resolve crucial developer experience (DX), onboarding, and deployment safety tasks. This specification defines the exact requirements for Tasks V1-T01 through V1-T06.

---

## V1-T01: GitHub App Auto-Discovery (Zero-Touch Onboarding)
**Goal:** Eliminate manual repository configuration so organizations can adopt Substrate instantly.
**Implementation:**
- Extend the Cloudflare Worker GitHub App integration.
- Listen for the `installation_repositories` webhook event (when the app is installed or a repo is added).
- For each new repository, automatically trigger a Github API scan for `openapi.yaml`, `schema.graphql`, or `.env`.
- If an API contract is found, the Worker automatically creates a new branch (`substrate-init`) and opens a Pull Request injecting a default `substrate.yaml` file into the repository root.

## V1-T02: CLI AI Architect (`substrate init --design`)
**Goal:** Provide an LLM-powered interactive CLI that helps backend engineers design OpenAPI contracts before they write a line of code.
**Implementation:**
- Add `--design` flag to `substrate init`.
- Implement a conversational loop using the `github.com/sashabaranov/go-openai` library (or similar generic LLM provider wrapper).
- Prompt the user for their desired endpoints (e.g., "I need a blog API with posts and comments").
- Stream the LLM response to generate a valid OpenAPI 3.0 YAML spec.
- Automatically save the generated spec to the local filesystem and create the corresponding `substrate.yaml`.

## V1-T03: Interactive Diff Viewer UI (Vercel-style)
**Goal:** Replace walls of text in PR comments with a highly visual, actionable UI.
**Implementation:**
- Extend the Svelte Dashboard (`dashboard/src/routes/diff/[id]`).
- When a PR diff is evaluated, generate a unique `DiffID` and store the JSON output in the PostgreSQL database.
- Append a Preview URL (`https://substrate.run/diff/<DiffID>`) to the top of the GitHub PR comment.
- The UI page must render a split-pane, side-by-side visual comparison (Red/Green highlighting) using the `DiffReport` JSON, showing exactly which schema lines broke.

## V1-T04: Deployment Safety Gate (`substrate check-deploy`)
**Goal:** Prevent bad deployments where a consumer microservice is deployed *before* its required provider.
**Implementation:**
- Add `check-deploy` command to the Substrate CLI.
- Execution context: Runs in the CI pipeline *right before* the CD step (e.g., `kubectl apply`).
- It hashes the current local repo state and pings the Registry API: `/api/v1/registry/can-deploy?repo=frontend&commit=abc`.
- The Registry verifies all outbound dependencies of `frontend` at commit `abc`. If `frontend` depends on `backend-api` v2.0, but the Registry shows `backend-api` is still at v1.0 in production, it returns `HTTP 409 Conflict`.
- The CLI exits with a non-zero code, failing the CI/CD pipeline and preventing a production outage.

## V1-T05: Local Validation CLI (`substrate validate`)
**Goal:** Allow developers to test schema breaks locally, completely offline or against a remote registry, before pushing commits.
**Implementation:**
- Complete the `validate` command stub in `engine/cmd/substrate/main.go`.
- Allow the CLI to execute `oasdiff` strictly locally against a `base` file (e.g., main branch checkout) without needing an HTTP server.
- Output clean terminal text formatting (Red/Green terminal codes).

## V1-T06: Legal & Licensing Audit
**Goal:** Ensure enterprise procurement teams approve the binary by guaranteeing license compliance.
**Implementation:**
- Execute `go-licenses` across the `engine` and `api` binaries.
- Ensure no `GPL` or `AGPL` transitive dependencies exist.
- Generate an `ATTRIBUTIONS.md` file and embed a "Credits / OSS Licenses" modal in the Svelte dashboard footer.

## V1-T07: V1.0 System E2E Tests
**Goal:** Prove the final V1.0 pipeline—from automated onboarding, AI design, schema diffing, to final deployment gates—works as a seamless pipeline.
**Implementation:**
- Write `scripts/e2e/v1_e2e_test.go`.
- **Step 1:** Simulate the GitHub App Webhook receiving a new repository installation, asserting it triggers the Auto-Discovery PR (V1-T01).
- **Step 2:** Simulate a developer invoking the CLI AI Architect (V1-T02) using a mocked LLM interface, asserting the generated YAML is valid.
- **Step 3:** Simulate a PR schema break and assert that the interactive Diff Viewer UI JSON is correctly generated and stored (V1-T03).
- **Step 4:** Execute the `substrate validate` (V1-T05) locally to fix the break.
- **Step 5:** Execute `substrate check-deploy` (V1-T04) and assert it blocks a premature consumer deployment.
