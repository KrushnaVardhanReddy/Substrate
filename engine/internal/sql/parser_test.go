package sql

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSchema(t *testing.T) {
	tempDir := t.TempDir()

	validDDL := `
		-- This is a comment
		SET statement_timeout = 0;

		CREATE TABLE users (
			id bigint PRIMARY KEY,
			email varchar(255) NOT NULL UNIQUE,
			age integer DEFAULT 18,
			status varchar(50) DEFAULT 'active',
			profile_id bigint
		);

		ALTER TABLE users ADD CONSTRAINT fk_profile FOREIGN KEY (profile_id) REFERENCES profiles(id);

		CREATE INDEX idx_users_email ON users(email);

		CREATE TYPE user_role AS ENUM ('admin', 'user', 'guest');

		CREATE VIEW active_users AS SELECT id, email FROM users WHERE status = 'active';

		CREATE SEQUENCE user_id_seq START 100 INCREMENT BY 1;
	`

	validPath := filepath.Join(tempDir, "valid.sql")
	err := os.WriteFile(validPath, []byte(validDDL), 0644)
	if err != nil {
		t.Fatalf("failed to write valid.sql: %v", err)
	}

	invalidDDL := `
		CREATE TABLE users (
			id bigint PRIMARY KEY
			email varchar(255) -- missing comma
		);
	`

	invalidPath := filepath.Join(tempDir, "invalid.sql")
	err = os.WriteFile(invalidPath, []byte(invalidDDL), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid.sql: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		expectErr bool
		validate  func(*testing.T, *SQLSchema)
	}{
		{
			name:      "valid schema",
			path:      validPath,
			expectErr: false,
			validate: func(t *testing.T, schema *SQLSchema) {
				if len(schema.Tables) != 1 {
					t.Errorf("expected 1 table, got %d", len(schema.Tables))
				}
				usersTable := schema.Tables["users"]
				if usersTable == nil {
					t.Fatalf("expected table 'users' to exist")
				}
				if len(usersTable.Columns) != 5 {
					t.Errorf("expected 5 columns, got %d", len(usersTable.Columns))
				}
				emailCol := usersTable.Columns["email"]
				if emailCol.DataType != "varchar" {
					t.Errorf("expected email DataType to be varchar, got %s", emailCol.DataType)
				}
				if emailCol.Length == nil || *emailCol.Length != 255 {
					t.Errorf("expected email Length to be 255, got %v", emailCol.Length)
				}
				if !emailCol.NotNull {
					t.Errorf("expected email NotNull to be true")
				}

				ageCol := usersTable.Columns["age"]
				if ageCol.Default == nil || *ageCol.Default != "18" {
					t.Errorf("expected age Default to be '18', got %v", ageCol.Default)
				}

				if len(usersTable.Constraints) != 3 {
					t.Errorf("expected 3 constraints (PK, UNIQUE, FK), got %d", len(usersTable.Constraints))
				}

				hasPK := false
				hasFK := false
				hasUnique := false
				for _, c := range usersTable.Constraints {
					if c.Type == "PRIMARY_KEY" {
						hasPK = true
					} else if c.Type == "FOREIGN_KEY" {
						hasFK = true
						if *c.RefTable != "profiles" {
							t.Errorf("expected FK ref table to be profiles, got %s", *c.RefTable)
						}
					} else if c.Type == "UNIQUE" {
						hasUnique = true
					}
				}

				if !hasPK {
					t.Errorf("expected primary key constraint")
				}
				if !hasFK {
					t.Errorf("expected foreign key constraint")
				}
				if !hasUnique {
					t.Errorf("expected unique constraint")
				}

				if len(usersTable.Indexes) != 1 {
					t.Errorf("expected 1 index, got %d", len(usersTable.Indexes))
				}
				if usersTable.Indexes["idx_users_email"] == nil {
					t.Errorf("expected index 'idx_users_email' to exist")
				}

				if len(schema.Views) != 1 {
					t.Errorf("expected 1 view, got %d", len(schema.Views))
				}
				if schema.Views["active_users"] == nil {
					t.Errorf("expected view 'active_users' to exist")
				}

				if len(schema.Enums) != 1 {
					t.Errorf("expected 1 enum, got %d", len(schema.Enums))
				}
				enum := schema.Enums["user_role"]
				if enum == nil {
					t.Fatalf("expected enum 'user_role' to exist")
				}
				if len(enum.Values) != 3 {
					t.Errorf("expected 3 enum values, got %d", len(enum.Values))
				}

				if len(schema.Sequences) != 1 {
					t.Errorf("expected 1 sequence, got %d", len(schema.Sequences))
				}
				seq := schema.Sequences["user_id_seq"]
				if seq == nil {
					t.Fatalf("expected sequence 'user_id_seq' to exist")
				}
				if seq.Start != 100 {
					t.Errorf("expected sequence Start to be 100, got %d", seq.Start)
				}
				if seq.Increment != 1 {
					t.Errorf("expected sequence Increment to be 1, got %d", seq.Increment)
				}
			},
		},
		{
			name:      "invalid schema",
			path:      invalidPath,
			expectErr: true,
			validate: func(t *testing.T, schema *SQLSchema) {
				if schema != nil {
					t.Errorf("expected schema to be nil on error")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := ParseSchema(tc.path)
			if (err != nil) != tc.expectErr {
				t.Fatalf("expected error: %v, got: %v", tc.expectErr, err)
			}
			tc.validate(t, schema)
		})
	}
}
