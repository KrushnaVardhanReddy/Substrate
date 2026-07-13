package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/sidecar/internal/validator"
)

type DriftAnomalyPayload struct {
	OrgName      string `json:"org_name"`
	RepoName     string `json:"repo_name"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	ErrorMessage string `json:"error_message"`
}

type Reporter struct {
	substrateURL string
	apiToken     string
	orgName      string
	repoName     string
	httpClient   *http.Client
}

func New(substrateURL, apiToken, orgName, repoName string) *Reporter {
	return &Reporter{
		substrateURL: substrateURL,
		apiToken:     apiToken,
		orgName:      orgName,
		repoName:     repoName,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *Reporter) ReportAnomaly(ctx context.Context, failure validator.ValidationFailure) {
	payload := DriftAnomalyPayload{
		OrgName:      r.orgName,
		RepoName:     r.repoName,
		Method:       failure.Method,
		Path:         failure.Path,
		ErrorMessage: failure.ErrorMessage,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal drift anomaly payload: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/telemetry/drift", r.substrateURL), bytes.NewBuffer(body))
	if err != nil {
		log.Printf("failed to create report request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+r.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		log.Printf("failed to send drift anomaly report: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		log.Printf("failed to report drift anomaly, unexpected status: %d", resp.StatusCode)
	}
}
