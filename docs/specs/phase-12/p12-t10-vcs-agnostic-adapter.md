# P12-T10: VCS-Agnostic Webhook & API Adapter Architecture

## Objective
Decouple Substrate from GitHub-specific APIs and webhook structures. Implement a generic Version Control System (VCS) adapter pattern that allows Substrate to seamlessly support GitHub, GitLab, Bitbucket, Gitea/Forgejo, and on-premise enterprise git solutions.

## Rationale
Enterprise environments rarely standardize on a single Git provider. While GitHub is popular, many large organizations mandate self-hosted solutions like GitLab Enterprise, Bitbucket Data Center, or Gitea/Forgejo. 

Currently, our Cloudflare worker (`github-app/src`) tightly couples webhook parsing and API calls to `api.github.com` and GitHub App authentication mechanics. Making this generic unlocks a much larger enterprise market and instantly solves our local Gitea testing limitation.

## Proposed Architecture: The Adapter Pattern

We will introduce a standard interface `VCSProvider` that abstracts both **Webhook Ingestion** and **API Execution**. 

### 1. Unified Event Interfaces
Instead of passing around `X-GitHub-Event`, we will map incoming webhooks to standard internal events:
- `StandardPushEvent` (branch, commit_sha, owner, repo)
- `StandardPullRequestEvent` (pr_number, head_sha, base_branch)

### 2. VCS Client Interface
Create an interface that all providers (GitHub, GitLab, Gitea) must implement:
```typescript
export interface VCSClient {
  // Authentication
  authenticate(): Promise<string>; 

  // File Operations
  fetchFileContent(owner: string, repo: string, path: string, ref: string): Promise<string | null>;
  
  // PR Operations
  postComment(owner: string, repo: string, prId: number, body: string): Promise<void>;
  setCommitStatus(owner: string, repo: string, sha: string, state: 'success'|'failure'|'pending', description: string): Promise<void>;
}
```

### 3. Request Router (Factory Pattern)
The main worker `fetch` loop will inspect incoming headers to determine the VCS provider:
- `X-GitHub-Event` -> Instantiate `GitHubProvider`
- `X-Gitlab-Event` -> Instantiate `GitLabProvider`
- `X-Gitea-Event` / `X-Forgejo-Event` -> Instantiate `GiteaProvider`

### 4. Configuration Driven
We will add `VCS_API_BASE_URL` to the `.dev.vars` / Wrangler environment. 
- If the provider is Gitea, it uses `VCS_API_BASE_URL` (e.g., `http://localhost:3000/api/v1`).
- If the provider is GitHub and `VCS_API_BASE_URL` is empty, it defaults to `https://api.github.com`.

## Implementation Steps (Phase 1)
To achieve this without breaking current functionality:

1. **Refactor Webhook Parsing:** Move GitHub-specific parsing into `src/providers/github/webhook.ts`.
2. **Create Gitea Provider:** Implement `src/providers/gitea/client.ts` implementing the `VCSClient` interface, using basic token auth.
3. **Update Worker Entrypoint:** Modify `index.ts` to route the incoming request to the correct provider factory based on headers.
4. **Environment Updates:** Inject `GITEA_API_URL` and `GITEA_TOKEN` into the local worker environment.

## Definition of Done
- The Substrate worker can process a push event from Forgejo/Gitea.
- The worker successfully fetches `substrate.yaml` from the local Gitea API.
- The Go backend receives the dependency updates and the Svelte dashboard updates visually.
