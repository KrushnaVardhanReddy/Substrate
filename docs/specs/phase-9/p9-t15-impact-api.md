# Spec: Public Impact API & MCP Tool (P9-T15)

## 1. Overview
As Substrate moves towards becoming a programmatic infrastructure platform, it is critical that CI/CD pipelines and AI Agents can query the "Blast Radius" of a proposed change without needing to view the Svelte dashboard. 

This spec introduces a new Go REST endpoint (`/api/v1/impact`) and a new MCP Tool (`get_blast_radius`).

## 2. Requirements

### 2.1 Impact API (`/api/v1/impact`)
- **Route:** `POST /api/v1/impact` (or `GET /api/v1/impact/{org}/{repo}` if we only need repo-level blast radius without a specific schema diff). Let's implement `GET /api/v1/impact/{org}/{repo}` to calculate the cascading downstream blast radius for a given provider repo.
- **Handler Logic:** 
  - Utilize the existing graph/registry logic to find all downstream consumers (and their consumers, recursively).
  - Calculate a `risk_score` (e.g., number of downstream repos impacted).
- **Response Format:**
  ```json
  {
    "provider": "myorg/payments-api",
    "risk_score": 3,
    "impacted_repos": [
      "myorg/checkout-ui",
      "myorg/invoice-worker",
      "myorg/reporting-dashboard"
    ]
  }
  ```

### 2.2 MCP Tool (`get_blast_radius`)
- Update `engine/cmd/substrate-mcp/main.go` to register a new tool called `get_blast_radius`.
- **Description:** "Calculates the downstream blast radius and risk score for a repository. Use this to determine how many other teams you will break before making a change."
- **Input Schema:**
  ```json
  {
    "repo": {
      "type": "string",
      "description": "The provider repository (e.g., 'myorg/payments-api')"
    }
  }
  ```
- **Handler Logic:**
  - Call the newly created `GET /api/v1/impact/{org}/{repo}` endpoint on the Registry API.
  - Return the JSON result to the AI agent.

## 3. Success Criteria
1. The Go Registry API correctly exposes `/api/v1/impact/{org}/{repo}` and accurately computes the Nth-degree blast radius.
2. The MCP Server successfully registers `get_blast_radius` and can query the API.
3. Unit tests are provided for the API handler.
