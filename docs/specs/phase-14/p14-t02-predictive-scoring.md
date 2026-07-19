# P14-T02: "Time to Break" Predictive Scoring

## 1. Objective
Transition Substrate from a purely *reactive* tool (blocking bad PRs) to a *proactive* intelligence platform. We will calculate a "Time to Break" predictive risk score (0-100) for every provider repository in the dependency graph. This score predicts the likelihood that a team will introduce a breaking change in the next N sprints, allowing Platform teams to proactively intervene.

## 2. Risk Score Heuristics
The predictive score is calculated using a weighted heuristic model based on historical data tracked by Substrate:
1. **Historical Break Frequency (Weight: 40%):** How many breaking changes has this repo attempted to merge in the last 90 days? (High attempts = high risk of one slipping through).
2. **Commit Velocity (Weight: 20%):** High commit volume on schema files (`.graphql`, `openapi.yaml`) increases the statistical probability of a break.
3. **Downstream Blast Radius (Weight: 30%):** If a repo has 50 consumers, a single mistake is catastrophic. Risk scales logarithmically with consumer count.
4. **Time Since Last Break (Weight: 10%):** Decay factor. If a team hasn't broken anything in 12 months, their risk score lowers significantly.

## 3. Architecture & Implementation
1. **Scoring Engine (`api/internal/scoring/predictive.go`):**
   - Create a background worker (River job) that runs nightly to recalculate the `predictive_risk_score` for all repositories in the database.
   - Store this score in a new `repo_metrics` table with a timestamp, allowing us to track risk trends over time (e.g., "Payments API risk score increased by 15% this week").
2. **Dashboard UI (`dashboard/src/routes/graph/`):**
   - Inject the `predictive_risk_score` into the Graph JSON payload.
   - Visually encode the nodes in the `SvelteFlow` graph:
     - Score > 80: Node glows with a pulsating red aura (High Flight Risk).
     - Score 50-79: Node is yellow.
     - Score < 50: Node is green.
3. **Weekly Digest (Slack/Email):**
   - Generate a weekly report for the `SCHEMAOWNERS` or Platform team: "Top 3 Riskiest APIs this week".

## 4. Deliverables for Jules
- Implement the `api/internal/scoring` package with the math logic.
- Create a DB migration for the `repo_metrics` table.
- Modify the `GET /api/v1/graph` endpoint to attach risk scores to nodes.
- Update the UI to render the risk auras.
