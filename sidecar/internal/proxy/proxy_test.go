package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxySampling(t *testing.T) {
	// Dummy target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer targetServer.Close()

	sampleChan := make(chan SampledRequest, 10)

	// Test 100% sample rate
	p, err := New(targetServer.URL, 1.0, sampleChan)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	reqBody := []byte(`{"test": "data"}`)
	req := httptest.NewRequest(http.MethodPost, "/test-path", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	p.ServeHTTP(rr, req)

	// Check response from target
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check sample channel
	select {
	case sample := <-sampleChan:
		if sample.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", sample.Method)
		}
		if sample.Path != "/test-path" {
			t.Errorf("expected path /test-path, got %s", sample.Path)
		}
		if !bytes.Equal(sample.Body, reqBody) {
			t.Errorf("expected body %s, got %s", string(reqBody), string(sample.Body))
		}
	default:
		t.Error("expected a sampled request in channel, got none")
	}

	// Test 0% sample rate
	p0, _ := New(targetServer.URL, 0.0, sampleChan)
	req2 := httptest.NewRequest(http.MethodGet, "/test-no-sample", nil)
	rr2 := httptest.NewRecorder()

	p0.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr2.Code)
	}

	select {
	case <-sampleChan:
		t.Error("expected no sampled request in channel")
	default:
		// Passed
	}
}
