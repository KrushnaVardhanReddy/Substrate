package proxy

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type SampledRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

type Proxy struct {
	target       *url.URL
	reverseProxy *httputil.ReverseProxy
	sampleRate   float64
	sampleChan   chan<- SampledRequest
}

func New(targetURL string, sampleRate float64, sampleChan chan<- SampledRequest) (*Proxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(target)

	return &Proxy{
		target:       target,
		reverseProxy: rp,
		sampleRate:   sampleRate,
		sampleChan:   sampleChan,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rand.Float64() < p.sampleRate {
		// Sample the request
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			// Restore the body so the reverse proxy can use it
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		select {
		case p.sampleChan <- SampledRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header.Clone(),
			Body:    bodyBytes,
		}:
		default:
			// Channel is full, drop the sample to avoid blocking the proxy
		}
	}

	p.reverseProxy.ServeHTTP(w, r)
}
