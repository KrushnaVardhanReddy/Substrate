# P10-T16: Webhook Auto-Discovery Integration

## Overview
Currently, consumer dependencies are manually mapped via a `substrate.yaml` file in the repository. The goal of this task is to integrate the AST discovery scanner developed in Phase 5 into the GitHub webhook pipeline so that dependencies are automatically detected upon `push`.

## Requirements
1. **Trigger on Push**: When a webhook push event occurs, immediately enqueue a `DiscoveryJob`.
2. **Clone and Scan**: The worker executing the job must securely pull down the repository contents using the GitHub App token and run the AST scanner over the source code.
3. **Database Upsert**: Map the detected external API calls to known contracts in the Substrate registry. Upsert the dependencies with a `confidence_score`.
4. **Fallback**: If discovery fails, log the failure but do not fail the overall push webhook workflow.

## Acceptance Criteria
- Webhook push triggers AST discovery automatically.
- Database records accurately reflect automatically discovered upstream dependencies.
