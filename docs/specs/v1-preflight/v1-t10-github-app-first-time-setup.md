# V1-T10: GitHub App First-Time Setup Support

## Objective
When a user installs the Substrate GitHub App and adds a `substrate.yaml` configuration alongside a brand new schema file in their very first Pull Request, there is no previous "base" schema on the `main` branch to diff against. 
Currently, the GitHub App throws a generic "Config Error" because it fails to fetch the baseline. This creates a terrible first-time user experience. We need to explicitly handle this scenario to welcome new users properly.

## Requirements

### 1. Differentiate Fetch Failures
- Target: `github-app/src/index.ts`
- Separate the logic that checks if `config.base_schema` is present from the logic that actually fetches the files.
- If `baseContent` is `null` but `headContent` is present, it means the file is brand new to the repository (first-time setup).
- If both are `null`, or `headContent` is `null`, it means the file paths are genuinely incorrect.

### 2. Welcome Comment
- If the first-time setup condition is met, post a specific PR comment:
  `## 🎉 Welcome to Substrate!`
  `We detected your new schema file. Since this is your first time adding it, there is no previous baseline to compare against.`
  `Once this PR is merged, Substrate will begin monitoring all future pull requests for breaking changes!`

### 3. Commit Status Override
- Do NOT set a `failure` commit status.
- Instead, set the GitHub Commit Status to `success` with the description: `First-time setup detected — Ready to merge!`.
- Return a `200 OK` response to GitHub so the webhook delivery is marked as successful.
