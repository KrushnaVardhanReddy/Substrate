# P14-T07: Phase 14 E2E Validation (No Mocks)

## Overview
Validate all Phase 14 predictive intelligence and viral growth features end-to-end.

## Requirements
1. **Live Webhooks**: Use ngrok or a real staging environment to receive live GitHub webhooks during tests.
2. **Coverage**: Test the Contract Score Badge SVG endpoint, the Archaeology CLI command against a real git repo, and the Contract Negotiation workflow.
3. **PR Interaction Tests**: Use the GitHub API to open real PRs and validate automated bot responses.
4. **Cleanup**: All test repos and PRs created during runs must be automatically deleted.
