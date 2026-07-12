# Substrate — Future Vision & Roadmap

This document outlines the long-term strategic vision and roadmap for Substrate. These features transition the product from a developer tool that prevents schema breaks into a Fortune 500 Enterprise Automated Governance Platform.

## Phase 7: Enterprise / Scale

*   **Traffic-Aware Diffing (Zero False Positives):** Integrate with Datadog, OpenTelemetry, or Prometheus to check production traffic. If a breaking change targets an endpoint/field with zero traffic in the last 30 days, Substrate auto-downgrades the severity from `BREAKING` to `WARNING (Unused)`.
*   **Auto-SDK & Client Generation (Zero-Touch Sync):** Automatically open PRs in all consumer repositories with newly generated TypeScript/Go/Python SDK clients when a provider safely updates an API.
*   **Local "Time-Travel" Mock Servers:** With `substrate mock --env production`, instantly spin up local mock servers by pulling exact production schemas from the Substrate Registry.
*   **The Platform ROI Dashboard:** A management dashboard calculating ROI (e.g., "Substrate prevented 14 cross-repo breaking changes this month... saved $70,000").
*   **Shift-Left IDE Plugins:** A VSCode/IntelliJ extension powered by the Substrate MCP server to warn developers *before* they commit changes that will break downstream consumers.
*   **Cross-Repo Auto-Fix PRs:** Use LLMs to automatically generate draft PRs in consumer repositories to remove their dependencies on fields being deprecated in a provider repo.

## Phase 8: Governance & Policy

*   **Execution Modes (`--mode`):**
    *   `--mode=legacy` (Relaxed): Only stops PRs on catastrophic breaks (e.g., deleted fields).
    *   `--mode=default`: Standard breaking change prevention with informational warnings.
    *   `--mode=strict` (Spec-First): For new microservices, enforcing flawless design (descriptions, versioning, naming conventions).
*   **Custom Rules Engine (CEL / OPA Rego):** Integrate CEL or OPA to encode internal PDF guidelines into `substrate.yaml` (e.g., "All APIs must return a correlation ID").
*   **Compatibility Gates:** Define gates based on service criticality (e.g., Tier 1 services allow 0 breaking/warnings; Beta services allow breaking changes if acknowledged).
*   **ITSM Integration (Jira & ServiceNow Dynamic Approvals):** Automatically open a ServiceNow Change Request or Jira Ticket for intentional breaking changes, routing approvals specifically to the Tech Leads of affected consumer teams.

## Phase 9: Security & Ecosystem

*   **Runtime Drift Detection (eBPF / Envoy Filter):** A lightweight Substrate Envoy filter or eBPF agent sitting at the API Gateway, sampling live traffic to ensure actual JSON payloads match the schemas in the Substrate Registry.
*   **Security & PII Auditing:** Detect if a PR accidentally exposes sensitive data (e.g., `ssn`, `password`, `token`, `card_number`), blocking the PR and flagging the security team.
*   **Cross-Repo Data Flow Taint Analysis:** Trace the flow of specific data fields globally. Prevent PII from flowing into unauthorized systems (e.g., blocking ingestion of an unencrypted PII field into a data warehouse).
*   **Compliance Mapping:** Map schema changes directly to SOC2, GDPR, or HIPAA requirements (e.g., adding `medical_history` triggers a `[HIPAA]` tag and alerts Compliance).
