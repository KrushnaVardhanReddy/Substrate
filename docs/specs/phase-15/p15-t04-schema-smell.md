# P15-T04: "Schema Smell" Detector (AI API Design Linter)

## Overview
Proactively detect API design anti-patterns beyond breaking changes (e.g., over-fat endpoints, non-descriptive field names, duplicated response objects without `$ref`). Scores APIs 0–100.

## Requirements
1. **AI Linter Engine**: Create a new package `engine/linter` that passes the OpenAPI spec to the LLM (using `STITCH_API_KEY` or `JULES_API_KEY`) and evaluates it against standard REST anti-patterns.
2. **Heuristics**:
   - Endpoints with >10 parameters.
   - Non-descriptive field names (e.g. `data1`, `flag`).
   - Missing `$ref` reuse for identical structures.
3. **Scoring**: Output a score out of 100 with actionable feedback.
4. **CLI Integration**: Add a command `substrate lint --schema <path>` to run the smell detector and print the report.
