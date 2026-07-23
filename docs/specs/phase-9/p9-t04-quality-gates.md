# P9-T04: Quality Gates (SonarQube-style)

## Overview
Allow orgs to set different failure thresholds per service tier. For example, Tier 1 services fail on any warning; Beta services may allow some breaking changes.

## Requirements
1. **Tier Configuration**: Parse `tier` and `quality_gate` from `substrate.yaml`, supporting levels: `strict`, `standard`, `permissive`.
2. **Gate Evaluation**: After running checks, evaluate the results against the configured gate for the repo's tier.
3. **CI Result**: Set the GitHub Check status to `failure` or `success` based on the gate result.
4. **Dashboard View**: Show the quality gate badge on the repo's card in the dashboard.

## Implementation Status: ✅ Implemented

### Database Schema
- Migration `0047_add_quality_gate.up.sql` adds a `quality_gate TEXT DEFAULT 'standard'` column to the `organizations` table.
- Supported values: `strict`, `standard`, `permissive`.

### E2E Validation
- Covered by `TestPhase9SystemE2E/Scenario_2:_Quality_Gates` in `scripts/e2e/phase9_e2e_test.go`.
- Test verifies gate evaluation returns correct pass/fail status based on configured tier.
