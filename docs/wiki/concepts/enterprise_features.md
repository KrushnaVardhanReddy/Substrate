# Enterprise Features

Substrate is heavily optimized for enterprise and highly-regulated environments.

## Bring Your Own Key (BYOK)
- **KMS Integration**: Enterprises can encrypt their schemas at rest using AWS KMS or HashiCorp Vault.
- **LLM Routing**: AI workloads can be routed through proprietary Azure OpenAI / AWS Bedrock VPC endpoints to ensure IP never leaves the enterprise perimeter.

## CRM & Billing Blast Radius
Integrations with Stripe and Salesforce allow Substrate to calculate the exact MRR and customer count impacted by breaking schema changes. This data is surfaced directly in GitHub PR comments.

## Quality Gates & Compliance
- **SonarQube-Style Gates**: Different failure thresholds can be set based on service tiers.
- **Compliance Mapping**: Schemas are automatically tagged with SOC2, GDPR, or HIPAA warnings when PII (like `ssn`) is detected.
