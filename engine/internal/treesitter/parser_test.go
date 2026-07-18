package treesitter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFieldUsages(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "treesitter_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy JS file
	jsContent := `
const user = {
  user_id: 123,
  name: "Alice"
};

console.log(user.user_id);

const { user_id } = user;

function printUser({ user_id, name }) {
  console.log(user_id);
}
`
	jsFilePath := filepath.Join(tempDir, "test.js")
	if err := os.WriteFile(jsFilePath, []byte(jsContent), 0644); err != nil {
		t.Fatalf("failed to write JS file: %v", err)
	}

	// Create a dummy TS file
	tsContent := `
interface User {
  user_id: number;
  name: string;
}

const user: User = {
  user_id: 123,
  name: "Bob"
};

console.log(user.user_id);
`
	tsFilePath := filepath.Join(tempDir, "test.ts")
	if err := os.WriteFile(tsFilePath, []byte(tsContent), 0644); err != nil {
		t.Fatalf("failed to write TS file: %v", err)
	}

	// Create an unsupported file
	txtFilePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(txtFilePath, []byte("user_id\n"), 0644); err != nil {
		t.Fatalf("failed to write TXT file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		targetField string
		expected    []int
		expectError bool
	}{
		{
			name:        "JS - property access and destructuring",
			filePath:    jsFilePath,
			targetField: "user_id",
			// Lines:
			// 3: user_id: 123, (property_identifier)
			// 7: console.log(user.user_id); (property_identifier)
			// 9: const { user_id } = user; (shorthand_property_identifier_pattern)
			// 11: function printUser({ user_id, name }) { (shorthand_property_identifier_pattern)
			// 12: console.log(user_id); (identifier - NOT property_identifier, wait, let's see)
			expected: []int{3, 7, 9, 11},
		},
		{
			name:        "TS - interface, property access",
			filePath:    tsFilePath,
			targetField: "user_id",
			// Lines:
			// 3: user_id: number;
			// 8: user_id: 123,
			// 12: console.log(user.user_id);
			expected: []int{3, 8, 12},
		},
		{
			name:        "Unsupported file extension",
			filePath:    txtFilePath,
			targetField: "user_id",
			expected:    nil,
		},
		{
			name:        "Missing file",
			filePath:    filepath.Join(tempDir, "missing.js"),
			targetField: "user_id",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usages, err := FindFieldUsages(tt.filePath, tt.targetField)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, usages)
			}
		})
	}
}
