# V1-T09: True Dogfooding via OpenAPI

## Objective
To strictly adhere to the "Spec-First" philosophy, the Substrate repository must analyze itself using its own CI/CD engine. Because the engine was previously undocumented, the GitHub App dogfooding checks were hanging on missing configuration. This task formalizes the internal Substrate API into a valid OpenAPI specification so that the engine can monitor and block its own breaking changes.

## Requirements

### 1. Define Substrate OpenAPI Spec
- File: `openapi.yaml` in the repository root.
- Document the core `api/internal/handlers` routes:
  - `GET /health`
  - `GET /api/v1/graph/{org}`
  - `POST /api/v1/webhook`
  - `POST /api/v1/diff`
  - `POST /api/v1/ai/analyze`
  - `POST /api/v1/ai/autofix`

### 2. Configure Substrate Config
- File: `substrate.yaml`
- Update the configuration to point to the newly created `openapi.yaml`.

### 3. Verify Dogfooding Status
- The GitHub App must detect `substrate.yaml`, locate the `openapi.yaml` file, and perform a diff on pull requests to the Substrate repository itself.
