package init

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type InitOptions struct {
	Service    string
	Spec       string
	SchemaType string
	Branch     string
	Force      bool
	NoWorkflow bool
	NoConfig   bool
}

var ErrPermissionDenied = errors.New("write permission denied")
var ErrInternal = errors.New("internal error")

func Init(opts InitOptions) error {
	// Auto-detect spec if not provided
	specPath := opts.Spec
	if specPath == "" {
		specPath = autoDetectSpec()
	}

	missingSpec := false
	if specPath == "" {
		specPath = "openapi.yaml"
		missingSpec = true
	}

	// Service name
	serviceName := opts.Service
	if serviceName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("%w: failed to get working directory: %v", ErrInternal, err)
		}
		serviceName = filepath.Base(cwd)
	}

	configExists := fileExists("substrate.yaml")
	workflowExists := fileExists(".github/workflows/substrate.yml")

	if !opts.Force {
		shouldExit := false
		if !opts.NoConfig && configExists {
			fmt.Println("⚠️  substrate.yaml already exists. Use --force to overwrite.")
			shouldExit = true
		}
		if !opts.NoWorkflow && workflowExists {
			fmt.Println("⚠️  .github/workflows/substrate.yml already exists. Use --force to overwrite.")
			shouldExit = true
		}
		if shouldExit {
			return nil
		}
	}

	createdFiles := []string{}

	if !opts.NoConfig {
		configContent := strings.NewReplacer(
			"<service-name>", serviceName,
			"<schema-type>", opts.SchemaType,
			"<spec-path>", specPath,
		).Replace(SubstrateYAMLTemplate)

		err := os.WriteFile("substrate.yaml", []byte(configContent), 0644)
		if err != nil {
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return fmt.Errorf("%w: failed to write substrate.yaml: %v", ErrInternal, err)
		}
		createdFiles = append(createdFiles, "substrate.yaml")
	}

	if !opts.NoWorkflow {
		err := os.MkdirAll(".github/workflows", 0755)
		if err != nil {
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return fmt.Errorf("%w: failed to create .github/workflows directory: %v", ErrInternal, err)
		}

		workflowContent := strings.NewReplacer(
			"<branch>", opts.Branch,
			"<spec-path>", specPath,
		).Replace(WorkflowTemplate)

		err = os.WriteFile(".github/workflows/substrate.yml", []byte(workflowContent), 0644)
		if err != nil {
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return fmt.Errorf("%w: failed to write .github/workflows/substrate.yml: %v", ErrInternal, err)
		}
		createdFiles = append(createdFiles, ".github/workflows/substrate.yml")
	}

	// Output messages
	if missingSpec {
		fmt.Println("⚠️  No OpenAPI spec found. Set spec_path in substrate.yaml to the correct path.")
	}

	fmt.Println("\n✅ Substrate initialized!\n\nCreated:")
	for _, f := range createdFiles {
		fmt.Printf("  %s\n", f)
	}

	fmt.Println("\nNext steps:\n  1. Review substrate.yaml and update 'owners' with your team details.\n  2. Commit both files and open a PR to see Substrate in action.\n  3. Docs: https://github.com/KrushnaVardhanReddy/Substrate")

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func autoDetectSpec() string {
	candidates := []string{
		"openapi.yaml",
		"openapi.json",
		"api/openapi.yaml",
		"api/openapi.json",
		"docs/openapi.yaml",
		"docs/openapi.json",
		"swagger.yaml",
		"swagger.json",
	}

	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}

	return ""
}
