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
- **AI Delegation Pipeline**: Substrate is built using a 3-tier AI factory: Local LLMs (Gemma via OpenCode) for specification architecture, Cloud Agents (Jules) for implementation, and IDE Agents (Antigravity) for code review/debugging.

## Svelte 5 & UI Guidelines
When generating specifications or prompt instructions for UI tasks, explicitly enforce these rules:
- **Runes Mode**: Substrate uses Svelte 5. State must be handled via runes (`$state`, `$derived`, `$effect`).
- **Lucide Icons**: Due to Svelte 5 prop restrictions, **never use barrel imports** for icons. Always use direct imports: `import Book from 'lucide-svelte/icons/book'` instead of `import { Book } from 'lucide-svelte'` to prevent `CompileError: Cannot use $$props in runes mode`.

Always read `docs/wiki/index.md` before answering complex architectural questions to ground your knowledge.
