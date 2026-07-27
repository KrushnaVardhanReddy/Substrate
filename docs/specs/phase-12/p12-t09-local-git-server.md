# P12-T09: Local Git Server Integration (Forgejo)

## Objective
Establish a real local Git server environment using Forgejo (a lightweight, community-driven Gitea fork) to simulate live GitHub webhooks. This transition moves us away from static JSON mock payloads, providing a high-fidelity testing environment for the Substrate Cloudflare Worker, Go backend, and Svelte UI.

## Rationale
While testing with `mock_pr_payload.json` via OpenCode was sufficient for initial API endpoint validation, it lacks the fidelity of a real Git workflow. 

By integrating a real Git server locally:
1. **True E2E Validation:** We test the actual receipt of HTTP requests, payload signing, and event routing exactly as it will happen in production.
2. **Behavioral Accuracy:** Real `git push` operations fire a cascade of events (e.g., `push` and `pull_request`) allowing us to test our deduplication and processing queues.
3. **No External Dependencies:** We can fully test our GitHub App integration without exposing our local development environment to the public internet via `ngrok` or relying on actual GitHub.com infrastructure.
4. **Compatibility:** Forgejo is API and Webhook-compatible with GitHub, making it a perfect drop-in replacement for local testing.

## Architecture

1. **Git Server:** Forgejo running locally in a Docker container (exposed on port `3000` for HTTP and `2222` for SSH).
2. **Webhook Target:** The Forgejo repository will be configured to send webhooks to the locally running Substrate Cloudflare Worker (`http://host.docker.internal:8787`).
3. **API & UI:** The worker processes the event and communicates with the local Go backend, which updates the Postgres DB, triggering SSE events to dynamically update the Cytoscape Graph UI.

## Implementation Steps

1. **Docker Compose Setup:** Create a `docker-compose.forgejo.yml` file to spin up the Forgejo container with persistent SQLite storage. *(Completed)*
2. **Initial Configuration:** Boot the container, navigate to `http://localhost:3005`, and initialize the default administrator account.
3. **Repository Setup:** 
   - Create a repository (e.g., `microservices-demo`).
   - Add the local Forgejo repository as a remote to the local `demo-repos/microservices-demo` directory.
4. **Webhook Configuration:**
   - In Forgejo Repository Settings, create a "Gitea" or "Forgejo" webhook (which shares the GitHub structure).
   - Set the URL to the local Wrangler worker (`http://host.docker.internal:8787`).
   - Select events: `Push` and `Pull Request`.
5. **Validation:** 
   - Perform a `git push` to Forgejo.
   - Verify the worker intercepts the event.
   - Verify the blast radius analytics run and the graph updates seamlessly.
