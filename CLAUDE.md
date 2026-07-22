# Substrate AI Agent Guidelines & Wiki Schema

You are acting as an AI developer and knowledge base maintainer for the Substrate project.

## The LLM-Wiki Pattern
We use a persistent, compounding knowledge base located in `docs/wiki/`. 
- **Raw Sources**: `docs/specs/`, `prompts/`, and `tasks.md`. Do not modify these unless explicitly instructed.
- **The Wiki**: `docs/wiki/`. You own this layer. Create, update, and cross-reference Markdown files here to synthesize knowledge.
- **Index**: `docs/wiki/index.md` is the content catalog. Always update it when adding new pages.
- **Log**: `docs/wiki/log.md` is an append-only timeline of ingests and updates. Use the format `## [YYYY-MM-DD] action | Description`.

## Substrate Architectural Pillars
- **Hexagonal Architecture**: Handlers (ports) are strictly separated from DB implementations (adapters) using `sqlc`.
- **E2E Testing (No Mocks)**: Substrate relies heavily on live integration testing in `scripts/e2e/`. Tests expect a real Postgres DB.
- **Enterprise Features**: Features like KMS BYOK, LLM routing, CRM Blast Radius, and Schema Insurance are core to the enterprise offering.

Always read `docs/wiki/index.md` before answering complex architectural questions to ground your knowledge.
