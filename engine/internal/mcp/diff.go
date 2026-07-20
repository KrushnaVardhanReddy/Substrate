package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

type Severity string

const (
	SeverityHigh Severity = "SEVERITY_HIGH"
)

type BreakingChange struct {
	Type        string
	Description string
	Severity    Severity
}

type MCPState struct {
	Tools     map[string]ToolDef
	Resources map[string]ResourceDef
	Prompts   map[string]PromptDef
}

type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ResourceDef struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType,omitempty"`
}

type PromptDef struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}


func DiffMCPState(base, current *MCPState) []BreakingChange {
	var changes []BreakingChange

	if base == nil || current == nil {
		return changes
	}

	// Diff Tools
	for name, baseTool := range base.Tools {
		currentTool, exists := current.Tools[name]
		if !exists {
			changes = append(changes, BreakingChange{
				Type:        "TOOL_REMOVED",
				Description: fmt.Sprintf("Tool %q was removed", name),
				Severity:    SeverityHigh,
			})
			continue
		}

		// Diff Parameters
		baseProps, _ := extractProperties(baseTool.InputSchema)
		currProps, _ := extractProperties(currentTool.InputSchema)
		baseReq := extractRequired(baseTool.InputSchema)
		currReq := extractRequired(currentTool.InputSchema)

		// Parameter removed or type changed
		for pName, bProp := range baseProps {
			cProp, exists := currProps[pName]
			if !exists {
				changes = append(changes, BreakingChange{
					Type:        "PARAMETER_REMOVED",
					Description: fmt.Sprintf("Parameter %q was removed from tool %q", pName, name),
					Severity:    SeverityHigh,
				})
			} else {
				// check type
				bType, _ := bProp["type"].(string)
				cType, _ := cProp["type"].(string)
				if bType != "" && cType != "" && bType != cType {
					changes = append(changes, BreakingChange{
						Type:        "PARAMETER_TYPE_CHANGED",
						Description: fmt.Sprintf("Parameter %q in tool %q changed type from %q to %q", pName, name, bType, cType),
						Severity:    SeverityHigh,
					})
				}
			}
		}

		// New required parameter
		for pName := range currReq {
			if !baseReq[pName] {
				// new required parameter
				changes = append(changes, BreakingChange{
					Type:        "NEW_REQUIRED_PARAMETER",
					Description: fmt.Sprintf("New required parameter %q added to tool %q", pName, name),
					Severity:    SeverityHigh,
				})
			}
		}
	}

	// Diff Resources
	for uri, baseRes := range base.Resources {
		currentRes, exists := current.Resources[uri]
		if !exists {
			changes = append(changes, BreakingChange{
				Type:        "RESOURCE_URI_REMOVED",
				Description: fmt.Sprintf("Resource URI %q was removed", uri),
				Severity:    SeverityHigh,
			})
			continue
		}

		if baseRes.MimeType != currentRes.MimeType && baseRes.MimeType != "" {
			changes = append(changes, BreakingChange{
				Type:        "RESOURCE_MIME_TYPE_CHANGED",
				Description: fmt.Sprintf("Resource %q changed MIME type from %q to %q", uri, baseRes.MimeType, currentRes.MimeType),
				Severity:    SeverityHigh,
			})
		}
	}

	// Diff Prompts
	for name, basePrompt := range base.Prompts {
		currentPrompt, exists := current.Prompts[name]
		if !exists {
			changes = append(changes, BreakingChange{
				Type:        "PROMPT_REMOVED",
				Description: fmt.Sprintf("Prompt %q was removed", name),
				Severity:    SeverityHigh,
			})
			continue
		}

		baseArgs := make(map[string]PromptArgument)
		for _, arg := range basePrompt.Arguments {
			baseArgs[arg.Name] = arg
		}

		for _, currArg := range currentPrompt.Arguments {
			if currArg.Required {
				bArg, exists := baseArgs[currArg.Name]
				if !exists || !bArg.Required {
					changes = append(changes, BreakingChange{
						Type:        "PROMPT_REQUIRED_ARGUMENT_ADDED",
						Description: fmt.Sprintf("New required argument %q added to prompt %q", currArg.Name, name),
						Severity:    SeverityHigh,
					})
				}
			}
		}
	}

	return changes
}

func extractProperties(schema map[string]any) (map[string]map[string]any, bool) {
	if schema == nil {
		return nil, false
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil, false
	}
	res := make(map[string]map[string]any)
	for k, v := range props {
		if vm, ok := v.(map[string]any); ok {
			res[k] = vm
		}
	}
	return res, true
}

func extractRequired(schema map[string]any) map[string]bool {
	res := make(map[string]bool)
	if schema == nil {
		return res
	}
	reqs, ok := schema["required"].([]any)
	if !ok {
		return res
	}
	for _, r := range reqs {
		if rs, ok := r.(string); ok {
			res[rs] = true
		}
	}
	return res
}

func sendRequest(ctx context.Context, stdin io.Writer, scanner *bufio.Scanner, method string, id int, params any) ([]byte, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
	}
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		req.Params = b
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqBytes = append(reqBytes, '\n')
	_, err = stdin.Write(reqBytes)
	if err != nil {
		return nil, err
	}

	type scanResult struct {
		line []byte
		err  error
	}
	resChan := make(chan scanResult, 1)

	go func() {
		if scanner.Scan() {
			// make a copy because scanner.Bytes() is overwritten by next Scan()
			line := make([]byte, len(scanner.Bytes()))
			copy(line, scanner.Bytes())
			resChan <- scanResult{line: line, err: nil}
		} else {
			resChan <- scanResult{err: scanner.Err()}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resChan:
		if res.err != nil {
			return nil, res.err
		}
		if len(res.line) == 0 {
			return nil, fmt.Errorf("empty response")
		}
		return res.line, nil
	}
}

func CaptureMCPState(ctx context.Context, command string, args []string) (*MCPState, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Capture stderr for debugging
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	scanner := bufio.NewScanner(stdout)
	state := &MCPState{
		Tools:     make(map[string]ToolDef),
		Resources: make(map[string]ResourceDef),
		Prompts:   make(map[string]PromptDef),
	}

	// 1. initialize
	initParams := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]string{
			"name":    "substrate-diff-engine",
			"version": "1.0.0",
		},
	}
	initLine, err := sendRequest(ctx, stdin, scanner, "initialize", 1, initParams)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize: %w. stderr: %s", err, stderrBuf.String())
	}
	var initResp JSONRPCResponse
	if err := json.Unmarshal(initLine, &initResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initialize response: %w", err)
	}
	if initResp.Error != nil {
		return nil, fmt.Errorf("initialize returned error: %v", initResp.Error.Message)
	}

	// Send initialized notification
	initializedReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	initializedReqBytes, _ := json.Marshal(initializedReq)
	initializedReqBytes = append(initializedReqBytes, '\n')
	stdin.Write(initializedReqBytes)

	// 2. tools/list
	toolsLine, err := sendRequest(ctx, stdin, scanner, "tools/list", 2, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	var toolsResp JSONRPCResponse
	if err := json.Unmarshal(toolsLine, &toolsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools/list response: %w", err)
	}
	if toolsResp.Error == nil && len(toolsResp.Result) > 0 {
		var toolsResult struct {
			Tools []ToolDef `json:"tools"`
		}
		if err := json.Unmarshal(toolsResp.Result, &toolsResult); err == nil {
			for _, t := range toolsResult.Tools {
				state.Tools[t.Name] = t
			}
		}
	}

	// 3. resources/list
	resourcesLine, err := sendRequest(ctx, stdin, scanner, "resources/list", 3, nil)
	if err == nil {
		var resResp JSONRPCResponse
		if err := json.Unmarshal(resourcesLine, &resResp); err == nil {
			if resResp.Error == nil && len(resResp.Result) > 0 {
				var resResult struct {
					Resources []ResourceDef `json:"resources"`
				}
				if err := json.Unmarshal(resResp.Result, &resResult); err == nil {
					for _, r := range resResult.Resources {
						state.Resources[r.URI] = r
					}
				}
			}
		}
	}

	// 4. prompts/list
	promptsLine, err := sendRequest(ctx, stdin, scanner, "prompts/list", 4, nil)
	if err == nil {
		var promptResp JSONRPCResponse
		if err := json.Unmarshal(promptsLine, &promptResp); err == nil {
			if promptResp.Error == nil && len(promptResp.Result) > 0 {
				var promptResult struct {
					Prompts []PromptDef `json:"prompts"`
				}
				if err := json.Unmarshal(promptResp.Result, &promptResult); err == nil {
					for _, p := range promptResult.Prompts {
						state.Prompts[p.Name] = p
					}
				}
			}
		}
	}

	return state, nil
}
