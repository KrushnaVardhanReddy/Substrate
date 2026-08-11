// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"encoding/json"
	"os"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func main() {
	d := json.NewDecoder(os.Stdin)
	e := json.NewEncoder(os.Stdout)

	for {
		var req JSONRPCRequest
		if err := d.Decode(&req); err != nil {
			break
		}

		if req.Method == "initialize" {
			initResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"protocolVersion": "2024-11-05",
					"serverInfo": {"name": "test-mcp", "version": "1.0"},
					"capabilities": {"tools": {}}
				}`),
			}
			e.Encode(initResp)
		} else if req.Method == "notifications/initialized" {
			// Do nothing
		} else if req.Method == "tools/list" {
			toolsResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"tools": [
						{
							"name": "test_tool",
							"inputSchema": {
								"type": "object",
								"properties": {
									"arg1": {"type": "string"}
								}
							}
						}
					]
				}`),
			}
			e.Encode(toolsResp)
		} else if req.Method == "resources/list" {
			resResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"resources": [
						{
							"uri": "file:///test",
							"name": "test_res",
							"mimeType": "application/json"
						}
					]
				}`),
			}
			e.Encode(resResp)
		} else if req.Method == "prompts/list" {
			promptResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"prompts": [
						{
							"name": "test_prompt",
							"arguments": [
								{
									"name": "arg1",
									"required": true
								}
							]
						}
					]
				}`),
			}
			e.Encode(promptResp)
			break // exit after prompts/list to finish the command gracefully
		}
	}
}
