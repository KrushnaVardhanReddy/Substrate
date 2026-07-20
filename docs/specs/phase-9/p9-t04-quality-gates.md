# P9-T04: Quality Gates (SonarQube-style)

## Overview
Allow orgs to set different failure thresholds per service tier. For example, Tier 1 services fail on any warning; Beta services may allow some breaking changes.

## Requirements
1. **Tier Configuration**: Parse `tier` and `quality_gate` from `substrate.yaml`, supporting levels: `strict`, `standard`, `permissive`.
2. **Gate Evaluation**: After running checks, evaluate the results against the configured gate for the repo's tier.
3. **CI Result**: Set the GitHub Check status to `failure` or `success` based on the gate result.
4. **Dashboard View**: Show the quality gate badge on the repo's card in the dashboard.
