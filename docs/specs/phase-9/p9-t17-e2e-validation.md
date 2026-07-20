# P9-T17: Phase 9 E2E Validation (No Mocks)

## Overview
Comprehensive end-to-end testing for all Phase 9 features. Must use real Postgres DBs, real Substrate CLI integrations, and real Git repositories — strictly no mocking.

## Requirements
1. **Test Setup**: Use Testcontainers or Docker Compose to spin up a real Postgres instance and a local **Forgejo/Gitea** instance for each test run.
2. **Coverage**: Write E2E tests covering: compliance mapping detection, quality gate enforcement, the `watch` daemon, and deprecation campaign issue creation via the Forgejo API (simulating GitHub's API).
3. **CI Pipeline**: Integrate the test suite into the CI pipeline.
4. **Teardown**: Ensure all resources (containers, test repos, Forgejo issues) are cleaned up after test runs.
