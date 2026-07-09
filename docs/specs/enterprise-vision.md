# Substrate — Enterprise Vision & Future Roadmap

This document captures the long-term strategic vision for Substrate. These features transition the product from a "startup developer tool" that prevents schema breaks into a **Fortune 500 Enterprise Automated Governance Platform**.

---

## 1. Automated Governance & Customization

### Execution Modes (`--mode`)
Large organizations migrate at different speeds. Applying strict schema checks to a 5-year-old legacy monolith creates friction and blocks adoption.
*   **`--mode=legacy` (Relaxed):** Only stops PRs on catastrophic breaks (e.g., deleted fields). Ignores missing documentation, missing pagination, or styling.
*   **`--mode=default`:** Standard breaking change prevention with informational warnings.
*   **`--mode=strict` (Spec-First):** For new microservices. Enforces flawless design: every field must have a description, endpoints must be versioned, naming conventions (`snake_case`) must be uniform, and all warnings are treated as hard failures.

### Custom Rules Engine (CEL / OPA Rego)
Every enterprise has highly specific internal governance rules (e.g., *"All APIs must return a correlation ID header"*, *"No columns dropped without a specific PR tag"*). 
Integrating **CEL (Common Expression Language)** or **Open Policy Agent (OPA)** allows platform teams to encode their internal PDF guidelines directly into `substrate.yaml`:
```yaml
custom_rules:
  - id: REQUIRE_CORRELATION_ID
    schema_type: openapi
    condition: "!has(head.headers['X-Correlation-ID'])"
    message: "Enterprise Policy: All APIs must return a correlation ID."
    severity: BREAKING
```

---

## 2. Closing the Loop: Developer Experience

### Traffic-Aware Diffing (Zero False Positives)
The biggest pain point in static analysis is the false positive: flagging a deleted field that *no one actually uses*. 
By integrating with **Datadog, OpenTelemetry, or Prometheus**, Substrate checks production traffic. If a breaking change targets an endpoint/field with **0 traffic in the last 30 days**, Substrate auto-downgrades the severity from `BREAKING` to `WARNING (Unused)`. This allows graceful deprecation and builds immense developer trust.

### Auto-SDK & Client Generation (Zero-Touch Sync)
Because the Substrate Contract Registry knows exactly which frontend consumes which backend, when a provider safely updates an API, Substrate **automatically opens PRs in all consumer repositories** with the newly generated TypeScript/Go/Python SDK clients. Substrate becomes the automated nervous system of the organization.

### Local "Time-Travel" Mock Servers
Developers struggle to test against distributed microservices locally. With the command `substrate mock --env production`, Substrate pulls the exact schemas of all live production services from the registry and instantly spins up local mock servers. The developer gets a perfect local replica of the production layer in seconds.

---

## 3. Extending the Moat: Full-Stack & Runtime

### Runtime Drift Detection (eBPF / Envoy Filter)
Static analysis in CI is great, but code drifts. A developer might push a hotfix, or a 3rd-party library might alter a payload silently. 
An ultra-lightweight Substrate Envoy filter or eBPF agent sits at the API Gateway, sampling 1% of live traffic to ensure the actual JSON payloads match the schemas in the Substrate Registry. If a service emits an undocumented payload, Substrate triggers an alert: *"Runtime Drift Detected: API is returning 'null' for required field."*

### Security & PII Auditing
Substrate can detect if a PR accidentally exposes sensitive data. If a schema update adds an API response field named `ssn`, `password`, `token`, or `card_number`, Substrate blocks the PR and flags the security team. 

### The Platform ROI Dashboard
To justify Enterprise tier pricing to a VP of Engineering, Substrate provides a management dashboard calculating literal ROI:
*"Substrate prevented 14 cross-repo breaking changes this month. At an average incident cost of $5,000, Substrate saved the company $70,000 and 84 hours of downtime."*
