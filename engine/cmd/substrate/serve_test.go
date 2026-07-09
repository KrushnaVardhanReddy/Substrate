package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServeHealth(t *testing.T) {
	mux := setupMux()

	t.Run("GET /health", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %v", rr.Code)
		}
		if rr.Body.String() != `{"status":"ok"}` {
			t.Errorf("Expected body `{\"status\":\"ok\"}`, got %v", rr.Body.String())
		}
	})

	t.Run("POST /health", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/health", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %v", rr.Code)
		}
	})
}

func TestServeDiff(t *testing.T) {
	mux := setupMux()

	t.Run("GET /diff", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/diff", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %v", rr.Code)
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/diff", bytes.NewBuffer([]byte(`{malformed`)))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %v", rr.Code)
		}
	})

	t.Run("Missing base_schema", func(t *testing.T) {
		body := `{"head_schema": "foo", "schema_type": "openapi"}`
		req, _ := http.NewRequest(http.MethodPost, "/diff", bytes.NewBuffer([]byte(body)))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %v", rr.Code)
		}
	})

	t.Run("Missing schema_type", func(t *testing.T) {
		body := `{"base_schema": "foo", "head_schema": "foo"}`
		req, _ := http.NewRequest(http.MethodPost, "/diff", bytes.NewBuffer([]byte(body)))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %v", rr.Code)
		}
	})

	t.Run("Unknown schema_type", func(t *testing.T) {
		body := `{"base_schema": "foo", "head_schema": "foo", "schema_type": "avro"}`
		req, _ := http.NewRequest(http.MethodPost, "/diff", bytes.NewBuffer([]byte(body)))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %v", rr.Code)
		}
	})

	t.Run("OpenAPI valid diff", func(t *testing.T) {
		baseBytes, _ := os.ReadFile("testdata/base.yaml")
		headBytes, _ := os.ReadFile("testdata/rev_warning.yaml")

		reqObj := DiffRequest{
			BaseSchema: string(baseBytes),
			HeadSchema: string(headBytes),
			SchemaType: "openapi",
		}
		body, _ := json.Marshal(reqObj)

		req, _ := http.NewRequest(http.MethodPost, "/diff", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %v, body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("SQL valid diff", func(t *testing.T) {
		baseBytes, _ := os.ReadFile("../../internal/sql/testdata/base_table_renamed.sql")
		headBytes, _ := os.ReadFile("../../internal/sql/testdata/rev_table_renamed.sql")

		reqObj := DiffRequest{
			BaseSchema: string(baseBytes),
			HeadSchema: string(headBytes),
			SchemaType: "sql",
		}
		body, _ := json.Marshal(reqObj)

		req, _ := http.NewRequest(http.MethodPost, "/diff", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %v, body: %s", rr.Code, rr.Body.String())
		}
	})
}
