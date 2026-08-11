// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package consumers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/consumers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestFetchSchema(t *testing.T) {
	t.Run("public registry url", func(t *testing.T) {
		store := &db.MockStore{
			GetPublicSchemaFunc: func(ctx context.Context, namespace, name, version string) (*db.PublicSchema, error) {
				assert.Equal(t, "stripe", namespace)
				assert.Equal(t, "api", name)
				assert.Equal(t, "v1", version)
				return &db.PublicSchema{SchemaContent: "openapi: 3.0.0"}, nil
			},
		}

		content, err := consumers.FetchSchema(context.Background(), "registry.substrate.io/@stripe/api/v1", store)
		assert.NoError(t, err)
		assert.Equal(t, "openapi: 3.0.0", content)
	})

	t.Run("public registry url - invalid format", func(t *testing.T) {
		store := &db.MockStore{}
		_, err := consumers.FetchSchema(context.Background(), "registry.substrate.io/stripe/api/v1", store)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid registry url format")
	})

	t.Run("public registry url - not found", func(t *testing.T) {
		store := &db.MockStore{
			GetPublicSchemaFunc: func(ctx context.Context, namespace, name, version string) (*db.PublicSchema, error) {
				return nil, errors.New("not found")
			},
		}
		_, err := consumers.FetchSchema(context.Background(), "registry.substrate.io/@stripe/api/v1", store)
		assert.Error(t, err)
	})

	t.Run("fallback http fetch", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("custom schema content"))
		}))
		defer ts.Close()

		store := &db.MockStore{}
		content, err := consumers.FetchSchema(context.Background(), ts.URL, store)
		assert.NoError(t, err)
		assert.Equal(t, "custom schema content", content)
	})
}
