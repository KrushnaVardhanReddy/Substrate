---
title: Rule Reference
---

## 5. Rule Reference

Substrate supports over 100 semantic breaking change rules across 8 different schema adapters.

| Schema Type | Substrate Engine | Example Breaking Rule | Example Safe Rule |
|-------------|------------------|-----------------------|-------------------|
| **OpenAPI** | `oasdiff` | `ENDPOINT_REMOVED` | `ENDPOINT_ADDED` |
| **GraphQL** | `gqlparser` | `GQL_FIELD_REMOVED` | `GQL_TYPE_ADDED` |
| **SQL (PG)**| `pg_query_go` | `COLUMN_REMOVED` | `TABLE_CREATED` |
| **Protobuf**| `buf` | `PROTO_FIELD_TYPE_CHANGED` | `PROTO_FIELD_ADDED` |
| **AsyncAPI**| `parser-go` | `ASYNCAPI_CHANNEL_REMOVED` | `ASYNCAPI_CHANNEL_ADDED` |
| **Avro**    | Custom | `AVRO_INCOMPATIBLE` | `AVRO_COMPATIBLE` |

*(Note: AI/ML Models and Salesforce Enterprise metadata are also natively supported).*

---
