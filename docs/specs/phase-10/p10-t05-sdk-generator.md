# Phase 10 - Task 05: Auto-SDK Generator PRs

## 1. Goal
Automatically generate TypeScript/Swift/Go clients via OpenAPI Generator when a schema is merged, opening PRs directly in the downstream consumer repos.

## 2. Requirements
- Trigger an OpenAPI Generator job upon schema merge.
- Automatically find downstream repos from Substrate Graph.
- Open PRs updating their API client SDKs.
