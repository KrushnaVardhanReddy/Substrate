# P15-T01: Schema Review Assignments ("CODEOWNERS for APIs")

## Objective
In enterprise environments, the teams that build the APIs (engineers) are rarely the teams that govern the APIs (API Platform / Governance / Architecture teams). GitHub's standard `CODEOWNERS` system fails here because it assigns PR reviews based on file paths, not API semantics. Substrate will introduce `SCHEMAOWNERS` to automatically request reviews from designated API architects whenever a schema is touched, regardless of where the code lives.

## Core Features
1. **The `SCHEMAOWNERS` Manifest:**
   A YAML file placed in the repository root (`.substrate/SCHEMAOWNERS.yaml`):
   ```yaml
   # Require the @api-platform team to review any changes to the payments API
   - path: "/v1/payments/*"
     reviewers:
       - "@myorg/api-platform"
   # Require the security team to review if auth headers change
   - rules:
       - SECURITY_HEADER_MODIFIED
     reviewers:
       - "@myorg/security-team"
   ```
2. **GitHub Review Automation:**
   If the Diff Engine detects changes that match the paths or rules defined in the `SCHEMAOWNERS` file, Substrate makes an API call to GitHub (`POST /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers`) to automatically add the required teams.

## Deliverables
- `api/internal/governance/owners.go`: YAML parser and matching engine.
- Update `api/internal/github/client.go` to support adding PR reviewers.
- E2E test verifying that a detected path change successfully requests a review from a mocked GitHub team.
