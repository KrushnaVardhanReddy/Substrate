# CC-T03: Live VCS E2E Integration (Forgejo)

## Objective
Enhance the "No Mocks" testing paradigm by integrating a fully live Git Version Control System (VCS) into the E2E harness. This eliminates the need for simulating GitHub webhook payloads by enabling organic, end-to-end `git push` lifecycles directly against a local repository that fires authentic webhooks to Substrate.

## Constraints
1. **Dockerized Environment:** The test harness should utilize the existing `docker-compose.forgejo.yml` container to spin up Forgejo.
2. **Ephemeral State:** The Git server must store its data inside an ephemeral volume or temporary directory that is automatically purged/reset when tests complete.

## Implementation Details

### 1. E2E Runner Enhancements
Update the E2E scripts to orchestrate the container:
- Before running the Go E2E test suite, execute `docker-compose -f docker-compose.forgejo.yml up -d`.
- Ensure the container is torn down and volumes are pruned automatically at the end of the script using `docker-compose down -v`.

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
