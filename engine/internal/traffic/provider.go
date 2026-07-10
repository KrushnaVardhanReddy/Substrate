package traffic

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Provider interface {
	// GetFieldUsage returns the number of calls for a specific field/path.
	GetFieldUsage(org, repo, fieldPath string, days int) (int64, error)
}

type NoOpProvider struct{}

func (p *NoOpProvider) GetFieldUsage(org, repo, fieldPath string, days int) (int64, error) {
	// Default fallback: if no provider is configured, assume the field is highly used
	// to preserve backwards compatibility (return -1 or a large number).
	return -1, nil
}

type PrometheusProvider struct {
	Endpoint string
}

func (p *PrometheusProvider) GetFieldUsage(org, repo, fieldPath string, days int) (int64, error) {
	if p.Endpoint == "" {
		return -1, nil
	}

	// Construct a mock Prometheus query: sum(rate(http_requests_total{org="...", repo="...", path="..."}[30d]))
	// For this task, we will just make a simple HTTP GET to the prometheus /api/v1/query endpoint.
	query := fmt.Sprintf("sum(increase(http_requests_total{org=\"%s\", repo=\"%s\", path=\"%s\"}[%dd]))", org, repo, fieldPath, days)

	reqURL := fmt.Sprintf("%s/api/v1/query?query=%s", p.Endpoint, url.QueryEscape(query))
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(reqURL)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("prometheus returned %d", resp.StatusCode)
	}

	// Parse Prometheus vector/matrix JSON response
	var promResp struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil && err != io.EOF {
		return -1, err
	}

	if len(promResp.Data.Result) == 0 || len(promResp.Data.Result[0].Value) < 2 {
		// No data = 0 traffic
		return 0, nil
	}

	// Parse the value
	valStr, ok := promResp.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, nil
	}

	var usage int64
	fmt.Sscanf(valStr, "%d", &usage)
	return usage, nil
}
