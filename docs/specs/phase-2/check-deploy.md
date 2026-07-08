# Substrate — `check-deploy` Command Specification

> **Status:** DRAFT
> **Phase:** Phase 2 / 3 (Requires Phase 3 Contract Registry)
> **Scope:** Defines the behavior, inputs, and outputs of the `substrate check-deploy` CLI command.

---

## Problem Statement

A provider repository (e.g., `users-api`) merges a pull request with an approved breaking change (the change was overridden in `substrate.yaml`). The PR is merged, and the CI/CD pipeline starts to deploy the `users-api` to production.

However, if the consumer repositories (e.g., `frontend`, `mobile-app`) have not yet deployed their required updates to handle the breaking change, deploying the provider will break production. 

**Requirement:** CI/CD pipelines need a command to check the global Substrate Contract Registry *before* deploying to production. If consumers are not ready, the deployment should be blocked.

---

## The `check-deploy` Command

The `substrate check-deploy` command queries the centralized Substrate Registry to verify if a specific deployment is safe.

### Usage
```bash
substrate check-deploy \
  --repo "github.com/myorg/users-api" \
  --commit "a1b2c3d4" \
  --env "production"
```

### Authentication
Requires `SUBSTRATE_API_KEY` to be set in the environment (obtained from the Substrate Dashboard).

### Execution Flow
1. **Query Registry:** The CLI sends an HTTP GET to `https://api.substrate.dev/v1/check-deploy?repo=...&commit=...&env=production`.
2. **Registry Verification:** 
   - The Substrate API looks up the schema for `a1b2c3d4`.
   - It identifies all registered consumers for `users-api`.
   - It checks the currently deployed versions of those consumers in `production` (consumers report their deployed versions via their own CI/CD).
   - It computes the compatibility matrix.
3. **Response:** The API returns the safety status.

---

## Outputs and Exit Codes

### 1. Safe to Deploy (Exit Code 0)
```text
✅ SAFE TO DEPLOY

Provider: github.com/myorg/users-api (a1b2c3d4)
Environment: production

Compatibility Check:
- frontend (deployed: v1.4.2) -> COMPATIBLE
- mobile-app (deployed: v2.0.0) -> COMPATIBLE
```

### 2. Unsafe to Deploy (Exit Code 1)
```text
❌ DEPLOYMENT BLOCKED

Provider: github.com/myorg/users-api (a1b2c3d4)
Environment: production

The following consumers in production are INCOMPATIBLE with this deployment:
- frontend (deployed: v1.4.1)
  - Missing field: `data.email`
  - Required action: Wait for `frontend` to deploy v1.4.2 or higher.

Use --force to ignore and deploy anyway.
```

---

## CI/CD Integration Example

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production
on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Check Substrate Registry
        run: substrate check-deploy --repo ${{ github.repository }} --commit ${{ github.sha }} --env production
        env:
          SUBSTRATE_API_KEY: ${{ secrets.SUBSTRATE_API_KEY }}

      - name: Deploy
        run: ./deploy.sh
```
