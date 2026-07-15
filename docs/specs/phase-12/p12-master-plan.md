# Phase 12: V2.0 Public Launch & Quality Assurance

## Objective
Harden the Substrate V2.0 platform for a public launch by implementing rigorous end-to-end (E2E) UI testing using Playwright, stress-testing the Go backend and WebAssembly bridge, and finalizing telemetry and deployment pipelines.

## Wave 1: Frontend UI Reliability (Playwright)
The SvelteKit frontend is highly interactive. We must ensure no UI regressions occur during refactors.

*   **P12-T01 — Zero-to-One Onboarding E2E**
    *   **Spec:** Automate the entire user journey. Simulate a new user entering a mock GitHub token, clicking "Connect," waiting for the "Scanning Repositories" progress bar to finish, and successfully redirecting to the `/org/demo/graph` route.
    *   **Success:** Playwright trace shows the graph view loads successfully with the sidebar visible.

*   **P12-T02 — Svelte Flow Interaction Tests**
    *   **Spec:** Load the graph. Simulate clicking a node to ensure the "Cascading Blast Radius" triggers (assert CSS class changes). Toggle the "Volatility Heatmap" switch and assert node colors change. Hover over an edge and assert the Tooltip is visible in the DOM.
    *   **Success:** Complex canvas interactions do not throw Svelte `$state` exceptions.

*   **P12-T03 — Visual API Studio & Diff Viewer E2E**
    *   **Spec:** Verify the bidirectional binding. Type invalid YAML into the studio textarea and assert the UI displays an error boundary gracefully without crashing the entire Svelte application.

## Wave 2: Backend Resilience & Scale (Go Testing)
The backend must handle high concurrency and massive payloads.

*   **P12-T04 — SSE Connection Resilience Test**
    *   **Spec:** Write Go tests to simulate 100+ concurrent clients connecting to the `GET /api/v1/events` endpoint. Abruptly cancel their contexts and ensure the Server-Sent Events channel registry unregisters them without leaking memory or deadlocking.

*   **P12-T05 — WASM Engine Boundary Tests**
    *   **Spec:** Automated JS-to-WASM bridge tests. Ensure that passing massively malformed strings (e.g., 50MB of garbage data) from the Svelte frontend into the Go WASM engine returns a safe JS error payload instead of a fatal WebAssembly panic.

*   **P12-T06 — 1,000-Node Stress Test**
    *   **Spec:** Generate a massive mock dependency graph (1,000 microservices, 3,000 edges) and load it into the Svelte Flow canvas.
    *   **Success:** Ensure the Dagre layout algorithm computes in under 2 seconds and the UI runs at 60fps without browser lockup.

## Wave 3: Launch Readiness
Final preparations before directing public traffic to the V2.0 dashboard.

*   **P12-T07 — Telemetry & Crash Reporting**
    *   **Spec:** Integrate a lightweight privacy-first telemetry system (e.g., PostHog or Sentry) so we know immediately if the Svelte Flow canvas crashes in production for a real user.

*   **P12-T08 — V2.0 Production Cutover**
    *   **Spec:** Final updates to the Docker/Helm CI/CD pipelines to ensure the WASM binaries and Svelte static assets are perfectly bundled into the single-binary deployment.

## Execution Plan
This phase will heavily utilize Jules for both Playwright automation (TypeScript) and backend stress testing (Go). Since these are test suites, we will instruct Jules to use the `--ui` mode for Playwright debugging if necessary and provide strict test artifacts.
