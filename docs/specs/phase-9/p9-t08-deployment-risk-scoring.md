# P9-T08: Deployment Risk Scoring

## Overview
Synthesize breaking change data, infrastructure changes, and downstream blast radius into a holistic "Deployment Risk Score" surfaced in GitHub PR comments.

## Requirements
1. **Score Inputs**: Combine (a) number of breaking changes, (b) blast radius node count, (c) presence of DB migrations, (d) whether E2E tests pass.
2. **Score Output**: Produce a composite score: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`.
3. **PR Comment Integration**: Display the score prominently at the top of the GitHub PR comment (e.g., `🔴 DEPLOYMENT RISK: HIGH`).
4. **API Endpoint**: Expose `GET /api/v1/risk/:org/:repo/:pr` returning the score as JSON.
