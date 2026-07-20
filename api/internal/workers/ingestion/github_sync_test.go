package ingestion_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers/ingestion"
	"github.com/stretchr/testify/assert"
)

func TestGitHubSyncWorker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stripe/openapi/master/openapi/spec3.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("openapi: 3.0.0\ninfo:\n  title: Stripe API"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	var publishedContent string
	var publishedNamespace string

	store := &db.MockStore{
		PublishPublicSchemaFunc: func(ctx context.Context, namespace, name, version, schemaType, content string) error {
			publishedNamespace = namespace
			publishedContent = content
			return nil
		},
	}

	worker := &ingestion.GitHubSyncWorker{
		Store: store,
	}

	repos := []ingestion.PublicRepoConfig{
		{
			Namespace:  "stripe",
			Name:       "api",
			Version:    "v1",
			SchemaType: "openapi",
			URL:        ts.URL + "/stripe/openapi/master/openapi/spec3.yaml",
		},
	}

	err := worker.SyncRepositories(context.Background(), repos)
	assert.NoError(t, err)

	assert.Equal(t, "stripe", publishedNamespace)
	assert.Contains(t, publishedContent, "Stripe API")
}
