# Phase 7: Zero-Config Org Rollout (Enterprise Feature)

> **Status:** DRAFT (Backlog)
> **Phase:** Phase 7 (Enterprise Integrations & ITSM)
> **Goal:** Eliminate per-repository onboarding friction, allowing an enterprise with 500+ repositories to adopt Substrate instantly upon GitHub App installation.

---

## 1. Overview & Problem Statement

Currently, Substrate requires a `substrate.yaml` configuration file to exist in the root of a repository. While Task `V1-T01` automates the *creation* of this file via an auto-generated Pull Request, it still requires a human engineer in every single repository to review, approve, and merge that PR. In an enterprise with hundreds or thousands of microservices, this per-repo friction drastically slows down org-wide adoption and time-to-value.

To solve this, Substrate will support **Zero-Config Intelligent Defaults** combined with **Centralized Org-Level Configuration**.

---

## 2. Feature Requirements

### A. Zero-Config Intelligent Defaults
The diff engine must be refactored to execute successfully *without* a `substrate.yaml` file.
- **Auto-Detection:** When the GitHub App receives a Pull Request webhook, the Worker inspects the PR diff. If any modified files match known schema heuristics (e.g., `openapi.yaml`, `schema.graphql`, `*.proto`, `schema.sql`), Substrate automatically analyzes them.
- **Default Behavior:** Substrate applies the strictest default rules (e.g., any breaking change is an error). It posts the standard PR comment and fails the GitHub Status Check.
- **Opt-Out vs Opt-In:** Repositories only need a `substrate.yaml` file if they wish to *override* a breaking change or *exclude* a specific file from being scanned.

### B. Centralized Org-Level Config (The `.github` Repo)
To allow enterprises to define global policies (e.g., "GraphQL breaking changes are warnings, not errors" or "Send all notifications to the #api-governance Slack channel"), Substrate will support a global configuration file.
- **Resolution Logic:** 
  1. The Substrate Worker looks for `substrate.yaml` in the active repository.
  2. If none exists (or if it exists but lacks certain settings), the Worker queries the organization's `.github` repository for a file named `substrate-org.yaml`.
  3. The settings are merged, with local repo settings overriding org-level settings.

### C. Dashboard "Global Enforcement"
- The Svelte Dashboard will include an **Organization Settings** tab (accessible only to users with GitHub Org Admin permissions).
- A toggle labeled **"Enforce Substrate Globally"** will be available.
- When enabled, the Substrate Go API will use the GitHub API to dynamically inject Substrate into the Branch Protection Rules of *all* active repositories in the organization, making it a required status check for merging.

---

## 3. Implementation Steps

1. **Diff Engine Fallback:** Update `engine/pkg/config/parser.go` to return a `DefaultConfig` struct when `substrate.yaml` is missing, rather than throwing a fatal error.
2. **Worker Heuristics:** Update `github-app/src/webhook.ts` to scan PR file payloads. If `filename.includes("openapi")`, route it to the diff engine regardless of config state.
3. **Org Config Fetcher:** Update `github-app/src/github-client.ts` to fetch `.github/substrate-org.yaml` if it exists, and implement a deep-merge utility to combine it with local configs.
4. **Dashboard Admin UI:** Add the "Organization Settings" page to the Svelte dashboard, protected by OAuth Admin scopes.
5. **Branch Protection Script:** Write the Go logic to iterate over all repos in a GitHub Org and apply the required status check API calls.
