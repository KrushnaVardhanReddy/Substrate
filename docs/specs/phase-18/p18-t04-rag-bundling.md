# Phase 18: Dependency-Aware RAG Bundling (P18-T04)

## Overview
When an AI agent modifies an API schema, it often breaks downstream dependencies. This task adds a "RAG Bundling" utility to the backend. Given a target microservice, this utility will automatically bundle its upstream and downstream schemas into a single context window text block, mathematically guaranteeing cross-repo contract safety by ensuring the LLM is aware of the blast radius.

## Technical Requirements

### 1. RAG Bundling Service
Create a new service in `api/internal/services/rag_bundler.go` with a method:
`func (s *RAGBundler) GetContextBundle(ctx context.Context, org, repo string) (string, error)`

This method should:
1. Fetch the dependency graph for the specified repo using the existing `dependency_graph` package.
2. Identify all direct upstream dependencies (parents) and downstream consumers (children).
3. Fetch the current OpenAPI spec JSON for the target repo AND all identified upstreams/downstreams.
4. Concatenate these schemas into a single, well-formatted Markdown string:
   ```markdown
   # Target Service: [repo]
   ```json
   { ... }
   ```
   
   # Downstream Consumer: [child]
   ```json
   { ... }
   ```
   ```

### 2. Expose via MCP
In `api/internal/mcp/server.go`, add a new tool `get_rag_bundle`:
- Description: "Returns a bundled context string containing the target service and all its upstream/downstream API schemas to ensure contract safety when making modifications."
- Arguments: `org` (string), `repo` (string).
- It should call `rag_bundler.GetContextBundle` and return the markdown string.

## Rules
- Handle cases where the graph returns no dependencies gracefully.
- Ensure the JSON is minified to save token space in the MCP context window.
