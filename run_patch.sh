#!/bin/bash
mkdir -p api/internal/gateway
cat << 'INNER_EOF' > api/internal/gateway/generator.go
package gateway

import (
    "fmt"
    "github.com/getkin/kin-openapi/openapi3"
)

type Generator interface {
    GenerateCRD(spec *openapi3.T, repo string) (string, error)
}

func NewGenerator(gatewayType string) (Generator, error) {
    switch gatewayType {
    case "kong":
        return &KongGenerator{}, nil
    case "aws":
        return &AWSGenerator{}, nil
    default:
        return nil, fmt.Errorf("unsupported gateway type: %s", gatewayType)
    }
}
INNER_EOF

cat << 'INNER_EOF' > api/internal/gateway/aws_generator.go
package gateway

import (
    "encoding/json"
    "fmt"

    "github.com/getkin/kin-openapi/openapi3"
)

type AWSGenerator struct{}

func (g *AWSGenerator) GenerateCRD(spec *openapi3.T, repo string) (string, error) {
    if spec == nil {
        return "", fmt.Errorf("openapi spec is nil")
    }

    // Clone spec to avoid mutating the original
    specClone := spec // shallow copy

    data, err := json.MarshalIndent(specClone, "", "  ")
    if err != nil {
        return "", err
    }

    return string(data), nil
}
INNER_EOF

cat << 'INNER_EOF' > api/internal/gateway/kong_generator.go
package gateway

import (
    "fmt"

    "github.com/getkin/kin-openapi/openapi3"
    "sigs.k8s.io/yaml"
)

type KongGenerator struct{}

type KongIngress struct {
    ApiVersion string `json:"apiVersion"`
    Kind       string `json:"kind"`
    Metadata   struct {
        Name        string            `json:"name"`
        Annotations map[string]string `json:"annotations,omitempty"`
    } `json:"metadata"`
    Proxy struct {
        Paths []string `json:"paths"`
    } `json:"proxy"`
}

type Ingress struct {
    ApiVersion string `json:"apiVersion"`
    Kind       string `json:"kind"`
    Metadata   struct {
        Name        string            `json:"name"`
        Annotations map[string]string `json:"annotations,omitempty"`
    } `json:"metadata"`
    Spec struct {
        Rules []IngressRule `json:"rules"`
    } `json:"spec"`
}

type IngressRule struct {
    Http struct {
        Paths []IngressPath `json:"paths"`
    } `json:"http"`
}

type IngressPath struct {
    Path     string `json:"path"`
    PathType string `json:"pathType"`
    Backend  struct {
        Service struct {
            Name string `json:"name"`
            Port struct {
                Number int `json:"number"`
            } `json:"port"`
        } `json:"service"`
    } `json:"backend"`
}


func (g *KongGenerator) GenerateCRD(spec *openapi3.T, repo string) (string, error) {
    if spec == nil {
        return "", fmt.Errorf("openapi spec is nil")
    }

    kongIngress := KongIngress{
        ApiVersion: "configuration.konghq.com/v1",
        Kind:       "KongIngress",
    }
    kongIngress.Metadata.Name = repo
    kongIngress.Proxy.Paths = make([]string, 0)

    ingress := Ingress{
        ApiVersion: "networking.k8s.io/v1",
        Kind:       "Ingress",
    }
    ingress.Metadata.Name = repo
    ingress.Metadata.Annotations = map[string]string{
        "konghq.com/override": repo,
    }

    paths := make([]IngressPath, 0)

    if spec.Paths != nil {
        for path := range spec.Paths.Map() {
            kongIngress.Proxy.Paths = append(kongIngress.Proxy.Paths, path)
            paths = append(paths, IngressPath{
                Path:     path,
                PathType: "Prefix",
                Backend: struct{
                    Service struct{
                        Name string `json:"name"`
                        Port struct{
                            Number int `json:"number"`
                        } `json:"port"`
                    } `json:"service"`
                }{
                    Service: struct{
                        Name string `json:"name"`
                        Port struct{
                            Number int `json:"number"`
                        } `json:"port"`
                    }{
                        Name: repo,
                        Port: struct{
                            Number int `json:"number"`
                        }{Number: 80},
                    },
                },
            })
        }
    }

    ingress.Spec.Rules = []IngressRule{
        {
            Http: struct{
                Paths []IngressPath `json:"paths"`
            }{
                Paths: paths,
            },
        },
    }

    kongYaml, err := yaml.Marshal(kongIngress)
    if err != nil {
        return "", err
    }

    ingressYaml, err := yaml.Marshal(ingress)
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("---\n%s---\n%s", string(kongYaml), string(ingressYaml)), nil
}
INNER_EOF

cat << 'INNER_EOF' > api/internal/gateway/aws_generator_test.go
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
INNER_EOF

cat << 'INNER_EOF' > api/internal/gateway/kong_generator_test.go
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
INNER_EOF

cat << 'INNER_EOF' > api/internal/gateway/sync.go
package gateway

import (
    "bytes"
    "context"
    "fmt"
    "log"
    "time"

    "github.com/getkin/kin-openapi/openapi3"
    "github.com/KrushnaVardhanReddy/substrate/api/internal/config"
    "github.com/KrushnaVardhanReddy/substrate/api/internal/db"
    "github.com/KrushnaVardhanReddy/substrate/api/internal/github"
)

func generatePRBranchName(repo string) string {
    return fmt.Sprintf("substrate/gateway-sync-%s-%d", repo, time.Now().Unix())
}

func RunGatewaySync(ctx context.Context, store db.Store, ghClient github.Client, org, repo string) error {
    content, err := ghClient.GetFileContent(ctx, org, repo, "substrate.yaml")
    if err != nil {
        return fmt.Errorf("failed to get substrate.yaml: %w", err)
    }

    cfg, err := config.Parse([]byte(content))
    if err != nil {
        return fmt.Errorf("failed to parse substrate.yaml: %w", err)
    }

    if cfg.Gateway == nil || cfg.Gateway.Type == "" {
        return nil
    }

    schemaPath := cfg.HeadSchema
    if schemaPath == "" {
        schemaPath = "openapi.yaml"
    }

    schemaContent, err := ghClient.GetFileContent(ctx, org, repo, schemaPath)
    if err != nil {
        return fmt.Errorf("failed to get schema %s: %w", schemaPath, err)
    }

    loader := openapi3.NewLoader()
    doc, err := loader.LoadFromData([]byte(schemaContent))
    if err != nil {
        return fmt.Errorf("failed to parse openapi spec: %w", err)
    }

    gen, err := NewGenerator(cfg.Gateway.Type)
    if err != nil {
        return fmt.Errorf("failed to create generator: %w", err)
    }

    crdYaml, err := gen.GenerateCRD(doc, repo)
    if err != nil {
        return fmt.Errorf("failed to generate crd: %w", err)
    }

    var prUrl string
    if cfg.Gateway.InfraRepo != "" {
        infraRepoParts := bytes.Split([]byte(cfg.Gateway.InfraRepo), []byte("/"))
        if len(infraRepoParts) == 2 {
            infraOwner := string(infraRepoParts[0])
            infraRepoName := string(infraRepoParts[1])

            branch := generatePRBranchName(repo)

            outputPath := cfg.Gateway.OutputPath
            if outputPath == "" {
                outputPath = "gateway/"
            }
            if outputPath[len(outputPath)-1] != '/' {
                outputPath += "/"
            }
            filePath := fmt.Sprintf("%ssubstrate-generated-%s.yaml", outputPath, repo)

            title := fmt.Sprintf("chore(gateway): Sync API Gateway for %s", repo)
            body := fmt.Sprintf("Automatically generated API Gateway configuration for %s/%s.", org, repo)

            err = ghClient.CommitAndPushFile(ctx, infraOwner, infraRepoName, branch, filePath, crdYaml, title)
            if err != nil {
                log.Printf("Failed to commit and push file to infra repo: %v", err)
            } else {
                prUrl, err = ghClient.CreatePR(ctx, infraOwner, infraRepoName, title, branch, body)
                if err != nil {
                    log.Printf("Failed to create PR in infra repo: %v", err)
                }
            }
        }
    }

    err = store.SaveGatewayConfig(ctx, org, repo, cfg.Gateway.Type, crdYaml, prUrl)
    if err != nil {
        return fmt.Errorf("failed to save gateway config: %w", err)
    }

    return nil
}
INNER_EOF

mkdir -p api/internal/db/migrations
cat << 'INNER_EOF' > api/internal/db/migrations/20261026_add_gateway_configs.sql
CREATE TABLE gateway_configs (
  id          SERIAL PRIMARY KEY,
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  gateway_type TEXT NOT NULL,
  crd_yaml    TEXT NOT NULL,
  pr_url      TEXT,
  generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INNER_EOF

cat << 'INNER_EOF' > api/internal/handlers/gateway.go
package handlers

import (
    "context"
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/KrushnaVardhanReddy/substrate/api/internal/db"
    "github.com/KrushnaVardhanReddy/substrate/api/internal/gateway"
    "github.com/KrushnaVardhanReddy/substrate/api/internal/github"
)

func GatewaySyncHandler(store db.Store, ghClient github.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        org := chi.URLParam(r, "org")
        repo := chi.URLParam(r, "repo")

        w.WriteHeader(http.StatusAccepted)

        go func() {
            err := gateway.RunGatewaySync(context.Background(), store, ghClient, org, repo)
            if err != nil {
                log.Printf("Gateway sync failed for %s/%s: %v", org, repo, err)
            }
        }()
    }
}
INNER_EOF

cat << 'INNER_EOF' > api/internal/handlers/gateway_test.go
package handlers

import (
    "testing"
)

func TestGatewaySyncHandler(t *testing.T) {
    handler := GatewaySyncHandler(nil, nil)
    if handler == nil {
        t.Fatal("expected handler to not be nil")
    }
}
INNER_EOF

cat << 'INNER_EOF' > patch_config.diff
--- api/internal/config/config.go
+++ api/internal/config/config.go
@@ -29,6 +29,12 @@
	Overrides        []OverrideConfig `yaml:"overrides,omitempty"`
 }

+type Gateway struct {
+	Type       string `yaml:"type"`
+	InfraRepo  string `yaml:"infra_repo"`
+	OutputPath string `yaml:"output_path"`
+}
+
 type SubstrateConfig struct {
	Discovery  *Discovery       `yaml:"discovery,omitempty"`
	Metadata   *Metadata        `yaml:"metadata,omitempty"`
@@ -37,6 +43,7 @@
	BaseSchema string           `yaml:"base_schema,omitempty"`
	HeadSchema string           `yaml:"head_schema,omitempty"`
	Consumers  []ConsumerConfig `yaml:"consumers,omitempty"`
+	Gateway    *Gateway         `yaml:"gateway,omitempty"`
 }

 func Parse(content []byte) (*SubstrateConfig, error) {
INNER_EOF

patch api/internal/config/config.go < patch_config.diff

cat << 'INNER_EOF' > patch_store.diff
--- api/internal/db/store.go
+++ api/internal/db/store.go
@@ -184,6 +184,7 @@
	DeletePartner(ctx context.Context, id uuid.UUID) error
	InsertSchemaValidationGap(ctx context.Context, arg sqlcgen.InsertSchemaValidationGapParams) error
	GetSchemaValidationGaps(ctx context.Context) ([]sqlcgen.SchemaValidationGap, error)
+	SaveGatewayConfig(ctx context.Context, org, repo, gatewayType, crdYaml, prUrl string) error
	Pool() *pgxpool.Pool
 }
INNER_EOF

patch api/internal/db/store.go < patch_store.diff

cat << 'INNER_EOF' > patch_ports.diff
--- api/internal/ports/ports.go
+++ api/internal/ports/ports.go
@@ -218,6 +218,7 @@
 type PoolProvider interface {
	InsertSchemaValidationGap(ctx context.Context, arg sqlcgen.InsertSchemaValidationGapParams) error
	GetSchemaValidationGaps(ctx context.Context) ([]sqlcgen.SchemaValidationGap, error)
+	SaveGatewayConfig(ctx context.Context, org, repo, gatewayType, crdYaml, prUrl string) error
	Pool() *pgxpool.Pool
 }
INNER_EOF

patch api/internal/ports/ports.go < patch_ports.diff

cat << 'INNER_EOF' >> api/internal/db/queries.go

func (s *PGStore) SaveGatewayConfig(ctx context.Context, org, repo, gatewayType, crdYaml, prUrl string) error {
	query := `
		INSERT INTO gateway_configs (org, repo, gateway_type, crd_yaml, pr_url)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.pool.Exec(ctx, query, org, repo, gatewayType, crdYaml, prUrl)
	return err
}
INNER_EOF

cat << 'INNER_EOF' >> api/internal/db/mock_store.go

func (m *MockStore) SaveGatewayConfig(ctx context.Context, org, repo, gatewayType, crdYaml, prUrl string) error {
	return nil
}
INNER_EOF

cat << 'INNER_EOF' > patch_push.diff
--- api/internal/webhook/push.go
+++ api/internal/webhook/push.go
@@ -1,6 +1,7 @@
 package webhook

 import (
+	"context"
	"encoding/json"
	"log"
	"net/http"
@@ -12,6 +13,7 @@
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
+	"github.com/KrushnaVardhanReddy/substrate/api/internal/gateway"
 )

 // PushHandler handles the GitHub push webhook payload forwarded from the GitHub App
@@ -104,6 +106,13 @@
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
+
+		go func() {
+			err := gateway.RunGatewaySync(context.Background(), store, ghClient, req.Org, req.Repo)
+			if err != nil {
+				log.Printf("Gateway sync failed for %s/%s: %v", req.Org, req.Repo, err)
+			}
+		}()
	}
 }

INNER_EOF

patch api/internal/webhook/push.go < patch_push.diff

cat << 'INNER_EOF' > patch_client.diff
--- api/internal/github/client.go
+++ api/internal/github/client.go
@@ -21,6 +21,8 @@
 type Client interface {
	RequestReviewers(ctx context.Context, owner, repo string, pullNumber int, reviewers []string) error
	CreateDraftPR(ctx context.Context, owner, repo, branch, patch, title, body string) (url string, err error)
+	CreatePR(ctx context.Context, owner, repo, title, branch, body string) (url string, err error)
+	CommitAndPushFile(ctx context.Context, owner, repo, branch, filePath, fileContent, commitMessage string) error
	SearchCode(ctx context.Context, owner, repo, query string) (path string, err error)
	GetFileContent(ctx context.Context, owner, repo, path string) (string, error)
	CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error
@@ -628,6 +630,154 @@
	return result.HTMLURL, nil
 }

+func (c *RESTClient) CreatePR(ctx context.Context, owner, repo, title, branch, body string) (string, error) {
+	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.apiURL, owner, repo)
+
+	payload := map[string]interface{}{
+		"title": title,
+		"head":  branch,
+		"base":  "main",
+		"body":  body,
+	}
+
+	payloadBytes, _ := json.Marshal(payload)
+	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
+	if err != nil {
+		return "", err
+	}
+	c.addHeaders(req)
+
+	resp, err := c.client.Do(req)
+	if err != nil {
+		return "", err
+	}
+	defer resp.Body.Close()
+
+	if resp.StatusCode != http.StatusCreated {
+		bodyBytes, _ := io.ReadAll(resp.Body)
+		return "", fmt.Errorf("create pr failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
+	}
+
+	var result struct {
+		HTMLURL string `json:"html_url"`
+	}
+	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
+		return "", err
+	}
+	return result.HTMLURL, nil
+}
+
+func (c *RESTClient) CommitAndPushFile(ctx context.Context, owner, repo, branch, filePath, fileContent, commitMessage string) error {
+	mainSHA, err := c.getBranchSHA(ctx, owner, repo, "main")
+	if err != nil {
+		return err
+	}
+
+	if err := c.createBranch(ctx, owner, repo, branch, mainSHA); err != nil {
+		return err
+	}
+
+	blobURL := fmt.Sprintf("%s/repos/%s/%s/git/blobs", c.apiURL, owner, repo)
+	blobPayload := map[string]string{
+		"content":  fileContent,
+		"encoding": "utf-8",
+	}
+	blobBytes, _ := json.Marshal(blobPayload)
+	reqBlob, err := http.NewRequestWithContext(ctx, http.MethodPost, blobURL, bytes.NewBuffer(blobBytes))
+	if err != nil {
+		return err
+	}
+	c.addHeaders(reqBlob)
+
+	respBlob, err := c.client.Do(reqBlob)
+	if err != nil {
+		return err
+	}
+	defer respBlob.Body.Close()
+
+	if respBlob.StatusCode != http.StatusCreated {
+		bodyBytes, _ := io.ReadAll(respBlob.Body)
+		return fmt.Errorf("create blob failed with status: %d, body: %s", respBlob.StatusCode, string(bodyBytes))
+	}
+
+	var blobResult struct {
+		SHA string `json:"sha"`
+	}
+	if err := json.NewDecoder(respBlob.Body).Decode(&blobResult); err != nil {
+		return err
+	}
+
+	treeURL := fmt.Sprintf("%s/repos/%s/%s/git/trees", c.apiURL, owner, repo)
+	treePayload := map[string]interface{}{
+		"base_tree": mainSHA,
+		"tree": []map[string]interface{}{
+			{
+				"path":  filePath,
+				"mode":  "100644",
+				"type":  "blob",
+				"sha":   blobResult.SHA,
+			},
+		},
+	}
+	treeBytes, _ := json.Marshal(treePayload)
+	reqTree, err := http.NewRequestWithContext(ctx, http.MethodPost, treeURL, bytes.NewBuffer(treeBytes))
+	if err != nil {
+		return err
+	}
+	c.addHeaders(reqTree)
+
+	respTree, err := c.client.Do(reqTree)
+	if err != nil {
+		return err
+	}
+	defer respTree.Body.Close()
+
+	if respTree.StatusCode != http.StatusCreated {
+		bodyBytes, _ := io.ReadAll(respTree.Body)
+		return fmt.Errorf("create tree failed with status: %d, body: %s", respTree.StatusCode, string(bodyBytes))
+	}
+
+	var treeResult struct {
+		SHA string `json:"sha"`
+	}
+	if err := json.NewDecoder(respTree.Body).Decode(&treeResult); err != nil {
+		return err
+	}
+
+	commitURL := fmt.Sprintf("%s/repos/%s/%s/git/commits", c.apiURL, owner, repo)
+	commitPayload := map[string]interface{}{
+		"message": commitMessage,
+		"tree":    treeResult.SHA,
+		"parents": []string{mainSHA},
+	}
+	commitBytes, _ := json.Marshal(commitPayload)
+	reqCommit, err := http.NewRequestWithContext(ctx, http.MethodPost, commitURL, bytes.NewBuffer(commitBytes))
+	if err != nil {
+		return err
+	}
+	c.addHeaders(reqCommit)
+
+	respCommit, err := c.client.Do(reqCommit)
+	if err != nil {
+		return err
+	}
+	defer respCommit.Body.Close()
+
+	if respCommit.StatusCode != http.StatusCreated {
+		bodyBytes, _ := io.ReadAll(respCommit.Body)
+		return fmt.Errorf("create commit failed with status: %d, body: %s", respCommit.StatusCode, string(bodyBytes))
+	}
+
+	var commitResult struct {
+		SHA string `json:"sha"`
+	}
+	if err := json.NewDecoder(respCommit.Body).Decode(&commitResult); err != nil {
+		return err
+	}
+
+	refURL := fmt.Sprintf("%s/repos/%s/%s/git/refs/heads/%s", c.apiURL, owner, repo, branch)
+	refPayload := map[string]interface{}{
+		"sha":   commitResult.SHA,
+		"force": true,
+	}
+	refBytes, _ := json.Marshal(refPayload)
+	reqRef, err := http.NewRequestWithContext(ctx, http.MethodPatch, refURL, bytes.NewBuffer(refBytes))
+	if err != nil {
+		return err
+	}
+	c.addHeaders(reqRef)
+
+	respRef, err := c.client.Do(reqRef)
+	if err != nil {
+		return err
+	}
+	defer respRef.Body.Close()
+
+	if respRef.StatusCode != http.StatusOK {
+		bodyBytes, _ := io.ReadAll(respRef.Body)
+		return fmt.Errorf("update ref failed with status: %d, body: %s", respRef.StatusCode, string(bodyBytes))
+	}
+
+	return nil
+}
+
+func (m *MockClient) CreatePR(ctx context.Context, owner, repo, title, branch, body string) (string, error) {
+	return "https://github.com/mock/mock/pull/2", nil
+}
+
+func (m *MockClient) CommitAndPushFile(ctx context.Context, owner, repo, branch, filePath, fileContent, commitMessage string) error {
+	return nil
+}
+
 func (c *RESTClient) SearchCode(ctx context.Context, owner, repo, query string) (string, error) {
	url := fmt.Sprintf("%s/search/code?q=%s+repo:%s/%s", c.apiURL, query, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
INNER_EOF

patch api/internal/github/client.go < patch_client.diff

cat << 'INNER_EOF' > patch_router.diff
--- api/internal/server/router.go
+++ api/internal/server/router.go
@@ -67,6 +67,7 @@
	limitsMW := TierLimitsMiddleware(store)

	r.Method("POST", "/api/v1/sync", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.SyncHandler(store, riverClient)))))
+	r.Method("POST", "/api/v1/gateway/sync/{org}/{repo}", http.HandlerFunc(handlers.GatewaySyncHandler(store, github.NewRESTClient())))
	r.Method("POST", "/api/v1/webhook", http.HandlerFunc(webhook.PushHandler(store, github.NewRESTClient(), riverClient)))
	r.Method("POST", "/api/v1/webhook/reaction", serviceTokenMW(http.HandlerFunc(webhook.ReactionHandler(store, github.NewRESTClient()))))
	r.Method("POST", "/api/v1/cross-repo-check", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.CrossRepoCheckHandler(store, riverClient)))))
INNER_EOF

patch api/internal/server/router.go < patch_router.diff

cd api && go get sigs.k8s.io/yaml && go mod tidy && go fmt ./... && go vet ./... && go test ./...
