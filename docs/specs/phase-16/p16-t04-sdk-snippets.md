# Phase 16: Auto-Generated SDK Code Snippets (P16-T04)

## Overview
Substrate aims to completely replace API documentation tools like ReadMe and Backstage. A core feature of any developer portal is providing copy-pasteable request snippets for endpoints. 
This task introduces a snippet generation utility on the frontend (SvelteKit) that reads the OpenAPI schema for a given endpoint and generates request examples in multiple languages (Curl, Python, Node.js, Go).

## Technical Requirements

### 1. Snippet Generator Utility
Create a new utility in `dashboard/src/lib/utils/snippetGenerator.ts`.
It should export a function like `generateSnippets(method: string, path: string, headers: any, queryParams: any, body: any)` returning a map of languages to formatted strings.
Supported languages:
- **cURL**: Standard bash curl command.
- **Python (Requests)**: Using the `requests` library.
- **Node.js (Fetch)**: Using standard ES6 `fetch`.
- **Go (net/http)**: Using the standard library.

### 2. Svelte Component
Create a `CodeSnippetViewer.svelte` component in `dashboard/src/lib/components/` that uses a tabbed interface (similar to the interactive sandbox or diff viewer) allowing the user to select their preferred language and copy the snippet.

### 3. Integration
Integrate `CodeSnippetViewer.svelte` into the main API catalog/documentation page (e.g., in `dashboard/src/routes/(app)/org/[org]/catalog/[repo]/page.svelte` or wherever endpoints are rendered).
Pass down the relevant OpenAPI endpoint details to generate the snippets on the fly.

## Rules
- You do NOT need a backend service for this. This should be purely client-side parsing of the OpenAPI operation object.
- Keep the generated snippets as idiomatic as possible.
- Include a "Copy to Clipboard" button on the snippet viewer.
