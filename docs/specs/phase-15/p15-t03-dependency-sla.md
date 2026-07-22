# P15-T03: Dependency SLA Tracking

## Overview
Let consumer teams declare `required_notice_days` in their `substrate.yaml`. Substrate will warn provider teams when a proposed breaking change breaches a declared SLA before the PR is merged. This creates an enterprise compliance paper trail for inter-team contracts.

## Requirements
1. **Parser Update**: Parse `required_notice_days` (int) from `substrate.yaml` for each consumed API.
2. **SLA Validation**: During a breaking change detection (`substrate check`), compare the current date with the expected release date (if known) or simply generate an SLA breach warning if `required_notice_days` > 0.
3. **GitHub PR Comment Integration**: In the PR comment, explicitly list the breached SLAs and the required notice days. For example: "⚠️ SLA Breach: 'billing-api' requires 30 days notice for breaking changes."
4. **Data Model**: Extend the Consumer contract model in Postgres to store `required_notice_days`.
