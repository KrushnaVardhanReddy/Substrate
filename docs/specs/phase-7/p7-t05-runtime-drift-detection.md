# P7-T05: Runtime Drift Detection (Sidecar)

## Objective
Static analysis (diffing schemas in CI/CD) is highly effective, but it relies on the schema being accurate. If a developer bypasses the schema and adds a hidden endpoint directly in the code, the schema "drifts" from reality.
Substrate will provide a lightweight Envoy/eBPF sidecar that samples live production traffic and compares it against the latest validated schema in the Substrate Registry to detect un-documented APIs or fields.

## Requirements

### 1. Substrate Sidecar Agent
- A standalone Go binary designed to run as a sidecar container in Kubernetes.
- Acts as a reverse proxy or packet sniffer (using eBPF).

### 2. Real-time Schema Validation
- The Sidecar fetches the latest schema from the Substrate Registry (`GET /api/v1/schema/{org}/{repo}`).
- It asynchronously samples 1-5% of incoming HTTP requests and responses.
- It compares the JSON payloads against the OpenAPI definitions.

### 3. Drift Reporting
- Target: `api/internal/handlers/telemetry.go`
- If the Sidecar observes an endpoint (e.g., `GET /v2/hidden/debug`) that does NOT exist in the Substrate Registry, it POSTs a `DriftAnomaly` back to the Substrate API.
- The Substrate Dashboard highlights these anomalies, warning the organization that their static spec is incomplete.
