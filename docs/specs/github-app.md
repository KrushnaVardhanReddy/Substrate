# Substrate — GitHub App Architecture Specification

> **Status:** APPROVED ✅
> **Spec-First Gate:** No Phase 2 implementation begins until this document is approved.
> **Scope:** Defines the complete architecture for the Substrate GitHub App — the webhook receiver, binary execution model, PR comment format, and configuration contract.

---

## Problem Statement

Phase 1 ships Substrate as a GitHub Action. Users must manually add the workflow YAML to every repo. This creates per-repo friction and limits adoption to developers who already know to look for it.

**Requirement:** External developers and organizations should be able to install Substrate once (at the repo or org level) and get automatic breaking change protection on every PR — with zero per-repo workflow setup.

---

## Solution: GitHub App + Cloudflare Worker + Container Service

```
GitHub PR Opened
      ↓  HTTPS POST (HMAC-SHA256 signed)
Cloudflare Worker          ← thin orchestration layer
      ↓  Fetches base + head spec files via GitHub API
      ↓  HTTP POST {base, head, config}
Cloudflare Container       ← substrate-engine Docker image (CGO enabled)
      ↓  Runs: substrate diff <base> <head>
      ↓  Returns: DiffReport JSON
Cloudflare Worker          ← result handler
      ↓  POST /repos/.../issues/.../comments   (PR comment)
      ↓  POST /repos/.../statuses/{sha}        (commit status check)
GitHub PR                  ← shows ✅ Pass / ❌ Fail
```

### Why Option C (Container Service) over Wasm or Action Trigger?

| Option | CGO/SQL? | Latency | Infrastructure | Verdict |
|---|---|---|---|---|
| A: Go → Wasm in Worker | ❌ Broken (CGO) | Fast | None | ❌ Blocked by SQL parser |
| B: Trigger GitHub Action | ✅ Works | ~60s (Action queue) | None extra | ❌ Too slow for PR UX |
| C: Container Service | ✅ Works | ~2-5s | One container service | ✅ **Selected** |

Option C is the cleanest architecture. The Worker stays thin (no compute). The container runs the full binary with CGO support (SQL parser works). Latency is 2–5 seconds — fast enough for a responsive PR experience.

---

## Component 1: Cloudflare Worker (Webhook Receiver)

**File:** `app/worker/src/index.ts`

**Responsibilities:**
1. Receive `pull_request` webhook events from GitHub
2. Validate the HMAC-SHA256 webhook signature (`X-Hub-Signature-256` header)
3. Fetch the base and head spec files from GitHub API using the installation access token
4. POST spec content to the Container Service
5. Parse the returned `DiffReport` JSON
6. Post PR comment and set commit status check

**Webhook events handled:**
- `pull_request.opened` — run full diff, post initial comment
- `pull_request.synchronize` — re-run diff on new commits, update comment
- `pull_request.reopened` — re-run diff

**Webhook events ignored:**
- `pull_request.closed`, `pull_request.labeled`, all other event types

**HMAC validation (mandatory):**
```typescript
const signature = request.headers.get('X-Hub-Signature-256');
const body = await request.text();
const expectedSig = 'sha256=' + hmac('sha256', WEBHOOK_SECRET, body);
if (!timingSafeEqual(signature, expectedSig)) {
  return new Response('Unauthorized', { status: 401 });
}
```

**Spec file fetching:**
- Read `substrate.yaml` from the PR's head commit to get `base_schema` and `head_schema` paths
- Fetch `base_schema` file content from the PR's base branch via GitHub Contents API
- Fetch `head_schema` file content from the PR's head commit via GitHub Contents API
- If no `substrate.yaml` exists on the head commit → trigger **Missing Config flow** (see below)

---

## Component 2: Container Service (substrate-engine)

**Docker Image:** `kpakkiragari/substrate-engine:latest` (the same image we already publish in Phase 1)

**Hosting:** Cloudflare Containers (preferred — Cloudflare-native, bound directly to the Worker) with Fly.io as a fallback option if Cloudflare Containers is not yet stable enough.

**Contract — Request (Worker → Container):**
```json
POST /diff
Content-Type: application/json

{
  "base_schema": "<content of base spec file>",
  "head_schema": "<content of head spec file>",
  "config": "<content of substrate.yaml>",
  "schema_type": "openapi"   // "openapi" | "sql" | "graphql"
}
```

**Contract — Response (Container → Worker):**
```json
HTTP 200 OK
Content-Type: application/json

{
  "breaking": [...],
  "warning": [...],
  "info": [...],
  "summary": {
    "breaking_count": 3,
    "warning_count": 1,
    "info_count": 5
  }
}
```

**Container entrypoint adjustment:** The existing `entrypoint.sh` accepts file paths as CLI args. For the container service, we add a new `serve` mode:
```bash
substrate serve --port 8080
```
This starts an HTTP server that accepts `POST /diff` requests, runs the diff engine, and returns `DiffReport` JSON. The existing CLI mode (for the GitHub Action) is unchanged.

---

## Component 3: GitHub App Registration

**App owner:** `kpakkiragari` (personal account for now — can migrate to `substratehq` org later by re-registering)

**App name:** `Substrate`

**Installation scope:** Both supported — GitHub natively handles "install for all repos" vs "install for selected repos" at the app level. No code change required; our Worker processes whatever webhook GitHub sends.

**Required permissions:**
| Permission | Level | Reason |
|---|---|---|
| `pull_requests` | Read & Write | Post PR comments |
| `statuses` | Read & Write | Set commit status checks |
| `contents` | Read | Fetch spec files from the repo |
| `metadata` | Read | Required by GitHub for all apps |

**Webhook events to subscribe:**
- `pull_request`

**Secrets (stored in Cloudflare Workers secrets):**
| Secret | Value |
|---|---|
| `GITHUB_APP_ID` | Numeric App ID from GitHub App settings |
| `GITHUB_APP_PRIVATE_KEY` | PEM private key generated at registration |
| `GITHUB_WEBHOOK_SECRET` | Random string set during App registration |

---

## Component 4: PR Comment Format

The PR comment is posted by the Worker using the GitHub installation access token. It follows a consistent markdown template:

**Breaking changes detected:**
```markdown
## 🔴 Substrate — Breaking Changes Detected

This PR introduces **3 breaking changes** to your OpenAPI contract.
Consumers of this API may break if this PR is merged without coordination.

| Severity | Rule | Path |
|---|---|---|
| 🔴 BREAKING | `response-property-removed` | `GET /users/{id}` → `data.email` |
| 🔴 BREAKING | `request-parameter-removed` | `POST /orders` → `body.currency` |
| 🟡 WARNING  | `response-property-type-changed` | `GET /products` → `items[].price` |

<details>
<summary>1 warning, 5 informational changes (click to expand)</summary>
...
</details>

---
*To acknowledge a breaking change, add an override to your `substrate.yaml`.*
*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*
```

**No breaking changes:**
```markdown
## ✅ Substrate — All Clear

No breaking changes detected in this PR. Safe to merge. 🎉

*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*
```

**Comment behavior:** The Worker upserts the comment (updates the existing one if it already exists, creates a new one if not). This prevents comment spam on repeated commits.

---

## Component 5: Commit Status Check

In addition to the PR comment, Substrate sets a commit status on the PR's HEAD commit:

```
Context: substrate/breaking-changes
State:   failure | success | pending
```

**Mapping:**
| DiffReport | Status Check |
|---|---|
| 0 breaking changes | `success` — "All clear — no breaking changes" |
| 1+ breaking changes, `on_breaking_change: block` | `failure` — "3 breaking changes detected" |
| 1+ breaking changes, `on_breaking_change: warn` | `success` — "3 breaking changes (advisory)" |

---

## `substrate.yaml` Configuration Additions (Phase 2)

Phase 2 adds two new top-level keys to `substrate.yaml`:

```yaml
# Existing Phase 1 keys (unchanged)
base_schema: path/to/base.yaml
head_schema: path/to/head.yaml

# New Phase 2 keys
on_breaking_change: block    # "block" (default) | "warn"
# block: fails the commit status check, PR cannot be merged without override
# warn:  posts a comment but status check passes — advisory only

# Phase 2 also fully inherits Phase 1's override config
overrides:
  - rule: response-property-removed
    path: GET /legacy/endpoint
    reason: "Planned removal, consumers notified"
    status: pending            # Phase 1b-T09 — pending acknowledgment
    until: 2026-09-01          # auto-expires
```

---

## Missing Config Behavior (No `substrate.yaml`)

When the GitHub App receives a PR event from a repo with no `substrate.yaml`:

1. Worker posts a **one-time setup comment** on the PR:
```markdown
## 👋 Substrate is installed but not configured

We noticed this repo has OpenAPI/SQL/GraphQL spec files but no `substrate.yaml`.

Run this in your repo root to activate breaking change protection:
\`\`\`bash
substrate init
\`\`\`
This generates a `substrate.yaml` in 30 seconds. [View setup guide →](https://github.com/KrushnaVardhanReddy/Substrate#quick-start)

*This comment will not appear again once `substrate.yaml` is added.*
```

2. Worker sets a **neutral commit status** (not success, not failure):
   - `State: pending`, Description: "substrate.yaml not found — run `substrate init` to activate"

3. Worker **does not block the merge** — no failure status is set.

**Why this approach:**
- Installs that don't configure are wasted growth. A setup prompt converts them.
- Auto-detecting spec files without `substrate.yaml` creates too many false positives.
- `substrate init` already exists (P1-T07) — this is the perfect activation hook.

---

## Deployment

**Worker:** Deployed via `wrangler deploy` to `substrate.kpakkiragari.workers.dev` (or a custom domain later).

**Container:** Deployed as a Cloudflare Container (or Fly.io service) running `kpakkiragari/substrate-engine:latest`. The Worker calls it via an internal service binding or HTTPS URL.

**Environment separation:**
- `production` — the live GitHub App
- `staging` — a second GitHub App registration (`Substrate (Staging)`) for safe testing

---

## First-Time Setup Checklist

- [ ] Register GitHub App at `github.com/settings/apps/new` under `kpakkiragari`
- [ ] Generate a private key, copy App ID
- [ ] Set webhook URL to the Cloudflare Worker URL
- [ ] Set webhook secret (random string)
- [ ] Add `GITHUB_APP_ID`, `GITHUB_APP_PRIVATE_KEY`, `GITHUB_WEBHOOK_SECRET` to Cloudflare Workers secrets
- [ ] Deploy the Container Service
- [ ] Deploy the Cloudflare Worker
- [ ] Install the App on the Substrate test repo and verify the first PR comment appears

---

## Security Considerations

- All webhook payloads are validated with HMAC-SHA256 before any processing.
- The GitHub App private key is stored only in Cloudflare Workers secrets — never committed.
- Installation access tokens are short-lived (1 hour) and are generated on-demand per request using the private key.
- The Container Service is not publicly accessible — only the Worker can reach it (via internal binding or private network).
- No user spec content is logged or persisted — diffs are computed in-memory and the result is returned immediately.
