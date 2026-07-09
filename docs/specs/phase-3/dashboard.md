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
