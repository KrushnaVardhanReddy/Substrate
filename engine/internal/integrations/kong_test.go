// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKongClient_PushServiceSpec(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Kong-Admin-Token") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/services/my-service/document" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewKongClient(ts.URL, "secret")

	err := client.PushServiceSpec(context.Background(), "my-service", []byte(`{"openapi": "3.0.0"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKongClient_PushServiceSpec_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewKongClient(ts.URL, "secret")

	err := client.PushServiceSpec(context.Background(), "my-service", []byte(`{"openapi": "3.0.0"}`))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
