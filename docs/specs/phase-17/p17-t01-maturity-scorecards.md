# Phase 17 - Task 01: API Maturity Scorecards

## 1. Goal
Provide Engineering Managers with a high-level "API Health Grade" (A–F) for each registered repository, based on an automated scoring engine that evaluates documentation quality, compliance risk, volatility, and QA test coverage.

## 2. Scoring Model
Each metric contributes up to 25 points (total 100):
| Metric | Source | Max Points |
|---|---|---|
| Documentation completeness | % of endpoints with non-empty `description` in OpenAPI spec | 25 |
| Compliance risk | Deduct 5pts per active PII/SOC2 warning in `compliance_flags` | 25 |
| Volatility | Deduct 3pts per breaking change in the last 30 days | 25 |
| QA shadow coverage | % of endpoints covered in `shadow_coverage` table | 25 |

Grade thresholds: A=90+, B=75+, C=60+, D=40+, F=<40.

## 3. Requirements
- Add `GET /api/v1/scorecard/{org}/{repo}` — computes and returns the score JSON: `{ grade, total, breakdown: { docs, compliance, volatility, coverage } }`.
- Persist results in a `scorecard_cache` table with a 1-hour TTL to avoid re-computation on every request.
- Add a new "Health" column to the Org Overview page (`dashboard/src/routes/(app)/org/[org]/+page.svelte`) showing the grade badge for each repo.

## 4. Database Schema
```sql
CREATE TABLE scorecard_cache (
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  grade       TEXT NOT NULL,
  total       INT NOT NULL,
  breakdown   JSONB NOT NULL,
  computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (org, repo)
);
```
