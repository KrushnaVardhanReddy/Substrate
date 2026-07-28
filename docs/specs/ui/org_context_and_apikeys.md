# Substrate UI Org Context & API Keys Refactor

## Objective
Fix navigation context loss when accessing the Marketplace and AI Playground, and implement the scaffolding for organization-scoped API Keys.

## Problem Statement
Currently, the `/marketplace` and `/playground` routes are global in the SvelteKit frontend. When a user navigates to them from an organization context (e.g., `/org/mcp-org`), the org context is lost. The sidebar correctly defaults to global navigation, leaving the user with no intuitive way to return to their organization. Furthermore, the "API Keys" sidebar item is a dead div with no functionality.

## Solution Specification

### 1. Route Restructuring
Move the global routes into the organization namespace:
- `dashboard/src/routes/(app)/marketplace` -> `dashboard/src/routes/(app)/org/[org]/marketplace`
- `dashboard/src/routes/(app)/playground` -> `dashboard/src/routes/(app)/org/[org]/playground`
- Create `dashboard/src/routes/(app)/org/[org]/apikeys`

### 2. Sidebar Updates
Update `dashboard/src/lib/components/Sidebar.svelte`:
- Replace global `href="/marketplace"` with `href="/org/{org}/marketplace"`
- Replace global `href="/playground"` with `href="/org/{org}/playground"`
- Replace `<div class="nav-item">API Keys</div>` with `<a href="/org/{org}/apikeys" class="nav-item {pathname === '/org/' + org + '/apikeys' ? 'active' : ''}">API Keys</a>`

### 3. API Keys Scaffolding
In `dashboard/src/routes/(app)/org/[org]/apikeys/+page.svelte`, implement a basic placeholder UI:
- Page Header: "API Keys"
- Empty State or basic Table for displaying tokens (Name, Prefix, Created At).
- A "Generate New Key" button.
*(Note: Full backend integration for API keys will be handled in a separate phase; this task focuses purely on fixing the UI navigation and layout scaffolding).*
