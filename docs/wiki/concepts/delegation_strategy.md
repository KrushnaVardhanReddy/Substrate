# Substrate Three-Tier Delegation Strategy

The Substrate project utilizes a strict three-tier AI delegation strategy to maximize development velocity while minimizing API token burn.

## 1. OpenCode + Local LLM (Gemma 4 12B QAT) - The Architect
- **Role:** Generates extensive boilerplate specification markdowns (`docs/specs/...`) and Jules prompts (`prompts/...`).
- **Cost:** Zero API cost. Runs locally.
- **Why:** Gemma 4 12B QAT is highly tuned for instruction following and runs extremely fast locally, making it perfect for generating boilerplate specs that match the existing project templates based on `tasks.md`.

## 2. Jules (Cloud Agents) - The Factory
- **Role:** Reads the local-generated specs and prompts, implements the actual code (handlers, DB migrations, Svelte components), writes unit tests, and submits a Pull Request.
- **Environment:** Runs in an isolated cloud container.
- **Invocation:** Triggered via `python3 scripts/jules_submit.py --file <prompt_file>`.

## 3. Antigravity (IDE Agent) - The Code Reviewer & Debugger
- **Role:** The overarching manager that stays within the user's IDE.
- **Responsibilities:**
  - Manages merge conflicts between parallel Jules PRs.
  - Fixes complex E2E Playwright test breakages.
  - Reviews code, updates tracking files (`tasks.md`, `handoff.md`), and serves as the final merge authority.
