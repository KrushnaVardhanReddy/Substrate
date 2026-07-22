# P15-T12: Enterprise BYOK (Bring Your Own Key)

## Objective
Implement a dual-BYOK architecture (Data-at-Rest Encryption + LLM Workload Routing) to satisfy the strictest enterprise InfoSec requirements, enabling highly-regulated companies (finance, healthcare, government) to safely run Substrate in their VPCs without risking their proprietary API architectures.

## 1. Encryption BYOK (Data-at-Rest)
When an enterprise deploys the Substrate single-binary, they can configure it to encrypt all sensitive data (API schemas, database mappings) before it is written to the underlying PostgreSQL database.
- **Implementation:** Add an encryption middleware to `api/internal/db/`.
- **Key Management:** Integrate with AWS KMS, GCP Cloud KMS, and HashiCorp Vault. The app fetches a Data Encryption Key (DEK) wrapped by the enterprise's KMS.
- **Value:** If the enterprise revokes the KMS key, the entire Substrate database is cryptographically shredded instantly.

### Database Architecture: Toggled Payload Columns
To ensure we do not need separate databases for Free/Pro vs. Enterprise users, the PostgreSQL schema must use a dual-column fallback strategy:
```sql
ALTER TABLE contracts
  ADD COLUMN encrypted_content BYTEA,    -- For Enterprise BYOK users (AES-256-GCM)
  ADD COLUMN is_encrypted BOOLEAN DEFAULT false,
  ADD COLUMN kms_key_arn VARCHAR(255);
```
The application's repository layer (`api/internal/db`) will dynamically choose which column to write to/read from based on the user's tier. This keeps the core diff `engine/` completely unaware of the encryption layer.

## 2. LLM BYOK (Data-in-Motion / AI)
Substrate's most advanced features (Migration Planner, Contract Negotiation) rely on LLMs. Enterprises will not allow their proprietary schemas to be sent to a multi-tenant OpenAI SaaS endpoint.
- **Implementation:** Expand `substrate.yaml` to accept an AI provider configuration block.
- **Providers:** Allow routing all LLM calls to the enterprise's own Azure OpenAI, AWS Bedrock, or local Ollama instances.
- **Value:** Data never leaves the corporate perimeter, and enterprises leverage their pre-negotiated cloud credits and zero-data-retention agreements.

## Deliverables
- `api/internal/crypto/` package handling Envelope Encryption via KMS/Vault.
- Updates to `engine/pkg/ai/` to abstract the OpenAI client and allow plugging in Azure/Bedrock SDKs based on the `substrate.yaml` configuration.
- Comprehensive unit testing for cryptographic round-trips and AI provider switching.
