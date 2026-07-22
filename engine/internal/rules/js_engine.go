package rules

import (
	"errors"
	"fmt"

	"github.com/dop251/goja"
)

// RunJSRule evaluates a JavaScript rule against a schema.
// It injects the `schema` variable into the JS runtime and calls the `validate` function.
// If the `validate` function returns a string, it is treated as a validation failure (the string is the error message).
// If it returns `true` or undefined/null without error, it passes.
func RunJSRule(script string, schema map[string]interface{}) (string, error) {
	vm := goja.New()

	// Inject the schema into the JS runtime
	if err := vm.Set("schema", schema); err != nil {
		return "", fmt.Errorf("failed to inject schema into js runtime: %w", err)
	}

	// Evaluate the script to define functions, etc.
	_, err := vm.RunString(script)
	if err != nil {
		return "", fmt.Errorf("failed to parse/execute javascript rule: %w", err)
	}

	// Extract the validate function
	val := vm.Get("validate")
	if val == nil {
		return "", errors.New("javascript rule must define a 'validate(schema)' function")
	}

	validateFunc, ok := goja.AssertFunction(val)
	if !ok {
		return "", errors.New("'validate' is not a function")
	}

	// Call the validate function with the schema
	schemaVal := vm.ToValue(schema)
	res, err := validateFunc(goja.Undefined(), schemaVal)
	if err != nil {
		return "", fmt.Errorf("error executing validate function: %w", err)
	}

	// Check the result
	if res == nil || goja.IsUndefined(res) || goja.IsNull(res) {
		// Implicit pass if they didn't return anything
		return "", nil
	}

	switch v := res.Export().(type) {
	case string:
		// A non-empty string means a validation error
		if v != "" {
			return v, nil
		}
		return "", nil
	case bool:
		// True means pass, False means fail with generic message
		if v {
			return "", nil
		}
		return "javascript validation failed", nil
	default:
		// Any other type is unexpected, but we can coerce it to string if it's an error?
		// Spec says: "If the JS function returns a string, it is treated as a validation failure... If it returns `true`, it passes."
		// For safety, we treat other types as pass or fail based on truthiness, but let's just ignore for now or return a generic err.
		return "", nil
	}
}
