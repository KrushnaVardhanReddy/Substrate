---
title: Intentional Breaking Changes
description: Acknowledging intentional breaking changes in Substrate.
---

Sometimes a breaking change is fully intended (e.g., you've coordinated with all consumer teams and they have successfully migrated off the old endpoint).

To prevent Substrate from eternally blocking your CI pipeline, you can **acknowledge** the break using the `overrides` block in your `substrate.yaml` file:

1. Copy the `rule_id` and the `path` exactly as they appear in the Substrate PR comment (e.g., `rule_id: ENDPOINT_REMOVED`).
2. Add an explanation in the `reason` field (e.g., linking to a Jira ticket or RFC).
3. Set an `expires` date. This ensures the override doesn't stay in the codebase forever.

When you commit this `substrate.yaml` update, the Substrate engine will read the override, downgrade the `❌ BREAKING` error to a `✅ Acknowledged` status, and **allow your pipeline to pass**.

## Examples by Schema Type

The override format is exactly the same regardless of your schema type, but the `rule_id` and `path` will match the specific engine output.

### OpenAPI (REST)
```yaml
overrides:
  - rule_id: ENDPOINT_REMOVED
    path: paths./legacy/api
    reason: "Consumers migrated to v2 (TICKET-123)"
    expires: 2026-12-31
```

### GraphQL
```yaml
overrides:
  - rule_id: GQL_FIELD_REMOVED
    path: type.User.field.phoneNumber
    reason: "Phone numbers deprecated in favor of email auth"
    expires: 2026-12-31
```

### PostgreSQL (SQL)
```yaml
overrides:
  - rule_id: COLUMN_REMOVED
    path: table.users.column.address
    reason: "Address data moved to separate address-service database"
    expires: 2026-12-31
```

### Protobuf (gRPC)
```yaml
overrides:
  - rule_id: PROTO_FIELD_TYPE_CHANGED
    path: message.User.field.id
    reason: "Migrating from int32 to int64 for scalability"
    expires: 2026-12-31
```