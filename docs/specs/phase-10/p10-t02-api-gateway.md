# P10-T02: API Gateway Auto-Sync (Cloud Integrations)

## Objective
Develop a new Substrate CLI command (`substrate gateway sync`) that automatically pushes validated OpenAPI schemas to AWS API Gateway or Kong immediately after a PR is successfully merged to `main`.

## Context
Polling API gateways for changes creates lag and race conditions. Instead, Substrate enforces a **"Push-on-Merge" CI/CD architecture**. When the GitHub Action detects a merge to the default branch, it automatically deploys the validated schema to the configured API Gateway, ensuring the code repository remains the strict source of truth.

## Requirements

### 1. Integration Adapters
- Implement modular Go adapters for connecting to:
  - AWS API Gateway (via AWS SDK `apigateway.Client`).
  - Kong (via Kong Admin API `net/http`).
- The adapters should authenticate securely using CI/CD environment variables (e.g., AWS IAM OIDC roles).

### 2. CI/CD "Push" Execution
- Create a new CLI command: `substrate gateway sync --schema openapi.yaml`.
- The CLI parses the `substrate.yaml` config to find gateway targets.
- It deploys the schema directly to Kong or AWS API Gateway, blocking the CI/CD pipeline if the gateway rejects the schema syntax.

### 3. Substrate.yaml Configuration
- Allow users to define gateway integrations in their `substrate.yaml`:
```yaml
gateways:
  - type: kong
    url: https://kong-admin.internal
  - type: aws_apigw
    api_id: "xyz123"
```

## Acceptance Criteria
1. The Go backend can successfully authenticate with AWS API Gateway and Kong to push an OpenAPI spec.
2. The CLI command exits with code `0` on success, or code `1` if the gateway rejects it.
3. No background polling daemons are used (strictly push-based).
