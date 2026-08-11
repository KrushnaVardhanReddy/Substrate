// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package consumers

import (
	"testing"
)

func TestParse_Valid(t *testing.T) {
	data := []byte(`
schema_version: "1.0"
provider: "github.com/myorg/backend-api"
consumes:
  - path: "GET /api/v1/users"
    fields:
      - "id"
      - "email"
`)

	m, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.SchemaVersion != "1.0" {
		t.Errorf("expected 1.0, got %s", m.SchemaVersion)
	}

	if m.Provider != "github.com/myorg/backend-api" {
		t.Errorf("expected github.com/myorg/backend-api, got %s", m.Provider)
	}

	if len(m.Consumes) != 1 {
		t.Fatalf("expected 1 consumer path, got %d", len(m.Consumes))
	}

	if m.Consumes[0].Path != "GET /api/v1/users" {
		t.Errorf("expected GET /api/v1/users, got %s", m.Consumes[0].Path)
	}

	if len(m.Consumes[0].Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(m.Consumes[0].Fields))
	}
}

func TestParse_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data string
		err  string
	}{
		{
			name: "missing schema_version",
			data: `
provider: "github.com/myorg/backend-api"
consumes:
  - path: "GET /api/v1/users"
    fields:
      - "id"
`,
			err: "missing schema_version",
		},
		{
			name: "missing provider",
			data: `
schema_version: "1.0"
consumes:
  - path: "GET /api/v1/users"
    fields:
      - "id"
`,
			err: "missing provider",
		},
		{
			name: "missing consumes",
			data: `
schema_version: "1.0"
provider: "github.com/myorg/backend-api"
`,
			err: "missing consumes",
		},
		{
			name: "missing path in consumes",
			data: `
schema_version: "1.0"
provider: "github.com/myorg/backend-api"
consumes:
  - fields:
      - "id"
`,
			err: "missing path in consumes",
		},
		{
			name: "missing fields in consumes",
			data: `
schema_version: "1.0"
provider: "github.com/myorg/backend-api"
consumes:
  - path: "GET /api/v1/users"
`,
			err: "missing fields in consumes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.data))
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tt.err {
				t.Errorf("expected %q, got %q", tt.err, err.Error())
			}
		})
	}
}
