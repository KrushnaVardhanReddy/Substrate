---
title: Getting Started
---

# Substrate User Guide

Welcome to the Substrate official documentation. Substrate is a proactive CI/CD gatekeeper that uses semantic diffing, cross-repo dependency graphs, and AI intelligence to prevent schema breaking changes from ever reaching production.

---

## 1. Getting Started

### What is Substrate?
Substrate analyzes your pull requests. If you alter an API endpoint, a GraphQL type, a SQL column, or a Kafka topic, Substrate identifies every downstream repository that relies on your code. If your change breaks them, Substrate blocks your PR and generates a safe AI-powered remediation patch.

### Installation

**1. GitHub App (Recommended)**
Install the [Substrate GitHub App](https://github.com/marketplace/actions/substrate-api-contract-guard) on your organization.

**2. CLI Binary (Local Dev)**
```bash
go install github.com/KrushnaVardhanReddy/substrate/engine/cmd/substrate@latest
```

---
