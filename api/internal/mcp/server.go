package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name          string
	Description   string
	InputSchema   map[string]any
	Handler       func(params json.RawMessage) (any, error)
	IsDestructive bool
}

type Resource struct {
	URI         string
	Name        string
	Description string
	MimeType    string
	Handler     func(uri string) (string, error)
}

type Prompt struct {
	Name        string
	Description string
	Handler     func() (string, error)
}

type Server struct {
	tools     map[string]Tool
	resources map[string]Resource
	prompts   map[string]Prompt
	store     db.Store
}

func NewServer(store db.Store) *Server {
	return &Server{
		tools:     make(map[string]Tool),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
		store:     store,
	}
}

func (s *Server) RegisterTool(tool Tool) {
	s.tools[tool.Name] = tool
}

func (s *Server) RegisterResource(resource Resource) {
	s.resources[resource.URI] = resource
}

func (s *Server) RegisterPrompt(prompt Prompt) {
	s.prompts[prompt.Name] = prompt
}

type contextKey string

const ProfileIDKey contextKey = "profileID"
const OrgKey contextKey = "org"

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

		responseBytes := s.HandleMessage(context.Background(), line)
		if responseBytes != nil {
			os.Stdout.Write(responseBytes)
			os.Stdout.Write([]byte("\n"))
		}
	}
}

func (s *Server) HandleMessage(ctx context.Context, line []byte) []byte {
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
				"name":    "substrate-headless-mcp",
				"version": "0.1.0",
			},
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
				"prompts":   map[string]any{},
			},
		}
	case "tools/list":
		var allowed map[string]bool
		if profileID, ok := ctx.Value(ProfileIDKey).(int); ok && s.store != nil {
			profile, err := s.store.GetAgentProfile(context.Background(), profileID)
			if err == nil {
				allowed = make(map[string]bool)
				for _, t := range profile.AllowedTools {
					allowed[t] = true
				}
			}
		}

		toolsList := []map[string]any{}
		for _, t := range s.tools {
			if allowed != nil && !allowed[t.Name] {
				continue
			}
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

		if profileID, ok := ctx.Value(ProfileIDKey).(int); ok && s.store != nil && tool.IsDestructive {
			profile, err := s.store.GetAgentProfile(context.Background(), profileID)
			if err == nil && profile.HITLEnabled {
				org, _ := ctx.Value(OrgKey).(string)
				if org == "" {
					org = profile.Org
				}

				item, err := s.store.CreateHITLQueueItem(context.Background(), db.HITLQueueItem{
					Org:       org,
					ProfileID: profile.ID,
					ToolName:  params.Name,
					Arguments: params.Arguments,
					Status:    "pending",
				})

				if err != nil {
					return s.errorResponse(req.ID, -32603, "Internal error queuing HITL task")
				}

				pendingRes := fmt.Sprintf(`{"status":"pending_approval","queue_id":%d}`, item.ID)
				result = map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": pendingRes,
						},
					},
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
		}

		// Audit logging defer
		var reqPayload []byte
		if p, err := json.Marshal(params); err == nil {
			reqPayload = p
		}

		defer func() {
			// Use a non-blocking goroutine to avoid impacting MCP latency
			go func() {
				if s.store != nil {
					var aID *int
					if profileID, ok := ctx.Value(ProfileIDKey).(int); ok {
						aID = &profileID
					}

					var resPayload []byte
					if r, err := json.Marshal(result); err == nil {
						resPayload = r
					}

					bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					s.store.InsertMCPAuditLog(bgCtx, aID, params.Name, reqPayload, resPayload)
				}
			}()
		}()

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

	case "resources/list":
		resourcesList := []map[string]any{}
		for _, r := range s.resources {
			resourcesList = append(resourcesList, map[string]any{
				"uri":         r.URI,
				"name":        r.Name,
				"description": r.Description,
				"mimeType":    r.MimeType,
			})
		}
		result = map[string]any{
			"resources": resourcesList,
		}

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.errorResponse(req.ID, -32602, "Invalid params")
		}

		var handler func(uri string) (string, error)
		var mimeType string

		// Match exact or prefix template (simplistic routing)
		for uri, r := range s.resources {
			if params.URI == uri || isMatch(uri, params.URI) {
				handler = r.Handler
				mimeType = r.MimeType
				break
			}
		}

		if handler == nil {
			return s.errorResponse(req.ID, -32602, "Resource not found")
		}

		content, err := handler(params.URI)
		if err != nil {
			return s.errorResponse(req.ID, -32603, fmt.Sprintf("Resource read error: %v", err))
		}

		result = map[string]any{
			"contents": []map[string]any{
				{
					"uri":      params.URI,
					"mimeType": mimeType,
					"text":     content,
				},
			},
		}

	case "prompts/list":
		promptsList := []map[string]any{}
		for _, p := range s.prompts {
			promptsList = append(promptsList, map[string]any{
				"name":        p.Name,
				"description": p.Description,
			})
		}
		result = map[string]any{
			"prompts": promptsList,
		}

	case "prompts/get":
		var params struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.errorResponse(req.ID, -32602, "Invalid params")
		}

		prompt, exists := s.prompts[params.Name]
		if !exists {
			return s.errorResponse(req.ID, -32602, "Prompt not found")
		}

		content, err := prompt.Handler()
		if err != nil {
			return s.errorResponse(req.ID, -32603, fmt.Sprintf("Prompt error: %v", err))
		}

		result = map[string]any{
			"description": prompt.Description,
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": content,
					},
				},
			},
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

// Simple prefix match for templated URIs (e.g., substrate://schemas/{org}/{repo})
func isMatch(template, actual string) bool {
	// A more robust implementation would use regex, but for now we just check if actual starts with the prefix.
	// For "substrate://schemas/{org}/{repo}", we can check if actual starts with "substrate://schemas/"

	importStr := "substrate://"
	if len(template) > len(importStr) && len(actual) > len(importStr) {
		// Find first {
		braceIdx := -1
		for i, c := range template {
			if c == '{' {
				braceIdx = i
				break
			}
		}
		if braceIdx != -1 {
			prefix := template[:braceIdx]
			if len(actual) >= len(prefix) && actual[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}
