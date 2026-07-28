package gateway

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestKongGenerator(t *testing.T) {
	g := &KongGenerator{}

	spec := &openapi3.T{
		Paths: openapi3.NewPaths(),
	}
	spec.Paths.Set("/users", &openapi3.PathItem{})
	spec.Paths.Set("/posts", &openapi3.PathItem{})

	crd, err := g.GenerateCRD(spec, "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(crd) == 0 {
		t.Fatal("expected crd to not be empty")
	}

	expectedContains := []string{
		"kind: KongIngress",
		"kind: Ingress",
		"name: my-repo",
		"/users",
		"/posts",
	}

	for _, s := range expectedContains {
		if !contains(crd, s) {
			t.Errorf("expected crd to contain %q", s)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
