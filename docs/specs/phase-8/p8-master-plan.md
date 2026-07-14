# Phase 8: Enterprise Readiness & Scale (Master Plan)

## Overview
Phase 8 transforms Substrate from a highly-functional startup tool into a robust, scalable, enterprise-grade platform. The goal is to establish the "V1.0 Moat"—features that make Substrate indispensable and highly secure for large engineering organizations.

## Core Initiatives (Specs)

### P8-T01: PostgreSQL Job Queue (River)
See `p8-t01-postgres-job-queue.md`

### P8-T02: Enterprise Authz (Casbin/OpenFGA)
See `p8-t02-enterprise-authz.md`

### P8-T03: CI/CD Cascading Rollback Gate
- **Objective**: Prevent a service from rolling back to an older schema if downstream consumers have already deployed code that relies on the newer schema.
- **Requirements**:
  - Add `substrate check-rollback` CLI command.
  - Query the API to ensure no active consumer dependencies will break if the provider reverts to the target git SHA.

### P8-T04: Spotify Backstage Plugin
- **Objective**: Meet enterprise developers where they already are by piping Substrate data into Backstage.io.
- **Requirements**:
  - Build a Node.js/React Backstage plugin (`@substrate/backstage-plugin`).
  - Display the dependency graph and API health scores directly on the Backstage Component pages.

### P8-T05: Distributed Tracing (OpenTelemetry)
- **Objective**: Allow operators to trace webhook requests end-to-end.
- **Requirements**:
  - Integrate `go.opentelemetry.io/otel` in the Go API and Diff Engine.
  - Export traces for webhook ingestion, schema parsing, and diff processing to OTLP collectors (Jaeger/Datadog).

### P8-T06: Management ROI Dashboard
See `p8-t06-roi-dashboard.md`

### P8-T07: Billing & Subscription Engine (Paywall Pause)
See `p8-t07-billing-engine.md`

### P8-T08: Single Binary VPC Deployment
- **Objective**: Allow air-gapped or high-security enterprises to deploy Substrate internally with zero dependencies.
- **Requirements**:
  - Use `adapter-static` in SvelteKit to pre-render the dashboard.
  - Use `//go:embed` to embed the static HTML/JS/CSS assets directly inside the Go API binary.
  - A single execution of `./substrate-server` will serve both the REST API and the Frontend UI.

### P8-T09: Docker & Helm Enterprise Delivery
- **Objective**: Standardize Kubernetes deployments for self-hosted customers.
- **Requirements**:
  - Author a multi-stage `Dockerfile` that builds the Svelte UI and Go API.
  - Author a Helm chart (`charts/substrate`) with templates for Deployments, Services, and PostgreSQL stateful sets.
