// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package graphql

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/stretchr/testify/assert"
)

func TestFederationKeyBroken(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "base.graphql")
	headPath := filepath.Join(tmpDir, "head.graphql")

	// Base schema with @key and @shareable
	baseSchema := `
type User @key(fields: "id storeId") {
	id: ID!
	storeId: ID!
	name: String! @shareable
	email: String!
}
`
	err := os.WriteFile(basePath, []byte(baseSchema), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		headSchema   string
		expectedID   string
		expectedRule string
	}{
		{
			name: "No changes",
			headSchema: `
type User @key(fields: "id storeId") {
	id: ID!
	storeId: ID!
	name: String! @shareable
	email: String!
}
`,
			expectedID:   "",
			expectedRule: "",
		},
		{
			name: "Remove non-key field",
			headSchema: `
type User @key(fields: "id storeId") {
	id: ID!
	storeId: ID!
	name: String! @shareable
}
`,
			expectedID:   "gql-field-removed-User-email",
			expectedRule: "GQL_FIELD_REMOVED",
		},
		{
			name: "Remove shareable non-key field",
			headSchema: `
type User @key(fields: "id storeId") {
	id: ID!
	storeId: ID!
	email: String!
}
`,
			expectedID:   "gql-field-removed-User-name",
			expectedRule: "GQL_FIELD_REMOVED",
		},
		{
			name: "Remove key field triggers FederationKeyBroken",
			headSchema: `
type User @key(fields: "id storeId") {
	id: ID!
	name: String! @shareable
	email: String!
}
`,
			expectedID:   "gql-federation-key-broken-User-storeId",
			expectedRule: "FederationKeyBroken",
		},
		{
			name: "Change type of key field triggers FederationKeyBroken",
			headSchema: `
type User @key(fields: "id storeId") {
	id: String!
	storeId: ID!
	name: String! @shareable
	email: String!
}
`,
			expectedID:   "gql-federation-key-broken-User-id",
			expectedRule: "FederationKeyBroken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.WriteFile(headPath, []byte(tt.headSchema), 0644)
			assert.NoError(t, err)

			rep, err := CompareGraphQL(basePath, headPath)
			assert.NoError(t, err)

			if tt.expectedRule == "" {
				assert.Empty(t, rep.BreakingChanges)
			} else {
				assert.NotEmpty(t, rep.BreakingChanges)
				found := false
				for _, ch := range rep.BreakingChanges {
					if ch.RuleID == tt.expectedRule && ch.ID == tt.expectedID {
						found = true
						assert.Equal(t, report.ChangeSeverityBreaking, ch.Severity)
						break
					}
				}
				assert.True(t, found, "Expected breaking change %s (%s) not found. Got: %v", tt.expectedID, tt.expectedRule, rep.BreakingChanges)
			}
		})
	}
}
