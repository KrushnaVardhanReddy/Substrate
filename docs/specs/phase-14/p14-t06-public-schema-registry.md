# P14-T06: Substrate Cloud Public Schema Registry

## Objective
Build "The npm for APIs"—a hosted public registry within Substrate Cloud where open-source projects and SaaS companies (like Stripe, GitHub, Twilio) can publish versioned API schemas. This transforms Substrate from an internal governance tool into a global dependency management platform.

## Core Features
1. **Public Registry & Publishing:**
   Allow users to publish their schemas publicly via the CLI (`substrate publish --public`). Provide a global namespace (e.g., `@stripe/api`, `@github/rest`) accessible through the Substrate Web UI and API.

2. **Global Dependency Monitoring:**
   Enable enterprise teams to add public schemas as dependencies in their `.substrate-consumer.yaml`. If a SaaS vendor pushes a breaking change to their public schema, Substrate alerts the consuming teams immediately before the vendor's API actually breaks in production.

3. **Monetization & Tiers:**
   - **Free Tier:** Monitor up to 5 public APIs and publish unlimited open-source/public APIs.
   - **Paid Tier:** Monitor unlimited public APIs, plus publish and share private APIs across B2B partnerships.

4. **Auto-Ingestion Pipeline:**
   Create an automated cron-job/worker that pulls OpenAPI specs from popular public repositories (e.g., `github/rest-api-description`) and auto-updates the Substrate Public Registry to bootstrap the ecosystem.

## Deliverables
- `api/internal/registry/public.go`: API routes and logic for publishing, versioning, and fetching public schemas.
- Updates to `api/internal/consumers/` to support fetching upstream dependencies from the public registry (e.g., `registry.substrate.io/@stripe/api@v2`).
- A scalable cron worker in `workers/ingestion/` to automatically sync popular 3rd-party APIs.
- Migration to create the necessary `public_namespaces` and `public_schemas` tables.
