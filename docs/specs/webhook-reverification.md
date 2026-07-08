# Substrate — Webhook Re-verification Specification

> **Status:** DRAFT
> **Phase:** Phase 2 / 3 (Requires Phase 3 Contract Registry)
> **Scope:** Defines the automated drift detection loop when a consumer's specification changes.

---

## Problem Statement

When a provider opens a PR, Substrate checks it against all consumers registered in the Substrate Contract Registry. It determines compatibility based on the consumer's *currently registered snapshot*.

However, what happens when a **consumer** changes their code and opens a PR? 
If a consumer (e.g., `frontend`) opens a PR that starts requesting a new field from `users-api` that doesn't exist yet, it's the *consumer* who is introducing a breaking drift.

**Requirement:** Substrate must monitor consumer PRs. When a consumer modifies their required spec (e.g., a GraphQL query or OpenAPI client definition), Substrate must re-verify that the consumer's proposed change is compatible with the provider's *current production schema*.

---

## Webhook Execution Flow

### 1. Consumer PR Opened
A developer opens a PR in `frontend` modifying `queries.graphql`.

### 2. Substrate GitHub App Triggered
The Cloudflare Worker receives the `pull_request` webhook for `frontend`.

### 3. Registry Lookup
1. The Worker parses the `substrate.yaml` in the `frontend` PR.
2. The config defines `frontend` as a **consumer** of `users-api`:
```yaml
consumers:
  - name: "users-api"
    type: graphql
    source: "github.com/myorg/users-api"
    path: "schema.graphql"
    branch: "main"
```

### 4. Reverse Diff Computation
Instead of diffing the provider against the consumer, Substrate diffs the **Consumer's PR** against the **Provider's `main` branch**.
1. Substrate fetches the proposed `queries.graphql` from the `frontend` PR.
2. Substrate fetches the latest `schema.graphql` from the `users-api` `main` branch (from the registry).
3. The diff engine runs a compatibility check: "Are the queries in `frontend (PR)` valid against `users-api (main)`?"

### 5. PR Feedback (Consumer Drift Detected)
If the consumer is asking for something the provider doesn't have, Substrate blocks the consumer's PR.

```markdown
## 🔴 Substrate — Drift Detected

This PR introduces queries that are **incompatible** with the provider's current production schema.
If merged, this code will fail in production because the provider (`users-api`) does not support these requirements yet.

| Provider | Missing Element |
|---|---|
| `users-api` | Field `User.phoneNumber` does not exist |

**Resolution:**
The provider must merge and deploy support for `User.phoneNumber` before this PR can be merged.
```

---

## The "Lock-Step" Coordination Pattern

This re-verification creates a perfect lock-step safety mechanism between teams:
1. **Consumer** tries to merge early -> Blocked by Substrate (Provider isn't ready).
2. **Provider** merges support for the new field.
3. Substrate automatically re-runs the status check on the **Consumer PR** -> Status turns Green ✅.
4. Consumer merges.

This eliminates runtime errors caused by mismatched deployment orders in microservice environments.
