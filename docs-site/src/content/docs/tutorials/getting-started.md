---
title: Getting Started
description: What is Substrate and how to install it.
---

### What is Substrate?
Substrate analyzes your pull requests. If you alter an API endpoint, a GraphQL type, a SQL column, or a Kafka topic, Substrate identifies every downstream repository that relies on your code. If your change breaks them, Substrate blocks your PR and generates a safe AI-powered remediation patch.

### Installation

**1. GitHub App (Recommended)**
Install the [Substrate GitHub App](https://github.com/marketplace/actions/substrate-api-contract-guard) on your organization.

**2. CLI Binary (Local Dev)**
```bash
go install github.com/KrushnaVardhanReddy/substrate/engine/cmd/substrate@latest
```