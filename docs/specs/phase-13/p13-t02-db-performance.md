# P13-T02: DB Performance Breakages (CLI Analytics)

## Objective
Introduce advanced performance analytics to the Substrate CLI and Engine to detect when database schema changes (e.g., missing indices, large table locks, suboptimal data types) will introduce performance regressions, even if they aren't strictly API-breaking.

## Context
Currently, Substrate excels at structural breaking changes (e.g., dropping a column). However, adding a complex view or modifying a high-traffic table without appropriate indices can cause catastrophic performance issues in production. This task adds "Performance Breakages" as a new category of impact analysis.

## Requirements

### 1. SQL Schema Analysis
- Enhance the SQL AST parser to analyze `CREATE INDEX`, `ALTER TABLE`, and query patterns.
- Implement heuristic rules to detect performance risks, such as:
  - Adding a column with a default value to a large table (table rewrite risk).
  - Changing column types that require a full table scan.
  - Creating queries that lack corresponding indices.

### 2. CLI Integration
- Extend the `substrate analyze` CLI command to report performance warnings.
- Output warnings clearly: "Warning: Altering column X on table Y may cause a prolonged lock."

### 3. Substrate Dashboard Integration
- Surface these performance warnings in the Substrate PR bot comments and the UI Impact Analysis view, categorized under "Performance Risks".

## Acceptance Criteria
1. The Go Engine's SQL parser can identify a missing index for a newly defined foreign key.
2. The CLI outputs a distinct "Performance Risk" warning when a high-risk migration is analyzed.
3. The PR integration includes performance risks in its standard blast radius report.
