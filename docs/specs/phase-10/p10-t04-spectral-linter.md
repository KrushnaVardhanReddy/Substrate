# Phase 10 - Task 04: AI Spectral Linter (API Governance)

## 1. Goal
Enforce plain-English API design rules (e.g. "All endpoints must use camelCase") during the PR diff process to maintain org-wide consistency.

## 2. Requirements
- Allow org admins to define natural language API rules in the Dashboard.
- Diff engine uses LLM/Spectral to evaluate OpenAPI schemas against rules.
- Block PRs that violate the governance rules.
