package rules

import (
	"testing"
)

func TestRunJSRule(t *testing.T) {
	script := `
		function validate(schema) {
			for (const path in schema.paths) {
				for (const method in schema.paths[path]) {
					const params = schema.paths[path][method].parameters || [];
					const hasHeader = params.some(p => p.in === 'header' && p.name === 'X-Correlation-ID');
					if (!hasHeader) {
						return "Missing X-Correlation-ID header on " + method.toUpperCase() + " " + path;
					}
				}
			}
			return true;
		}
	`

	tests := []struct {
		name       string
		script     string
		schema     map[string]interface{}
		wantErrMsg string
		wantErr    bool
	}{
		{
			name:   "Passes when header is present",
			script: script,
			schema: map[string]interface{}{
				"paths": map[string]interface{}{
					"/users": map[string]interface{}{
						"get": map[string]interface{}{
							"parameters": []interface{}{
								map[string]interface{}{
									"in":   "header",
									"name": "X-Correlation-ID",
								},
							},
						},
					},
				},
			},
			wantErrMsg: "",
			wantErr:    false,
		},
		{
			name:   "Fails when header is missing",
			script: script,
			schema: map[string]interface{}{
				"paths": map[string]interface{}{
					"/users": map[string]interface{}{
						"post": map[string]interface{}{
							"parameters": []interface{}{
								map[string]interface{}{
									"in":   "header",
									"name": "Content-Type",
								},
							},
						},
					},
				},
			},
			wantErrMsg: "Missing X-Correlation-ID header on POST /users",
			wantErr:    false,
		},
		{
			name:   "Fails when parameters are completely missing",
			script: script,
			schema: map[string]interface{}{
				"paths": map[string]interface{}{
					"/status": map[string]interface{}{
						"get": map[string]interface{}{},
					},
				},
			},
			wantErrMsg: "Missing X-Correlation-ID header on GET /status",
			wantErr:    false,
		},
		{
			name:       "Invalid JS syntax",
			script:     "function validate(schema) { return ",
			schema:     map[string]interface{}{},
			wantErrMsg: "",
			wantErr:    true,
		},
		{
			name:       "Missing validate function",
			script:     "function somethingElse() {}",
			schema:     map[string]interface{}{},
			wantErrMsg: "",
			wantErr:    true,
		},
		{
			name:   "Returns undefined implicit pass",
			script: "function validate(schema) { }",
			schema: map[string]interface{}{},
			wantErrMsg: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := RunJSRule(tt.script, tt.schema)

			if (err != nil) != tt.wantErr {
				t.Fatalf("RunJSRule() error = %v, wantErr %v", err, tt.wantErr)
			}

			if msg != tt.wantErrMsg {
				t.Errorf("RunJSRule() msg = %v, want %v", msg, tt.wantErrMsg)
			}
		})
	}
}
