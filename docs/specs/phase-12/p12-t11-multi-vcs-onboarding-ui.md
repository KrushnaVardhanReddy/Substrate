# P12-T11: Multi-VCS Onboarding UI

## Objective
Update the Substrate dashboard onboarding flow to natively support multiple Git providers (GitHub, GitLab, and Self-Hosted/Gitea), aligning with the VCS-Agnostic Webhook backend (P12-T10).

## Current State
The onboarding wizard at `dashboard/src/routes/onboarding/+page.svelte` currently assumes a hardcoded GitHub integration. It only asks the user for a GitHub Personal Access Token before scanning repositories.

## Proposed UX Flow

### Step 1: Provider Selection
Before asking for a token, present the user with a choice of VCS providers:
- **GitHub** (Icon: GitHub logo) - "Connect to github.com"
- **GitLab** (Icon: GitLab logo) - "Connect to gitlab.com"
- **Self-Hosted** (Icon: Server/Database icon) - "Connect to Gitea, Forgejo, or Enterprise Server"

### Step 2: Credential Entry
Based on the selection in Step 1, the credential form should adapt:

#### If GitHub or GitLab:
- Input 1: **Personal Access Token** (Placeholder: "ghp_..." or "glpat-...")
- *The Base URL is hidden and defaults to `https://api.github.com` or `https://gitlab.com/api/v4` under the hood.*

#### If Self-Hosted:
- Input 1: **Server Base URL** (Placeholder: "http://localhost:3005/api/v1")
- Input 2: **Personal Access Token** (Placeholder: "Your server token")

### Step 3: Connection & Scanning
When the user clicks "Connect", the frontend payload sent to the backend must include the new fields:
```json
{
  "provider": "github" | "gitlab" | "custom",
  "base_url": "https://...",
  "token": "..."
}
```
*Note: Ensure the SvelteKit API endpoint (`dashboard/src/routes/api/v1/repos/+server.ts` or similar) passes these fields through to the Go backend.*

## Design Requirements (Premium Aesthetics)
- Follow the Phase 11 aesthetic guidelines (Dark Mode, Glassmorphism).
- The provider selection should use large, clickable "Cards" that glow on hover.
- Use smooth Svelte transitions (`fade` or `slide`) between Step 1 and Step 2.

## Definition of Done
- The onboarding wizard successfully collects the provider type, URL, and token.
- The UI gracefully handles validation (e.g., ensuring Base URL is a valid HTTP/HTTPS string when "Self-Hosted" is selected).
- The payload is correctly structured and sent to the backend API.
