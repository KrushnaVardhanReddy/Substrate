package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/sidecar/internal/proxy"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

type ValidationFailure struct {
	Method       string
	Path         string
	ErrorMessage string
}

type Validator struct {
	substrateURL   string
	apiToken       string
	orgName        string
	repoName       string
	schema         *openapi3.T
	router         routers.Router
	mu             sync.RWMutex
	failureChan    chan<- ValidationFailure
	httpClient     *http.Client
	pollInterval   time.Duration
}

func New(substrateURL, apiToken, orgName, repoName string, pollInterval time.Duration, failureChan chan<- ValidationFailure) *Validator {
	return &Validator{
		substrateURL: substrateURL,
		apiToken:     apiToken,
		orgName:      orgName,
		repoName:     repoName,
		failureChan:  failureChan,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		pollInterval: pollInterval,
	}
}

func (v *Validator) Start(ctx context.Context) {
	// Initial fetch
	v.fetchSchema(ctx)

	ticker := time.NewTicker(v.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.fetchSchema(ctx)
		}
	}
}

func (v *Validator) fetchSchema(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/schema/%s/%s", v.substrateURL, v.orgName, v.repoName), nil)
	if err != nil {
		log.Printf("failed to create schema request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+v.apiToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		log.Printf("failed to fetch schema: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to fetch schema, status: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read schema body: %v", err)
		return
	}

	type schemaResponse struct {
		Schema string `json:"schema"`
	}
	var sr schemaResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		log.Printf("failed to unmarshal schema response: %v", err)
		return
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(sr.Schema))
	if err != nil {
		log.Printf("failed to parse schema: %v", err)
		return
	}

	if err := doc.Validate(ctx); err != nil {
		log.Printf("fetched schema is invalid: %v", err)
		return
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		log.Printf("failed to create router from schema: %v", err)
		return
	}

	v.mu.Lock()
	v.schema = doc
	v.router = router
	v.mu.Unlock()
	log.Println("successfully updated schema")
}

func (v *Validator) ValidateRequest(ctx context.Context, sample proxy.SampledRequest) {
	v.mu.RLock()
	router := v.router
	v.mu.RUnlock()

	if router == nil {
		// Schema not yet loaded
		return
	}

	// Create a dummy http.Request to use with the router and filter
	req, err := http.NewRequest(sample.Method, sample.Path, bytes.NewBuffer(sample.Body))
	if err != nil {
		return
	}
	req.Header = sample.Headers

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		// Route not found, this is a drift anomaly
		select {
		case v.failureChan <- ValidationFailure{
			Method:       sample.Method,
			Path:         sample.Path,
			ErrorMessage: fmt.Sprintf("endpoint not found in schema: %v", err),
		}:
		default:
		}
		return
	}

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		select {
		case v.failureChan <- ValidationFailure{
			Method:       sample.Method,
			Path:         sample.Path,
			ErrorMessage: fmt.Sprintf("request validation failed: %v", err),
		}:
		default:
		}
	}
}
