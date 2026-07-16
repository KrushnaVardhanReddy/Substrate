package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func safeDiff(t *testing.T, a, b string) (result string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DiffSchemas panicked with input a=%q b=%q: %v", a[:min(len(a), 50)], b[:min(len(b), 50)], r)
		}
	}()

	timer := time.AfterFunc(10*time.Second, func() {
		os.Exit(1)
	})
	defer timer.Stop()

	return engine.DiffSchemas(a, b)
}

func assertValidJSON(t *testing.T, result string) {
	t.Helper()
	if !json.Valid([]byte(result)) {
		t.Errorf("result is not valid JSON: %s", result[:min(len(result), 200)])
	}
}

const validOpenAPIv3 = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /ping:
    get:
      responses:
        "200":
          description: OK`

const validOpenAPIv3Modified = `openapi: 3.0.0
info:
  title: Test API Modified
  version: 1.0.0
paths:
  /ping:
    get:
      responses:
        "200":
          description: OK
  /pong:
    get:
      responses:
        "200":
          description: OK`

func TestDiffSchemas_BoundaryInputs(t *testing.T) {
	cases := []struct {
		name string
		a, b string
	}{
		{"EmptyStrings", "", ""},
		{"ValidOpenAPI", validOpenAPIv3, validOpenAPIv3Modified},
		{"NullBytes", "\x00\x00\x00", "\x00"},
		{"UnicodeGarbage", "日本語テスト\u0000☃️", "test"},
		{"LargePayload", strings.Repeat("x", 10_000_000), strings.Repeat("y", 10_000_000)},
		{"TruncatedYAML", "openapi: 3.0.0\ninfo:\n  title:", ""},
		{"JSONvYAML", `{"openapi":"3.0.0"}`, "openapi: 3.0.0"},
		{"OneSidedEmpty", validOpenAPIv3, ""},
		{"OneSidedNil", "", validOpenAPIv3},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := safeDiff(t, tc.a, tc.b)
			assertValidJSON(t, result)
		})
	}
}

func FuzzDiffSchemas(f *testing.F) {
	f.Add("", "")
	f.Add("openapi: 3.0.0", "openapi: 3.1.0")
	f.Add(validOpenAPIv3, validOpenAPIv3Modified)

	f.Fuzz(func(t *testing.T, a, b string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on inputs a=%q b=%q: %v", a[:min(len(a), 100)], b[:min(len(b), 100)], r)
			}
		}()
		result := engine.DiffSchemas(a, b)
		if !json.Valid([]byte(result)) {
			t.Errorf("non-JSON result for a=%q b=%q: %s", a[:min(len(a), 50)], b[:min(len(b), 50)], result[:min(len(result), 100)])
		}
	})
}
