package graphql_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/graphql"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareGraphQL(t *testing.T) {
	tests := []struct {
		name             string
		baseSchema       string
		headSchema       string
		expectErr        bool
		expectBreaking   int
		expectWarning    int
		expectSafe       int
		expectedRuleIDs  []string
		expectedWarnings []string
	}{
		{
			name:           "no changes",
			baseSchema:     `type User { id: ID! }`,
			headSchema:     `type User { id: ID! }`,
			expectErr:      false,
			expectBreaking: 0,
			expectWarning:  0,
		},
		{
			name:            "type removed",
			baseSchema:      `type User { id: ID! }`,
			headSchema:      ``,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_TYPE_REMOVED"},
		},
		{
			name:            "field removed",
			baseSchema:      `type User { id: ID!, name: String! }`,
			headSchema:      `type User { id: ID! }`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_FIELD_REMOVED"},
		},
		{
			name:            "field type changed",
			baseSchema:      `type User { age: Int }`,
			headSchema:      `type User { age: String }`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_FIELD_TYPE_CHANGED"},
		},
		{
			name:           "field type safe non-null change",
			baseSchema:     `type User { age: Int }`,
			headSchema:     `type User { age: Int! }`,
			expectErr:      false,
			expectBreaking: 0,
		},
		{
			name:            "required argument added",
			baseSchema:      `type Query { users: [User!]! } type User { id: ID! }`,
			headSchema:      `type Query { users(limit: Int!): [User!]! } type User { id: ID! }`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_REQUIRED_ARGUMENT_ADDED"},
		},
		{
			name:           "optional argument added",
			baseSchema:     `type Query { users: [User!]! } type User { id: ID! }`,
			headSchema:     `type Query { users(limit: Int): [User!]! } type User { id: ID! }`,
			expectErr:      false,
			expectBreaking: 0,
			expectSafe:     1,
		},
		{
			name:            "enum value removed",
			baseSchema:      `enum Role { ADMIN USER }`,
			headSchema:      `enum Role { ADMIN }`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_ENUM_VALUE_REMOVED"},
		},
		{
			name:             "field deprecated",
			baseSchema:       `type User { id: ID!, name: String }`,
			headSchema:       `type User { id: ID!, name: String @deprecated(reason: "old") }`,
			expectErr:        false,
			expectBreaking:   0,
			expectWarning:    1,
			expectedWarnings: []string{"GQL_FIELD_DEPRECATED"},
		},
		{
			name:            "union member removed",
			baseSchema:      `type User { id: ID! } type Org { id: ID! } union SearchResult = User | Org`,
			headSchema:      `type User { id: ID! } type Org { id: ID! } union SearchResult = User`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_UNION_MEMBER_REMOVED"},
		},
		{
			name:            "interface removed from object",
			baseSchema:      `interface Node { id: ID! } type User implements Node { id: ID! }`,
			headSchema:      `interface Node { id: ID! } type User { id: ID! }`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_INTERFACE_REMOVED"},
		},
		{
			name:            "directive removed",
			baseSchema:      `directive @mydir on FIELD_DEFINITION`,
			headSchema:      ``,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_DIRECTIVE_REMOVED"},
		},
		{
			name:            "directive location removed",
			baseSchema:      `directive @mydir on FIELD_DEFINITION | OBJECT`,
			headSchema:      `directive @mydir on FIELD_DEFINITION`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_DIRECTIVE_LOCATION_REMOVED"},
		},
		{
			name:            "input field required added",
			baseSchema:      `input UserInput { name: String }`,
			headSchema:      `input UserInput { name: String, age: Int! }`,
			expectErr:       false,
			expectBreaking:  1,
			expectedRuleIDs: []string{"GQL_INPUT_FIELD_ADDED_REQUIRED"},
		},
		{
			name:           "input field optional added",
			baseSchema:     `input UserInput { name: String }`,
			headSchema:     `input UserInput { name: String, age: Int }`,
			expectErr:      false,
			expectBreaking: 0,
			expectSafe:     1,
		},
		{
			name:       "invalid schema parsing",
			baseSchema: `type User { id: ID! `,
			headSchema: `type User { id: ID! }`,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			basePath := filepath.Join(dir, "base.graphql")
			headPath := filepath.Join(dir, "head.graphql")

			err := os.WriteFile(basePath, []byte(tt.baseSchema), 0644)
			require.NoError(t, err)

			err = os.WriteFile(headPath, []byte(tt.headSchema), 0644)
			require.NoError(t, err)

			rep, err := graphql.CompareGraphQL(basePath, headPath)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, rep)
			assert.Equal(t, report.SchemaTypeGraphQL, rep.SchemaType)
			assert.Equal(t, tt.expectBreaking, len(rep.BreakingChanges))
			assert.Equal(t, tt.expectWarning, len(rep.Warnings))
			assert.Equal(t, tt.expectSafe, len(rep.SafeChanges))

			if len(tt.expectedRuleIDs) > 0 {
				var foundIDs []string
				for _, bc := range rep.BreakingChanges {
					foundIDs = append(foundIDs, bc.RuleID)
				}
				assert.ElementsMatch(t, tt.expectedRuleIDs, foundIDs)
			}

			if len(tt.expectedWarnings) > 0 {
				var foundIDs []string
				for _, wc := range rep.Warnings {
					foundIDs = append(foundIDs, wc.RuleID)
				}
				assert.ElementsMatch(t, tt.expectedWarnings, foundIDs)
			}
		})
	}
}
