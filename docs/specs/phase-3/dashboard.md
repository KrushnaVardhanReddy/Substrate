# Phase 3: SvelteKit Dashboard & Design System

## Overview

The Substrate Dashboard visualises the cross-repository contract intelligence data stored in the Registry API (built in P3-T01). It allows engineers to see exactly what downstream consumers will break if they merge a PR.

**Status:** ⏳ READY
**Owner:** Jules (Logic) + Stitch (UI Mockups)
**Blocked by:** Nothing. P3-T01 API is complete.

---

## 1. Stack and Architecture

- **Framework:** SvelteKit (Static adapter)
- **Language:** TypeScript
- **Styling:** Vanilla CSS (no Tailwind, per user preference)
- **Design System:** Substrate Dark UI
- **Location:** `dashboard/` (root directory alongside `engine/` and `github-app/`)

---

## 2. Design System Tokens (Stitch Requirements)

All Stitch HTML/CSS generation must adhere strictly to these tokens.

### Colors
- **Background Base:** `#090A0F` (Deep space black/blue)
- **Background Surface:** `#14161F` (Elevated cards)
- **Background Hover:** `#1D202D`
- **Accent Primary:** `#00BFA5` (Substrate Teal - signifies "Safe")
- **Accent Breaking:** `#FF5252` (Substrate Red - signifies "Breaking")
- **Accent Warning:** `#FFAB40` (Substrate Orange)
- **Text Primary:** `#F8F9FA`
- **Text Secondary:** `#9BA1B0`
- **Border Subtle:** `#2A2E3D`

### Typography
- **Font Family:** `Inter`, `system-ui`, `sans-serif`
- **Code Font:** `JetBrains Mono`, `monospace` (for spec file paths and rule IDs)

### Micro-interactions
- **Hover States:** All interactive elements must have a subtle background shift (e.g., to Background Hover) and a `0.2s ease` transition.
- **Borders:** Panels should use subtle 1px borders using `Border Subtle`.

---

## 3. Core Layout Structure (P3-T03)

The initial dashboard implementation consists of a single global layout wrapper:

### Sidebar (Left, Fixed Width 280px)
- **Header:** Substrate Logo / Title.
- **Section 1: Repositories:** A scrollable list of repositories registered in the organisation (e.g., `myorg/backend-api`).
- **Section 2: Settings:** API Tokens, Organization config.

### Top Navigation (Top, Fixed Height 64px)
- **Breadcrumbs:** e.g., `myorg/backend-api > Dependencies`.
- **User Profile:** Avatar and org switcher.

### Main Content Area (Dynamic)
- The central view where data visualisations and tables will be rendered.
- **Padding:** 32px standard padding.

---

## 4. API Integration (Jules Requirements)

The dashboard will fetch data from the Go API Server built in P3-T01.

| Component | Endpoint | Method | Response Payload |
|---|---|---|---|
| **Sidebar Repos** | `/api/v1/repos/:org` | `GET` | `[{ id, name, full_name }]` |
| **Dependency Graph** | `/api/v1/graph/:org` | `GET` | `[{ provider, consumer, status }]` |

*Note: All API calls require the `Authorization: Bearer <TOKEN>` header.*

---

## 5. Development Pipeline

1. **Stitch Mockup:** Generate the static HTML/CSS for the Core Layout (Sidebar + Nav + Content Area) using the Design System Tokens.
2. **SvelteKit Init:** Run `npm create svelte@latest dashboard` to scaffold the project.
3. **Jules Porting:** Port the Stitch HTML/CSS into Svelte `+layout.svelte` and `+page.svelte` files.
4. **API Wiring:** Connect the Sidebar to the `/api/v1/repos/:org` endpoint.

---

## 6. P3-T14: Interactive Diff Viewer URL (GitHub Worker → Dashboard)

**Status:** 🔄 IN PROGRESS (Jules task P3-T14)
**Dependency:** P3-T03 ✅

### Overview
Every Substrate PR comment currently ends with:
```
*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*
```

This task adds a **"View in Dashboard →"** link directly in the PR comment footer so engineers can click through to the live dependency graph and diff report on the Substrate Dashboard.

### How it works

The GitHub Worker Cloudflare environment already has a `DASHBOARD_URL` binding (to be added). When the formatter builds the PR comment, it appends a dashboard deep-link to every comment footer.

The deep-link format:
```
https://<DASHBOARD_URL>/diff?owner=<owner>&repo=<repo>&pr=<prNumber>
```

Example:
```
https://substrate.mycompany.com/diff?owner=myorg&repo=backend-api&pr=45
```

If `DASHBOARD_URL` is not set in the Worker's environment, the link is **omitted silently** — the comment still renders correctly, just without the dashboard link. This ensures zero regressions in existing deployments.

### Changes required

**1. `github-app/src/types.ts`**
Add `DASHBOARD_URL?: string` to the `Env` interface (optional — existing deployments without the variable still work).

**2. `github-app/src/formatter.ts`**
- Update `formatPRComment` signature to accept an optional `dashboardUrl?: string` parameter.
- Replace the `*Powered by Substrate*` footer line in all three branches (breaking, warning, all-clear) with a new `formatFooter(dashboardUrl?, owner?, repo?, prNumber?)` helper that builds the correct footer.
- If `dashboardUrl` is provided: footer becomes:
  ```
  [View in Dashboard →](<dashboardUrl>/diff?owner=<owner>&repo=<repo>&pr=<prNumber>)
  *Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*
  ```
- If `dashboardUrl` is NOT provided: footer is just the existing powered-by line (no regression).

**3. `github-app/src/index.ts`**
- When calling `formatPRComment(report, config)`, also pass `env.DASHBOARD_URL`, `event.owner`, `event.repo`, and `event.prNumber`.

**4. `github-app/test/index.test.ts` and `github-app/test/formatter.test.ts`**
- Add test cases asserting the dashboard link appears when `DASHBOARD_URL` is set.
- Add test cases asserting the comment renders correctly when `DASHBOARD_URL` is absent.

---

## 7. P3-T05: Dependency Graph Visualization

**Status:** 🔄 IN PROGRESS (Antigravity task)
**Dependency:** P3-T04 ✅

### Overview
Visualise the upstream/downstream contract dependencies for an organization. This allows developers to see what schemas exist and what downstream consumers depend on them.

### Component Architecture
- **Route:** `dashboard/src/routes/org/[org]/graph/+page.svelte`
- **Load Function (`+page.ts`):** 
  - Fetch graph data from `GET /api/v1/graph/[org]`.
  - Transform the linear `{ provider, consumer, status }` edge list into a hierarchical nodes/edges structure suitable for rendering.
- **Main Canvas (`<main class="main-canvas">`):**
  - Render an SVG layer containing connection paths (bezier curves from upstream to downstream nodes).
  - Render a Node layer using absolute positioning (or a CSS grid/flexbox based pseudo-hierarchy). For v1, simple 3-column layout (Upstream, Central, Downstream) can be used.
- **Node Card (`.node-card`):**
  - Displays Repository Name, Version/Commit, and Status (Safe/Breaking).
  - Selectable (updates a bound `selectedNode` variable).
- **Detail Panel (`<aside class="detail-panel">`):**
  - Rendered conditionally when a node is selected.
  - Displays the raw schema or metadata, and a list of direct consumers or providers.

---

## 8. P3-T11: Compatibility Matrix Dashboard

**Status:** 🔄 READY (Antigravity task)
**Dependency:** P3-T05 ✅

### Overview
Inspired by PactFlow's compatibility matrix, this is the central enterprise control panel. It visualizes which version of a provider (e.g., `users-api`) is compatible with which version of its downstream consumers (e.g., `frontend`, `mobile-app`).

### Component Architecture
- **Route:** `dashboard/src/routes/org/[org]/matrix/+page.svelte`
- **Load Function (`+page.ts`):**
  - Fetch matrix data from the API (or mock it initially). Data structure will represent a grid of Provider Versions vs Consumer Versions, with cell states (`COMPATIBLE`, `INCOMPATIBLE`, `UNKNOWN`).
- **Matrix UI (`.matrix-container`):**
  - A dynamic data table/grid.
  - **Y-Axis (Rows):** Provider versions (e.g., `users-api v1.2.0`).
  - **X-Axis (Columns):** Consumer environments/versions (e.g., `frontend PROD`, `frontend DEV`).
  - **Cells:** Render a visual indicator:
    - ✅ Green Tick for `COMPATIBLE` (safe schema match).
    - ❌ Red Cross for `INCOMPATIBLE` (breaking change detected).
    - ➖ Gray Dash for `UNKNOWN` (not tested/irrelevant).
- **Interactions:**
  - Clicking a cell opens a modal or side panel detailing the specific check (the breaking change that occurred, or the timestamp of the successful verification).
  - Filters to select which Provider Repo to view the matrix for.

### Design
- Use `app.css` and the design tokens (Dark mode, `#0f172a` bg, green/red status colors).
- Ensure the table is horizontally scrollable for many consumers.

---

## 9. P3-T15: Dashboard E2E Testing (Playwright)

**Status:** 🔄 READY (Jules task)
**Dependency:** P3-T11 ✅

### Overview
To ensure the Phase 3 Enterprise Dashboard remains stable across releases, we require automated End-to-End (E2E) testing for the frontend UI. Playwright is the chosen framework to simulate user interactions within the SvelteKit dashboard.

### Testing Strategy
- **Isolation via Network Mocking:** The frontend tests must run independently of the Go backend. All API calls (e.g., `/api/v1/repos/*`, `/api/v1/graph/*`, `/api/v1/matrix/*`) must be intercepted and mocked using Playwright's `page.route()`.
- **Setup:** Configure `playwright.config.ts` to spin up the SvelteKit dev server (`npm run dev`) before executing tests.

### Required Scenarios (`dashboard/tests/e2e/`)
1. **Navigation Flow (`navigation.spec.ts`)**:
   - Verify that clicking sidebar items correctly routes to `/org/[org]/graph` and `/org/[org]/matrix`.
   - Verify the active class is correctly applied to the navigation items.
2. **Dependency Graph (`graph.spec.ts`)**:
   - Mock a successful graph data response.
   - Assert that node cards (e.g., `core/auth`) render on the screen.
   - Assert that clicking a node opens the `<aside class="detail-panel">` and displays the correct metadata.
3. **Compatibility Matrix (`matrix.spec.ts`)**:
   - Mock a successful matrix data response with `COMPATIBLE` and `INCOMPATIBLE` states.
   - Assert that the matrix grid renders correctly.
   - Assert that clicking an `INCOMPATIBLE` cell opens the side panel displaying the "Breaking Change Detected" alert.
