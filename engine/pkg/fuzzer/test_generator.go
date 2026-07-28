package fuzzer

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

const testTemplate = `package tests

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestGeneratedFuzzing(t *testing.T) {
	e := httpexpect.Default(t, "{{.BaseURL}}")
{{range .Tests}}
	t.Run("{{.Name}}", func(t *testing.T) {
		e.{{.Method}}("{{.Path}}").
			WithJSON(map[string]interface{}{
				{{range $k, $v := .Payload}}"{{$k}}": {{$v}},
				{{end}}
			}).
			Expect().
			Status(http.StatusUnprocessableEntity) // Or expected failure status
	})
{{end}}
}
`

type TestData struct {
	BaseURL string
	Tests   []TestCase
}

type TestCase struct {
	Name    string
	Method  string
	Path    string
	Payload map[string]interface{}
}

// GenerateTests reads the OpenAPI spec from specPath and generates test cases
// using httpexpect to fuzz constraints like maxLength, minimum, and required.
func GenerateTests(specPath string, baseURL string) (string, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return "", fmt.Errorf("failed to load spec: %w", err)
	}

	err = doc.Validate(context.Background())
	if err != nil {
		return "", fmt.Errorf("invalid spec: %w", err)
	}

	if baseURL == "" {
		baseURL = "http://localhost:8080"
		if len(doc.Servers) > 0 {
			baseURL = doc.Servers[0].URL
		}
	}

	var tests []TestCase

	for path, pathItem := range doc.Paths.Map() {
		operations := map[string]*openapi3.Operation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"DELETE": pathItem.Delete,
			"PATCH":  pathItem.Patch,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}

			if op.RequestBody != nil && op.RequestBody.Value != nil {
				content := op.RequestBody.Value.Content["application/json"]
				if content != nil && content.Schema != nil && content.Schema.Value != nil {
					schema := content.Schema.Value

					// Check 'required' constraints
					for _, req := range schema.Required {
						payload := generateBasePayload(schema)
						delete(payload, req)
						tests = append(tests, TestCase{
							Name:    fmt.Sprintf("%s %s missing required %s", method, path, req),
							Method:  method,
							Path:    path,
							Payload: payload,
						})
					}

					// Check properties constraints
					for propName, propRef := range schema.Properties {
						if propRef.Value == nil {
							continue
						}
						prop := propRef.Value

						if prop.MaxLength != nil {
							payload := generateBasePayload(schema)
							excessString := strings.Repeat("a", int(*prop.MaxLength)+1)
							payload[propName] = fmt.Sprintf("%q", excessString)
							tests = append(tests, TestCase{
								Name:    fmt.Sprintf("%s %s %s exceeds maxLength", method, path, propName),
								Method:  method,
								Path:    path,
								Payload: payload,
							})
						}

						if prop.Min != nil {
							payload := generateBasePayload(schema)
							payload[propName] = *prop.Min - 1
							tests = append(tests, TestCase{
								Name:    fmt.Sprintf("%s %s %s below minimum", method, path, propName),
								Method:  method,
								Path:    path,
								Payload: payload,
							})
						}
					}

					// Phase 10 - Task 01: Automated Security Fuzzing (OWASP)
					// Generate standard OWASP malicious payloads for robustness testing

					// SQLi payload
					sqliPayload := generateBasePayload(schema)
					for propName := range schema.Properties {
						sqliPayload[propName] = `"' OR 1=1 --"` // Quote it for the template
					}
					tests = append(tests, TestCase{
						Name:    fmt.Sprintf("%s %s SQL injection", method, path),
						Method:  method,
						Path:    path,
						Payload: sqliPayload,
					})

					// Path Traversal payload
					ptPayload := generateBasePayload(schema)
					for propName := range schema.Properties {
						ptPayload[propName] = `"../../../etc/passwd"`
					}
					tests = append(tests, TestCase{
						Name:    fmt.Sprintf("%s %s Path traversal", method, path),
						Method:  method,
						Path:    path,
						Payload: ptPayload,
					})

					// Null byte payload
					nullPayload := generateBasePayload(schema)
					for propName := range schema.Properties {
						nullPayload[propName] = `"test\x00"`
					}
					tests = append(tests, TestCase{
						Name:    fmt.Sprintf("%s %s Null byte injection", method, path),
						Method:  method,
						Path:    path,
						Payload: nullPayload,
					})

					// Extremely long string
					longStrPayload := generateBasePayload(schema)
					for propName, propRef := range schema.Properties {
						if propRef.Value != nil && propRef.Value.Type != nil && len(propRef.Value.Type.Slice()) > 0 && propRef.Value.Type.Slice()[0] == "string" {
							longStrPayload[propName] = fmt.Sprintf("%q", strings.Repeat("A", 10000))
						}
					}
					tests = append(tests, TestCase{
						Name:    fmt.Sprintf("%s %s Extremely long string", method, path),
						Method:  method,
						Path:    path,
						Payload: longStrPayload,
					})
				}
			}
		}
	}

	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	data := TestData{
		BaseURL: baseURL,
		Tests:   tests,
	}

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func generateBasePayload(schema *openapi3.Schema) map[string]interface{} {
	payload := make(map[string]interface{})
	for propName, propRef := range schema.Properties {
		if propRef.Value == nil {
			continue
		}
		prop := propRef.Value

		if prop.Type != nil && len(prop.Type.Slice()) > 0 {
			switch prop.Type.Slice()[0] { // openapi3 schema Type is a *Types slice in newer kin-openapi
			case "string":
				payload[propName] = `"test"`
			case "integer", "number":
				minVal := 0.0
				if prop.Min != nil {
					minVal = *prop.Min
				}
				payload[propName] = minVal
			case "boolean":
				payload[propName] = true
			}
		}
	}
	return payload
}
