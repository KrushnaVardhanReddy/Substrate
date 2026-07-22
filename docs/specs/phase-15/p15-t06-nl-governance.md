# P15-T06: Natural Language Governance Rules

## Overview
Extend the Custom Rules Engine (CEL) with a plain-English interface. Platform teams type rules like "All payment APIs must require authentication" and Substrate's AI auto-generates the CEL rule with a preview before saving.

## Requirements
1. **AI Translation Layer**: Add an endpoint in the API server `POST /api/governance/generate-cel` that takes natural language and returns a valid CEL expression.
2. **Validation**: The generated CEL must be strictly validated against the API schemaAST before being returned.
3. **UI Integration**: Add a natural language text box in the Dashboard's Governance Rules section that fetches the CEL translation.
