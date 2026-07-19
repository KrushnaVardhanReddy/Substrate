# P9-T07: AI Migration Planner

## Objective
When a developer introduces a complex breaking change (e.g., splitting a `name` field into `first_name` and `last_name`), they often lack the operational context to safely migrate the data and the API simultaneously without downtime. The AI Migration Planner generates a multi-step, zero-downtime migration strategy explicitly tailored to the breaking change detected.

## Core Features
1. **Change Contextualization:**
   The diff engine detects the breaking change (e.g., `FIELD_REMOVED: name`, `FIELD_ADDED: first_name`, `FIELD_ADDED: last_name`).
2. **Strategy Generation (LLM):**
   Substrate passes the diff context to an LLM (OpenAI/Anthropic) to generate an Expand-and-Contract deployment strategy.
   - **Step 1 (Expand):** Add the new fields (`first_name`, `last_name`), but keep `name` as an optional field. 
   - **Step 2 (Migrate):** Write a script/query to backfill data from `name` to the new fields.
   - **Step 3 (Dual Write):** Application logic writes to both old and new fields.
   - **Step 4 (Contract):** Once all consumers are updated (checked via Substrate registry), drop the `name` field.
3. **PR Comment Integration:**
   The generated plan is injected into the GitHub PR as an actionable checklist for the developer.

## Deliverables
- `engine/pkg/ai/migration.go`: The LLM prompt builder and integration.
- `api/internal/handlers/diff.go`: Logic to trigger the migration planner if the breaking change severity is high.
- Formatting logic to output the response nicely in the existing GitHub PR comment bot.
