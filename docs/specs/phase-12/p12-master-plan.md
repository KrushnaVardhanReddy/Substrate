# Phase 12: V2.0 Public Launch & Quality Assurance

## Objective
Harden the Substrate V2.0 platform for a public launch by implementing rigorous end-to-end (E2E) UI testing using Playwright, stress-testing the Go backend and WebAssembly bridge, and finalizing telemetry and deployment pipelines.

## Wave 1: Frontend UI Reliability (Playwright)
The SvelteKit frontend is highly interactive. We must ensure no UI regressions occur during refactors.

*   **P12-T01 — Zero-to-One Onboarding E2E**
    *   **Spec:** Automate the entire user journey. Simulate a new user entering a mock GitHub token, clicking "Connect," waiting for the "Scanning Repositories" progress bar to finish, and successfully redirecting to the dynamic `/org/{PUBLIC_ORG_NAME}/graph` route (or `/org/demo/graph` fallback).
    *   **Success:** Playwright trace shows the graph view loads successfully with the sidebar visible.

*   **P12-T02 — Cytoscape Interaction Tests**
    *   **Spec:** Load the graph. Simulate clicking a node to ensure the "Cascading Blast Radius" triggers (assert CSS class changes). Toggle the "Volatility Heatmap" switch and assert node colors change. Hover over an edge and assert the Tooltip is visible in the DOM.
    *   **Success:** Complex canvas interactions do not throw Svelte `$state` exceptions.

*   **P12-T03 — AI Playground & Diff Viewer E2E**
    *   **Spec:** Verify the AI Playground (formerly API Studio). Type malformed schema into the playground editor, stream the AI analysis from the backend, and assert the UI displays the analysis findings and applies the auto-fix gracefully without crashing.

## Wave 2: Backend Resilience & Scale (Go Testing)
The backend must handle high concurrency and massive payloads.

*   **P12-T04 — SSE Connection Resilience Test**
    *   **Spec:** Write Go tests to simulate 100+ concurrent clients connecting to the `GET /api/v1/events` endpoint. Abruptly cancel their contexts and ensure the Server-Sent Events channel registry unregisters them without leaking memory or deadlocking.

*   **P12-T05 — WASM Engine Boundary Tests**
    *   **Spec:** Automated JS-to-WASM bridge tests. Ensure that passing massively malformed strings (e.g., 50MB of garbage data) from the Svelte frontend into the Go WASM engine returns a safe JS error payload instead of a fatal WebAssembly panic.

*   **P12-T06 — 200-Node Stress Test**
    *   **Spec:** Generate a massive mock dependency graph (200 microservices, 600 strictly acyclic edges) and load it into the Cytoscape canvas. (Note: Edges must be strictly directed without cycles to prevent the Dagre layout engine from infinite looping).
    *   **Success:** Ensure the Dagre layout algorithm computes in under 2 seconds and the UI runs at 60fps without browser lockup.

## Wave 3: Launch Readiness
Final preparations before directing public traffic to the V2.0 dashboard.

*   **P12-T07 — Telemetry & Crash Reporting**
    *   **Spec:** Integrate a lightweight privacy-first telemetry system (e.g., PostHog or Sentry) so we know immediately if the Cytoscape canvas crashes in production for a real user.

*   **P12-T08 — V2.0 Production Cutover**
    *   **Spec:** Final updates to the Docker/Helm CI/CD pipelines to ensure the WASM binaries and Svelte static assets are perfectly bundled into the single-binary deployment.

*   **P12-T09 — Local Git Server Integration (Forgejo)**
    *   **Spec:** Deploy a local Forgejo container to simulate real GitHub webhooks for the Cloudflare worker and Go API, moving away from static mock payloads.

*   **P12-T10 — VCS-Agnostic Webhook & API Adapter**
    *   **Spec:** Refactor the Cloudflare worker to implement an Adapter pattern, decoupling it from GitHub. Support generic webhook parsing and generic API clients for GitHub, GitLab, and Gitea/Forgejo to unlock enterprise self-hosted environments.



## Execution Plan
This phase will heavily utilize Jules for both Playwright automation (TypeScript) and backend stress testing (Go). Since these are test suites, we will instruct Jules to use the `--ui` mode for Playwright debugging if necessary and provide strict test artifacts.
