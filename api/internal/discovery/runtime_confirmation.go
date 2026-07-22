package discovery

import (
	"context"
	"fmt"
	"regexp"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

// OTelSpan represents a simplified OpenTelemetry span relevant to discovery.
type OTelSpan struct {
	ServiceName string `json:"service.name"`
	HTTPURL     string `json:"http.url"`
}

var baseURLRegex = regexp.MustCompile(`^(https?://[^/]+)`)

func ExtractBaseURL(rawURL string) string {
	// e.g. https://api.example.com/v1/users -> https://api.example.com
	matches := baseURLRegex.FindStringSubmatch(rawURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return rawURL
}

func ProcessRuntimeSignals(ctx context.Context, store ports.DiscoveryStore, spans []OTelSpan) error {
	var lastErr error
	for _, span := range spans {
		baseURL := ExtractBaseURL(span.HTTPURL)

		fmt.Printf("Processing span: service=%s, url=%s, base=%s\n", span.ServiceName, span.HTTPURL, baseURL)

		err := store.UpdateDependencyConfidence(ctx, span.ServiceName, baseURL, 20.0)
		if err != nil {
			fmt.Printf("failed to update confidence for %s -> %s: %v\n", span.ServiceName, baseURL, err)
			lastErr = err
		}
	}
	return lastErr
}
