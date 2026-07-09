package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server provides an MCP server implementing JSON-RPC 2.0 over stdio.
type Server struct {
	tools map[string]Tool
}

// Tool represents an exposed MCP tool.
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     func(params json.RawMessage) (any, error)
}

// NewServer creates a new MCP Server.
func NewServer() *Server {
	return &Server{
		tools: make(map[string]Tool),
	}
}

// RegisterTool adds a new tool to the MCP server.
func (s *Server) RegisterTool(tool Tool) {
	s.tools[tool.Name] = tool
}

// ServeStdio reads lines from stdin and writes responses to stdout.
func (s *Server) ServeStdio() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			continue
		}

		responseBytes := s.HandleMessage(line)
		if responseBytes != nil {
			os.Stdout.Write(responseBytes)
			os.Stdout.Write([]byte("\n"))
		}
	}
}

// HandleMessage parses a JSON-RPC request and returns the serialized JSON-RPC response.
func (s *Server) HandleMessage(line []byte) []byte {
	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return s.errorResponse(nil, -32700, "Parse error")
	}

	if req.JSONRPC != "2.0" {
		return s.errorResponse(req.ID, -32600, "Invalid Request")
	}

	var result any

	switch req.Method {
	case "initialize":
		result = map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "substrate-mcp",
				"version": "0.1.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		}
	case "tools/list":
		toolsList := []map[string]any{}
		for _, t := range s.tools {
			toolsList = append(toolsList, map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": t.InputSchema,
			})
		}
		result = map[string]any{
			"tools": toolsList,
		}
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if parseErr := json.Unmarshal(req.Params, &params); parseErr != nil {
			return s.errorResponse(req.ID, -32602, "Invalid params")
		}

		tool, exists := s.tools[params.Name]
		if !exists {
			return s.errorResponse(req.ID, -32601, "Method not found: "+params.Name)
		}

		callRes, callErr := tool.Handler(params.Arguments)
		if callErr != nil {
			result = map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": fmt.Sprintf("Error: %v", callErr),
					},
				},
				"isError": true,
			}
		} else {
			// Assume callRes is either a string or something we can marshal
			var textRes string
			switch v := callRes.(type) {
			case string:
				textRes = v
			default:
				b, _ := json.Marshal(v)
				textRes = string(b)
			}
			result = map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": textRes,
					},
				},
			}
		}
	default:
		return s.errorResponse(req.ID, -32601, "Method not found")
	}

	resultBytes, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return s.errorResponse(req.ID, -32603, "Internal error")
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resultBytes,
	}
	respBytes, _ := json.Marshal(resp)
	return respBytes
}

func (s *Server) errorResponse(id *int, code int, message string) []byte {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	respBytes, _ := json.Marshal(resp)
	return respBytes
}
