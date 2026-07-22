# P10-T05: Auto-SDK Generator PRs

## Overview
Automatically generate TypeScript/Swift/Go clients via OpenAPI Generator when a schema is merged, opening PRs directly in downstream consumer repos.

## Requirements
1. **Trigger**: On `push` to main (after schema merge), trigger the SDK generator workflow.
2. **Generator**: Use `openapi-generator-cli` to generate TypeScript, Swift, and Go SDKs from the merged OpenAPI spec.
3. **PR Opening**: For each registered consumer repo, open a PR with the regenerated SDK client code.
4. **Configuration**: Allow orgs to specify which SDK languages to generate in `substrate.yaml`.
