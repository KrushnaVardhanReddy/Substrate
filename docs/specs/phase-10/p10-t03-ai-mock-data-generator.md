# P10-T03: AI Mock Data Generator (QA)

## Overview
Scan QA repositories for JSON test fixtures and use the AI engine to auto-update mock data when the upstream API schema changes.

## Requirements
1. **Fixture Discovery**: Scan a QA repo's `__fixtures__` or `testdata/` directories for JSON files.
2. **Schema Alignment**: Compare the fixture structure to the updated OpenAPI schema using the existing diff engine.
3. **AI Auto-Update**: Pass mismatches to the LLM with a prompt to regenerate valid mock data that matches the new schema.
4. **PR Creation**: Open a PR in the QA repo with the updated fixtures.
