# P9-T11: Automated Deprecation Campaigns

## Overview
Track sunsetting endpoints, auto-open issues in downstream consumer repos, and nag them until 0% usage is reached.

## Requirements
1. **Deprecation Marking**: Support marking endpoints as deprecated in `substrate.yaml` with a `sunset_date`.
2. **Consumer Discovery**: Use the existing blast radius graph to find all downstream repos consuming the deprecated endpoint.
3. **Auto-Open Issues**: Use the GitHub API to open a standardized deprecation notice issue in each consumer repo.
4. **Usage Polling**: Periodically re-check if consumers have migrated away (by re-running the spec diff). Close the issue automatically when usage drops to 0%.
