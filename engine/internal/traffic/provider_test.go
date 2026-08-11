// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package traffic

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrometheusProvider_GetFieldUsage(t *testing.T) {
	tests := []struct {
		name          string
		handler       http.HandlerFunc
		expectedUsage int64
		expectedErr   bool
	}{
		{
			name: "Successful query returning 0 traffic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Empty result or empty value
				fmt.Fprintln(w, `{"data": {"result": []}}`)
			},
			expectedUsage: 0,
			expectedErr:   false,
		},
		{
			name: "Successful query returning 150 traffic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{"data": {"result": [{"value": [1234567890.123, "150"]}]}}`)
			},
			expectedUsage: 150,
			expectedErr:   false,
		},
		{
			name: "HTTP error returns -1",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, `{"error": "internal server error"}`)
			},
			expectedUsage: -1,
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			provider := &PrometheusProvider{Endpoint: server.URL}
			usage, err := provider.GetFieldUsage("test-org", "test-repo", "/test/path", 30)

			if (err != nil) != tt.expectedErr {
				t.Errorf("GetFieldUsage() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			if usage != tt.expectedUsage {
				t.Errorf("GetFieldUsage() got = %v, want %v", usage, tt.expectedUsage)
			}
		})
	}
}

func TestNoOpProvider_GetFieldUsage(t *testing.T) {
	provider := &NoOpProvider{}
	usage, err := provider.GetFieldUsage("test-org", "test-repo", "/test/path", 30)
	if err != nil {
		t.Errorf("GetFieldUsage() error = %v, expectedErr false", err)
	}
	if usage != -1 {
		t.Errorf("GetFieldUsage() got = %v, want -1", usage)
	}
}
