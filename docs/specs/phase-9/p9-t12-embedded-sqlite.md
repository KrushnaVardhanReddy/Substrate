# Spec: Embedded SQLite (LibSQL) Local Caching (P9-T12)

## 1. Overview
To ensure the Substrate CLI and MCP Server can execute schema diffs in sub-millisecond time without network latency, we will embed SQLite (using `github.com/tursodatabase/libsql-client-go`) directly into the CLI binary. 

## 2. Requirements

### 2.1 LibSQL Integration
- The CLI (`substrate`) and MCP Server (`substrate-mcp`) must initialize a local SQLite file (e.g., `~/.substrate/cache.db`) on startup.
- If the database does not exist, run initial migrations to create `schemas` and `dependencies` tables.

### 2.2 Background Sync
- A background goroutine should periodically (e.g., every 5 minutes) pull the latest graph data from the centralized Registry API (`GET /api/v1/graph`) and upsert it into the local LibSQL cache.

### 2.3 Zero-Latency Execution
- When a user runs `substrate diff` or the MCP server processes a request, it must query the local LibSQL database instead of making a network call to the Registry API.

## 3. Success Criteria
1. Running `substrate diff` executes without network calls and completes in under 50ms.
2. The `cache.db` file is correctly created and populated.
