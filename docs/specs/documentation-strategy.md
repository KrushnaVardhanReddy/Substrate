# Substrate Documentation Architecture

> **Status:** LIVING DOCUMENT
> **Purpose:** Defines the standard and structure for Substrate's official documentation. This must be the gold standard for both human readability and AI/MCP ingestion.

## Reference Models
We are modeling our documentation strictly after:
1. **Buf (`buf.build`)** — For CLI and YAML configuration reference structure.
2. **Vercel / Next.js** — For "Copy-Paste" CI/CD integration Quickstarts.
3. **Pact (`pact.io`)** — For conceptual explanations of consumer-driven cross-repo contracts.

## Core Structure Blueprint
When executing task **P3-T12**, the documentation site (built with VitePress or Starlight) must follow this exact structure:

### 1. Getting Started
- **What is Substrate?** (Concepts, Architecture, & Mermaid Diagrams)
- **Installation** (CLI binary, Homebrew, Docker)
- **CI/CD Quickstarts** (GitHub Actions, GitLab CI, CircleCI copy-paste YAMLs)

### 2. Configuration Reference
- **The `substrate.yaml` file** (Every parameter, type, required status, and default value)
- **Overrides** (How to acknowledge intentional breaking changes via the `overrides` block)

### 3. CLI Reference
- `substrate diff` (Base vs Head validation)
- `substrate serve` (HTTP mode for the GitHub App)
- `substrate init` (Including `--design` and `--ai` flags)

### 4. Contract Registry (Phase 3)
- **Concepts:** The Push-to-Main sync webhook vs. the PR Cross-Repo check.
- **Connecting your GitHub Org**
- **The Compatibility Matrix explained**

### 5. Rule Reference (The "Linter" Rules)
*Every single rule must have its own section with `Code`, `Description`, and `Recommendation`.*
- **OpenAPI Rules** (All 37 semantic breaking change rules)
- **Protobuf Rules** (`buf breaking` mappings)
- **SQL Rules** (PostgreSQL DDL breaking changes)
- **AsyncAPI & Avro Rules**

## AI / MCP Guidelines
Because this documentation will be fed directly to LLMs via our MCP server:
- Use consistent markdown heading levels (`#`, `##`, `###`).
- Favor tables for configurations and rule lists over prose.
- Include literal string examples for all JSON/YAML structures.
- Keep the tone direct, technical, and free of marketing fluff.
