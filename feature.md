# Substrate

## What is Substrate?
Substrate is an enterprise-grade API governance, schema validation, and dependency tracing platform. It serves as the single source of truth for microservice contracts, allowing organizations to automatically track who depends on what, and actively preventing breaking changes before they reach production. 

## What We Are Trying to Build
We are building a highly autonomous, AI-augmented governance engine that bridges the gap between static schema validation and live observability. Substrate is designed to provide "spec-first" development guarantees by intercepting schema changes, calculating their blast radius, and intelligently enforcing backward compatibility across massive microservice ecosystems.

With the latest phases (Phase 19-21), we are expanding Substrate from a static GitHub-driven registry into a **Universal Ingestion Hub** that supports Airbyte data pipelines, OpenTelemetry (OTel) traffic metrics, and Kafka event streams.

## Core Features

### 1. The Dependency Graph (Service Tracing)
- **Live Topology:** Automatically maps out the entire ecosystem of providers and consumers.
- **Contract Registration:** Parses OpenAPI, AsyncAPI, Avro, and Protobuf schemas to understand exact API surface areas.
- **Zombie Detection:** Identifies unused, orphaned, or deprecated endpoints to help reduce technical debt.

### 2. Blast Radius & Diff Engine
- **Breaking Change Detection:** Deep, structural diffing of schemas to detect when an endpoint is removed, a required field is added, or a type is changed.
- **Impact Analysis:** Calculates the exact blast radius of a change, instantly flagging downstream consumers (frontends, databases, other microservices) that will break.
- **Scorecards:** Grades repositories based on their schema hygiene and backward-compatibility track record.

### 3. AI & MCP (Model Context Protocol) Integration
- **Headless AI Governance:** Exposes Substrate's diff engine and graph directly to LLMs (like Claude/Jules) via MCP tools.
- **Auto-Remediation:** AI can propose PRs to downstream consumers to fix breaking changes automatically.
- **Human-in-the-Loop (HITL):** Governance gates that require human approval for catastrophic schema changes or specific AI-driven pull requests.

### 4. Universal Data & Event Ingestion
- **Airbyte Adapter (Phase 19):** Ingests and normalizes external data catalogs and third-party SaaS schemas directly into the Substrate graph.
- **OTel Traffic Ingestion (Phase 20):** Enriches static nodes with real live traffic weights (req/min, p99 latency), shifting risk assessment from binary (broken/not broken) to quantitative.
- **Kafka Schema Governance (Phase 21):** Extends strict breaking-change detection to event-driven architectures via Confluent/Apicurio webhooks.

### 5. Enterprise Ecosystem & Security
- **PASETO v4 Local Security:** Hardened internal service-to-service authentication.
- **BYOK (Bring Your Own Key) KMS:** Tenant-level data isolation and encryption.
- **Webhooks & Sync:** Real-time synchronization with GitHub repositories, triggering automatic analysis on every commit.
- **Insurance & ROI:** Quantifies the financial value of outages prevented and measures the ROI of the governance pipeline.
