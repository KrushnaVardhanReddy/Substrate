# P10-T20: Phase 10 E2E Validation (No Mocks)

## Overview
End-to-end testing of GraphQL Supergraphs, CDC Manifests, eBPF probes, and MCP diffing. Must use real live gateway environments and zero mock APIs.

## Requirements
1. **Test Environments**: Use Testcontainers for Postgres and a real GitHub test org for webhook integration tests.
2. **Coverage**: Tests must cover the AI Mock Data Generator, Spectral Linter, Auto-SDK Generator, and Zombie API detection.
3. **No Mocks**: All external service calls (GitHub, Stripe) must use sandboxed/test-mode credentials, never mock interceptors.
4. **CI Integration**: The test suite must run as a required CI step on the `feature/dev` branch.
