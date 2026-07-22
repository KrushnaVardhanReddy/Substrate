# P13-T03: AI Chaos Engineering Auto-Tests

## Overview
Substrate aims to automatically prove that an API breakage is real by dynamically generating a Playwright/Jest test and running it against a mock server based on the previous schema version. This acts as "AI Chaos Engineering."

## Requirements
1. **Dynamic Test Generation**: Hook into the `AIAnalyzeHandler` to generate JavaScript E2E test code that issues a request highlighting the missing/modified field.
2. **Sandbox Execution**: Spin up an isolated JavaScript runtime sandbox (using `deno` or an isolated v8 context) to execute the generated test code.
3. **Mock Upstream Integration**: Start an ephemeral mock server representing the downstream consumer's expected state.
4. **Failure Proof Logs**: Collect the failing test output and append it to the GitHub PR comment as definitive proof of breakage.

## Acceptance Criteria
- AI successfully generates runnable test scripts.
- The tests execute securely in a sandbox and fail as expected, validating the blast radius.
