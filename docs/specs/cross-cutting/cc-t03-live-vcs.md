# CC-T03: Live VCS E2E Integration (Forgejo)

## Objective
Enhance the "No Mocks" testing paradigm by integrating a fully live Git Version Control System (VCS) into the E2E harness. This eliminates the need for simulating GitHub webhook payloads by enabling organic, end-to-end `git push` lifecycles directly against a local repository that fires authentic webhooks to Substrate.

## Constraints
1. **Strictly No Docker:** The E2E test harness (`run_full_e2e.sh`) must remain blazing fast and lightweight. Do NOT rely on `docker run`.
2. **Pre-compiled Binaries Only:** The harness should dynamically download (and cache) a pre-compiled standalone binary of [Forgejo](https://forgejo.org/) or Gitea for the current OS/Arch.
3. **Ephemeral Storage:** The Git server must store its data inside a temporary directory that is automatically purged when `run_full_e2e.sh` exits.

## Implementation Details

### 1. E2E Runner Enhancements
Update `scripts/e2e/run_full_e2e.sh`:
- Detect if the Forgejo binary exists locally. If not, download it from the official releases page.
- Spin up the Forgejo binary as a background process on a designated port (e.g., `3000`).
- Ensure the background process is properly tracked and killed on script exit alongside PGlite and SvelteKit.

### 2. E2E Setup Phase
Create a Go helper or bash script (e.g., `scripts/e2e/helpers/forgejo.go`) that executes before the test suites:
- Use the Forgejo REST API to create a mock user and organization (e.g., `mcp-org`).
- Initialize a blank repository (`enterprise-repo`).
- Configure a webhook on the repository pointing to `http://localhost:8090/api/v1/webhook` (Substrate API).

### 3. Test Assertions
When tests require an external webhook event:
- Clone the local Forgejo repository.
- Write/Commit changes to the OpenAPI spec.
- Run `git push`.
- Assert that Substrate receives and processes the webhook successfully.
