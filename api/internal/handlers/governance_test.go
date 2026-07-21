package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCELHandler(t *testing.T) {
	origKey := os.Getenv("OPENAI_API_KEY")
	origSubKey := os.Getenv("SUBSTRATE_AI_API_KEY")
	origURL := os.Getenv("SUBSTRATE_AI_BASE_URL")
	defer func() {
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("SUBSTRATE_AI_API_KEY", origSubKey)
		os.Setenv("SUBSTRATE_AI_BASE_URL", origURL)
	}()

	// Fake OpenAI server
	fakeAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")

		celResp := `request.path.matches('^/api/')`
		if len(req.Messages) > 1 && req.Messages[1].Content == "invalid rule" {
			celResp = `invalid syntax <<`
		}

		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "` + celResp + `"
				}
			}]
		}`))
	}))
	defer fakeAI.Close()

	handler := GenerateCELHandler()

	tests := []struct {
		name         string
		reqBody      map[string]interface{}
		setupEnv     func()
		expectedCode int
		expectedCEL  string
		expectedErr  string
	}{
		{
			name: "valid request - no API key fallback",
			reqBody: map[string]interface{}{
				"prompt": "All payment APIs must require authentication",
			},
			setupEnv: func() {
				os.Unsetenv("OPENAI_API_KEY")
				os.Unsetenv("SUBSTRATE_AI_API_KEY")
			},
			expectedCode: http.StatusOK,
			expectedCEL:  `request.path.matches('^/api/payment/') && !request.headers.contains('authorization')`,
		},
		{
			name: "valid request - with API key",
			reqBody: map[string]interface{}{
				"prompt": "some rule",
			},
			setupEnv: func() {
				os.Setenv("OPENAI_API_KEY", "fake-key")
				os.Setenv("SUBSTRATE_AI_BASE_URL", fakeAI.URL+"/v1")
			},
			expectedCode: http.StatusOK,
			expectedCEL:  `request.path.matches('^/api/')`,
		},
		{
			name: "invalid request - invalid syntax from AI",
			reqBody: map[string]interface{}{
				"prompt": "invalid rule",
			},
			setupEnv: func() {
				os.Setenv("OPENAI_API_KEY", "fake-key")
				os.Setenv("SUBSTRATE_AI_BASE_URL", fakeAI.URL+"/v1")
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  "invalid CEL expression generated",
		},
		{
			name: "missing prompt",
			reqBody: map[string]interface{}{
				"prompt": "",
			},
			setupEnv:     func() {},
			expectedCode: http.StatusBadRequest,
			expectedErr:  "prompt is required",
		},
		{
			name:         "invalid json",
			reqBody:      nil,
			setupEnv:     func() {},
			expectedCode: http.StatusBadRequest,
			expectedErr:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			var body []byte
			if tt.reqBody != nil {
				body, _ = json.Marshal(tt.reqBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest("POST", "/api/governance/generate-cel", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			var resp GenerateCELResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			assert.NoError(t, err)

			if tt.expectedCEL != "" {
				assert.Equal(t, tt.expectedCEL, resp.CEL)
			}
			if tt.expectedErr != "" {
				assert.Contains(t, resp.Error, tt.expectedErr)
			}
		})
	}
}
