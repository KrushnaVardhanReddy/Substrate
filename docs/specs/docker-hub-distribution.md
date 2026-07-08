# Substrate — Distribution Model Specification

> **Status:** APPROVED ✅
> **Spec-First Gate:** No changes to `action.yml`, release CI, or Docker Hub configuration are made until this document is approved.
> **Scope:** Defines how Substrate distributes the GitHub Action to end users from a private repository, with zero source code exposure.

---

## Problem Statement

The Substrate repository is private (source code, specs, and roadmap are confidential). However, GitHub Actions require the action repository to be **public** for external repos to use `uses: owner/repo@version`. Building the image from source at action runtime (`image: "Dockerfile"`) also requires the source to be accessible.

**Requirement:** External users must be able to install and run the Substrate GitHub Action without the source repository being public.

---

## Solution: Pre-Built Docker Hub Public Image

The action is distributed as a **pre-built Docker image** on Docker Hub. The source repo remains private. Only the compiled binary image is made public.

```
Private Repo (source)  →  CI Build  →  Docker Hub (public image)  →  GitHub Action
```

### Why Docker Hub over GitHub Container Registry (GHCR)?

| Factor | Docker Hub | GHCR |
|---|---|---|
| Visibility from private repo | ✅ Image can be public independently | 🟡 Image tied to repo visibility by default |
| `action.yml` `image:` field support | ✅ `docker://` protocol supported | ✅ Supported |
| Account requirement | Docker Hub org account | GitHub org (already have) |
| Industry standard for Actions | ✅ Widely used | ✅ Increasingly used |

**Decision:** Docker Hub. No technical reason to prefer GHCR, and Docker Hub is more familiar to the developer community using GitHub Actions.

---

## Docker Hub Configuration

### Image Name
```
kpakkiragari/substrate-engine
```

### Tags Published Per Release

| Tag | Example | Purpose |
|---|---|---|
| Version tag | `kpakkiragari/substrate-engine:v0.2.0` | Pinned, immutable release |
| `latest` | `kpakkiragari/substrate-engine:latest` | Always points to newest release |

### Required Secrets (GitHub Actions)

| Secret Name | Value | Where to Set |
|---|---|---|
| `DOCKERHUB_USERNAME` | `kpakkiragari` | GitHub repo Settings → Secrets → Actions |
| `DOCKERHUB_TOKEN` | Docker Hub access token (read/write) | GitHub repo Settings → Secrets → Actions |

**How to generate a Docker Hub token:** Docker Hub → Account Settings → Security → New Access Token → permissions: `Read, Write, Delete`.

---

## `action.yml` Contract

The `runs` block in `action.yml` references the Docker Hub image directly. It must never reference `"Dockerfile"` (requires public source).

```yaml
runs:
  using: "docker"
  image: "docker://kpakkiragari/substrate-engine:v0.X.Y"   # always pin to exact version
  args:
    - ${{ inputs.base_schema }}
    - ${{ inputs.head_schema }}
    - ${{ inputs.config }}
```

**Rule:** The version tag in `action.yml` must always match the latest published Docker Hub image tag. The release CI workflow auto-updates this on every tag push.

---

## Release CI Workflow Contract

**File:** `.github/workflows/release.yml`

**Trigger:** Push of any tag matching `v*.*.*`

**Required Permissions:**
The workflow must explicitly declare `permissions: contents: write` so that it can push the updated `action.yml` file back to the repository.

**Steps (in order):**
1. Checkout source
2. Set up Docker Buildx (multi-platform builds)
3. Log in to Docker Hub using `DOCKERHUB_USERNAME` + `DOCKERHUB_TOKEN` secrets
4. Build the `Dockerfile` in the repo root
5. Push two tags: `kpakkiragari/substrate-engine:v0.X.Y` (exact) and `kpakkiragari/substrate-engine:latest`
6. Auto-update `action.yml` to pin to the new version tag and commit back to `main`

**Invariant:** After every release tag push, `action.yml` and the Docker Hub image are always in sync. No manual step required.

---

## Release Process (for every future version)

```bash
# 1. Ensure all changes are merged to main
# 2. Tag the release
git tag v0.2.0
git push origin v0.2.0

# 3. CI workflow fires automatically:
#    - Builds Dockerfile
#    - Pushes kpakkiragari/substrate-engine:v0.2.0 and :latest to Docker Hub
#    - Auto-commits action.yml pinned to v0.2.0

# 4. GitHub Marketplace picks up the new tag automatically
```

---

## First-Time Setup Checklist

- [ ] Create public repository `substrate-engine` in your Docker Hub account
- [ ] Generate Docker Hub access token (Read, Write, Delete)
- [ ] Add `DOCKERHUB_USERNAME` secret to GitHub repo
- [ ] Add `DOCKERHUB_TOKEN` secret to GitHub repo
- [ ] Push first tag (`v0.1.1` or higher) to trigger initial image push
- [ ] Verify `kpakkiragari/substrate-engine:latest` appears on Docker Hub
- [ ] Verify `action.yml` is auto-updated by the CI workflow

---

## Build Architecture & CGO

**Requirement:** The Go binary must be compiled with `CGO_ENABLED=1`.

The Phase 1b SQL Diff Engine relies on `github.com/pganalyze/pg_query_go`, which is a Go wrapper around the official PostgreSQL C parser. Because it uses C code:
- The Docker `builder` stage cannot use a plain Alpine image without a C compiler.
- We must use `golang:alpine` (or a specific modern version) and explicitly install `gcc` and `musl-dev` before running `go build`.
- The final binary is dynamically linked to `musl` libc, so the final runtime image must also be Alpine-based (`FROM alpine:latest`).

*Any future changes to the `Dockerfile` must preserve the CGO enabled build step and the C compiler dependencies.*

---

## Security Considerations

- The Docker image contains only the compiled Go binary and its runtime dependencies. No source files, no spec documents, no prompts, no `.env` files are baked in.
- The `Dockerfile` uses a multi-stage build: build stage compiles the binary, final stage is a minimal `alpine` image with only the binary.
- Docker Hub credentials are stored only as GitHub Actions secrets — never in code or committed files.
- The `DOCKERHUB_TOKEN` should have the minimum required permissions (Read, Write, Delete on the `substrate-engine` repo only).
