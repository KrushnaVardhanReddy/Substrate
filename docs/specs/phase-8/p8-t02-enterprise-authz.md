# Spec: P8-T02 - Enterprise Authz (Casbin/OpenFGA)

## 1. Overview
In Phase 8, Substrate transitions from a single-tenant mental model to a full multi-tenant enterprise architecture. Currently, any logged-in user can view any repository if they know the organization name. We must implement strict Role-Based Access Control (RBAC) to restrict viewing and mutating data (webhooks, overriding rules, viewing reports) based on the user's GitHub permissions.

## 2. Requirements

### 2.1 Access Model
- **Roles:** `Admin`, `Member`, `Read-Only`
- Use GitHub's installation permissions to derive default RBAC roles. If a user is an Organization Admin in GitHub, they are a Substrate Admin for that org.

### 2.2 Enforcing Authz in Go API
- Integrate `casbin/casbin` or implement a fast middleware-based RBAC checker.
- Create an `AuthzMiddleware` that intercepts requests to `/api/v1/org/{org}/*` and verifies the user's JWT has the appropriate permissions for the requested organization.

### 2.3 Dashboard UI
- If a user lacks `Admin` privileges, hide the "Settings" and "Webhook" tabs on the Svelte dashboard.
- If a user lacks permission entirely, display a 403 Forbidden screen gracefully.

## 3. Implementation Steps
1. Add `casbin` or standard auth middleware logic.
2. Update the GitHub Login callback handler to fetch and cache user org memberships into the Substrate DB (or securely inside the JWT).
3. Apply `AuthzMiddleware` to all sensitive routes in `router.go`.
