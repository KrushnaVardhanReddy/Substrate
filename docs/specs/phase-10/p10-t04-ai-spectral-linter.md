# P10-T04: AI Spectral Linter (API Governance)

## Overview
Enforce plain-English API design rules (e.g., "All endpoints must use camelCase") during the PR diff process to maintain org-wide consistency.

## Requirements
1. **Rule Storage**: Store custom governance rules as plain-text strings in the database, linked to an org.
2. **AI Evaluation**: Pass the OpenAPI diff and governance rules to the LLM during PR checks. Ask it to flag any violations.
3. **PR Comment Section**: Append a "📏 API Governance" section in the PR comment listing any violations.
4. **Rule Management UI**: Add a simple "Governance Rules" CRUD page in the dashboard settings.
