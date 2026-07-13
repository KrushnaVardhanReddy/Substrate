package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/sidecar/internal/proxy"
	"github.com/KrushnaVardhanReddy/substrate/sidecar/internal/reporter"
	"github.com/KrushnaVardhanReddy/substrate/sidecar/internal/validator"
)

func main() {
	var (
		listenAddr   = flag.String("listen", ":8080", "Address to listen on")
		targetURL    = flag.String("target", "http://localhost:8081", "Target URL to proxy to")
		substrateURL = flag.String("substrate-url", "http://localhost:8090", "URL of the Substrate API")
		orgName      = flag.String("org", "", "Organization name")
		repoName     = flag.String("repo", "", "Repository name")
		apiToken     = flag.String("token", "", "Substrate API Token")
		sampleRate   = flag.Float64("sample-rate", 0.05, "Fraction of requests to sample (0.0 to 1.0)")
	)
	flag.Parse()

	if *orgName == "" || *repoName == "" || *apiToken == "" {
		log.Fatal("missing required flags: -org, -repo, and -token are required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sampleChan := make(chan proxy.SampledRequest, 1000)
	failureChan := make(chan validator.ValidationFailure, 1000)

	// Initialize components
	p, err := proxy.New(*targetURL, *sampleRate, sampleChan)
	if err != nil {
		log.Fatalf("failed to create proxy: %v", err)
	}

	v := validator.New(*substrateURL, *apiToken, *orgName, *repoName, 5*time.Minute, failureChan)
	r := reporter.New(*substrateURL, *apiToken, *orgName, *repoName)

	// Start validator schema polling
	go v.Start(ctx)

	// Start processing sampled requests
	go func() {
		for sample := range sampleChan {
			v.ValidateRequest(ctx, sample)
		}
	}()

	// Start processing validation failures
	go func() {
		for failure := range failureChan {
			r.ReportAnomaly(ctx, failure)
		}
	}()

	// Start proxy server
	server := &http.Server{
		Addr:    *listenAddr,
		Handler: p,
	}

	go func() {
		log.Printf("Starting Substrate proxy on %s, forwarding to %s (sample rate: %.2f)", *listenAddr, *targetURL, *sampleRate)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("proxy server failed: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("proxy shutdown failed: %v", err)
	}
}
