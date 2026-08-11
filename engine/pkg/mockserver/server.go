// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package mockserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

type SchemaResponse struct {
	Schema     string `json:"schema"`
	SchemaType string `json:"schema_type"`
}

func StartMockServer(ctx context.Context, timestamp string, port int) error {
	fmt.Printf("Starting mock server for timestamp %s on port %d...\n", timestamp, port)

	// Fetch schema from registry
	registryURL := viper.GetString("REGISTRY_API_URL")
	if registryURL == "" {
		registryURL = "http://localhost:8090" // Default API registry URL
	}

	owner := viper.GetString("SUBSTRATE_OWNER")
	if owner == "" {
		owner = "unknown"
	}
	repo := viper.GetString("SUBSTRATE_REPO")
	if repo == "" {
		repo = "unknown"
	}

	// Ensure we pass the timestamp as a query parameter to fetch the historical schema
	schemaEndpoint := fmt.Sprintf("%s/api/v1/schema/%s/%s?timestamp=%s", registryURL, owner, repo, timestamp)

	var schemaData string
	var schemaType string

	req, err := http.NewRequestWithContext(ctx, "GET", schemaEndpoint, nil)
	if err == nil {
		token := viper.GetString("REGISTRY_API_TOKEN")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				var schemaResp SchemaResponse
				if err := json.Unmarshal(body, &schemaResp); err == nil {
					schemaData = schemaResp.Schema
					schemaType = schemaResp.SchemaType
				}
			}
		}
	}

	var doc *openapi3.T
	if schemaType == "openapi" && schemaData != "" {
		loader := openapi3.NewLoader()
		parsedDoc, err := loader.LoadFromData([]byte(schemaData))
		if err == nil {
			doc = parsedDoc
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Very simple mocking logic based on OpenAPI paths
		if doc != nil && doc.Paths != nil {
			pathItem := doc.Paths.Find(r.URL.Path)
			if pathItem != nil {
				var op *openapi3.Operation
				if r.Method == http.MethodGet {
					op = pathItem.Get
				} else if r.Method == http.MethodPost {
					op = pathItem.Post
				} else if r.Method == http.MethodPut {
					op = pathItem.Put
				} else if r.Method == http.MethodDelete {
					op = pathItem.Delete
				}

				if op != nil {
					// Try to find a 200 response
					respRef := op.Responses.Value("200")
					if respRef != nil && respRef.Value != nil {
						content := respRef.Value.Content.Get("application/json")
						if content != nil && content.Schema != nil && content.Schema.Value != nil {
							// Return a dummy object matching the schema type
							if content.Schema.Value.Type.Is("object") || content.Schema.Value.Type.Is("") {
								w.WriteHeader(http.StatusOK)
								w.Write([]byte(`{}`))
								return
							} else if content.Schema.Value.Type.Is("array") {
								w.WriteHeader(http.StatusOK)
								w.Write([]byte(`[]`))
								return
							}
						}
					}
					// If no specific response mocked, return an empty 200 JSON object
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
					return
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		if schemaData != "" {
			w.Write([]byte(fmt.Sprintf(`{"message": "Mock data matching schema for %s"}`, timestamp)))
		} else {
			// Fallback mock response
			w.Write([]byte(fmt.Sprintf(`{"message": "Mock server response for %s"}`, timestamp)))
		}
	})

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}
