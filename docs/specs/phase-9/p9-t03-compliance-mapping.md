# P9-T03: Compliance Mapping

## Overview
Auto-tag schemas with SOC2/GDPR/HIPAA warnings when fields like `ssn`, `medical_history`, or `credit_card` are detected in OpenAPI specs.

## Requirements
1. **Field Scanner**: Scan OpenAPI schema field names and descriptions against a curated list of sensitive field patterns.
2. **Tagging**: Inject `x-compliance: [GDPR, HIPAA]` extensions into matched fields.
3. **PR Comment**: Surface compliance tags in the GitHub PR comment with a dedicated "⚠️ Compliance Flags" section.
4. **Configurable Rules**: Allow orgs to define custom field patterns in `substrate.yaml` under `compliance.patterns`.
