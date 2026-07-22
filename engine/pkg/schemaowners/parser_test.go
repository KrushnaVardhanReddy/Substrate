package schemaowners

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string][]string
	}{
		{
			name: "basic",
			content: `
# This is a comment
frontend/src/api/*.ts @frontend-lead
backend/src/api/*.go @backend-lead @alice
`,
			want: map[string][]string{
				"frontend/src/api/*.ts": {"@frontend-lead"},
				"backend/src/api/*.go":  {"@backend-lead", "@alice"},
			},
		},
		{
			name:    "empty",
			content: ``,
			want:    map[string][]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
