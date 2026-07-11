---
title: Configuration Reference
---

## 2. Configuration Reference (`substrate.yaml`)

The `substrate.yaml` file lives in the root of your repository. It is entirely optional; without it, Substrate runs with sensible defaults.

```yaml
service: my-users-api           # Required: The canonical name of your service
schema_type: openapi            # Required: openapi | graphql | sql | protobuf | avro | asyncapi
spec_path: docs/openapi.yaml    # Required: Path to your schema file

# --- Phase 3: Consumers (Manual Declarations) ---
consumers:
  - name: frontend-dashboard
    provider_repo: myorg/frontend

# --- Phase 4: Traffic-Aware Diffing ---
traffic:
  provider: prometheus
  endpoint: "http://prometheus.internal:9090"
  lookback_days: 30
  downgrade_threshold: 0        # Changes to endpoints with 0 traffic are downgraded to WARNING

# --- Phase 4: Compliance Auditing ---
compliance:
  slack_webhook: "https://hooks.slack.com/services/..."
  security_channel: "#security-alerts"

# --- Intentional Breaking Changes ---
overrides:
  - rule_id: ENDPOINT_REMOVED
    path: paths./legacy/api
    reason: "Consumers migrated to v2"
    expires: 2026-12-31
```
