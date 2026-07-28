# Phase 17 - Task 03: GitOps API Gateway Sync (CRD Generation)

## 1. Goal
Automatically generate Kubernetes CRDs (specifically Kong Ingress resources or AWS API Gateway OpenAPI import payloads) from registered Substrate schemas when a PR is merged, eliminating manual DevOps configuration for API routing.

## 2. Requirements
- Add `POST /api/v1/gateway/sync/{org}/{repo}` — triggered by the merge webhook. Reads the current registered OpenAPI spec and generates a CRD YAML.
- Support two output modes configurable via `substrate.yaml`:
  - `gateway: kong` — generates a `KongIngress` + `Ingress` YAML with path rules matching the OpenAPI routes.
  - `gateway: aws` — generates an AWS API Gateway import payload JSON.
- Open a PR in the infra repository (configured in `substrate.yaml` as `infra_repo`) with the generated CRD file at `gateway/substrate-generated-{repo}.yaml`.
- Store the last generated CRD in a `gateway_configs` table for audit history.

## 3. `substrate.yaml` Extension
```yaml
gateway:
  type: kong           # kong | aws
  infra_repo: org/infra-repo
  output_path: gateway/
```

## 4. Database Schema
```sql
CREATE TABLE gateway_configs (
  id          SERIAL PRIMARY KEY,
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  gateway_type TEXT NOT NULL,
  crd_yaml    TEXT NOT NULL,
  pr_url      TEXT,
  generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
