# P15-T13: Phase 15 E2E Validation (No Mocks)

## Overview
Live orchestration testing of KMS BYOK, LLM workload routing, and Granular Check Suites.

## Requirements
1. **Live AWS/Vault**: Tests must use real AWS KMS test keys and a Vault dev server (via Docker).
2. **Coverage**: Validate: NL Governance Rule generation, Marketplace plugin install/uninstall, Schema Insurance claim filing.
3. **No Mocks**: Zero mock interceptors allowed. Use sandbox API keys for all external services.
4. **CI Step**: Must pass before any Phase 15 feature is considered complete.
