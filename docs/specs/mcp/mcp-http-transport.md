# MCP HTTP SSE Transport

## Overview
The Model Context Protocol (MCP) server currently supports a headless Stdio transport (`ServeStdio`), which works for CLI environments. However, to expose the MCP server over a network (e.g., allowing cloud AI agents to connect to the Substrate Registry API), we need to implement the HTTP Server-Sent Events (SSE) transport.

## Architecture
The HTTP SSE transport consists of two endpoints:
1. **SSE Connection Endpoint (`GET /mcp/sse`)**
   - The client connects to this endpoint and expects an SSE stream (`Content-Type: text/event-stream`).
   - The server accepts the connection, generates a unique session ID (e.g., UUID), and sends an initial `endpoint` event:
     ```
     event: endpoint
     data: /mcp/messages?sessionId=<uuid>
     
     ```
   - The server holds the connection open, waiting for JSON-RPC responses to send back to the client.

2. **Message Endpoint (`POST /mcp/messages?sessionId=<uuid>`)**
   - The client sends JSON-RPC requests (e.g., `tools/call`, `resources/list`) to this endpoint via POST.
   - The server routes the JSON payload to the `mcpServer.HandleMessage(payload)`.
   - The server immediately responds to the POST request with HTTP 202 Accepted.
   - The server routes the JSON-RPC response back to the client via the SSE connection mapped to the `sessionId`.

## Security
- Both the `/mcp/sse` and `/mcp/messages` endpoints must be protected by the `ServiceTokenMiddleware` to ensure only authorized AI agents (or the dashboard) can communicate with the MCP server.

## Integration
- The new `HTTPTransport` struct will be added to the `api/internal/mcp` package.
- The routes will be attached to the main multiplexer (`chi.Router`) in `api/cmd/server/main.go`.
