# P7-T04: Cross-Repo Auto-Fix PRs

## Objective
When a provider repository merges a breaking change that breaks a downstream consumer, the burden of updating the consumer usually falls on the provider team. 
Substrate will leverage its existing AI Autofix engine to automatically generate a draft Pull Request in the downstream consumer repository that fixes their implementation to comply with the new upstream schema.

## Requirements

### 1. Cross-Repo Patch Generation
- Target: `api/internal/handlers/ai_autofix.go`
- Extend the AI prompt to accept both the upstream breaking change diff AND the downstream consumer's source code (fetched via GitHub API).
- Instruct the AI to generate a patch for the downstream consumer code (e.g., removing the usage of a deleted field).

### 2. Automated Draft PR
- Target: `api/internal/github/client.go`
- Implement a method `CreateDraftPR(owner, repo, branch, patch, title, body)`.
- If a breaking change is merged (detected via `POST /api/v1/webhook` push event), Substrate automatically branches off `main` in the consumer repo, applies the AI-generated patch, and opens a Draft PR titled:
  `chore(substrate): Auto-fix breaking change from upstream [UpstreamRepo]`

### 3. Security Guardrails
- These PRs must ALWAYS be opened as Drafts.
- No auto-merging is permitted.
- The PR description must clearly outline the AI's reasoning so the human reviewers can validate the logic.
