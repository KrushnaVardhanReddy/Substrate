package gateway

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestAWSGenerator(t *testing.T) {
	g := &AWSGenerator{}

	spec := &openapi3.T{
		Paths: openapi3.NewPaths(),
	}
	spec.Paths.Set("/users", &openapi3.PathItem{})

	crd, err := g.GenerateCRD(spec, "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(crd) == 0 {
		t.Fatal("expected crd to not be empty")
	}

	if !contains(crd, "/users") {
		t.Errorf("expected crd to contain %q", "/users")
	}
}
