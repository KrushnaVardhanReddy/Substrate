# P5-T08: Chi Router Migration & Panic Recovery

## Objective
The current `http.NewServeMux()` router implementation (`api/internal/server/router.go`) lacks built-in panic recovery. If a backend scanner (like the OpenAPI or GraphQL diff engines) processes a malicious or oversized "Poison Pill" schema, a nil-pointer or out-of-bounds panic will crash the entire Substrate API server. 

To future-proof the API for enterprise scale and telemetry, we must migrate the routing layer to `github.com/go-chi/chi/v5` and inject `middleware.Recoverer`, `middleware.Logger`, and `middleware.Timeout`.

## 1. Migration Requirements

### Core Router (`api/internal/server/router.go`)
- Run `go get github.com/go-chi/chi/v5` and `go get github.com/go-chi/cors` in the `api/` directory.
- Replace `http.NewServeMux()` with `chi.NewRouter()`.
- The current Go 1.22 routing paths (e.g. `mux.HandleFunc("GET /api/v1/health")`) must be refactored to Chi idiomatic paths (`r.Get("/api/v1/health")`, `r.Post(...)`).

### Middleware Injection
You must apply the following middlewares globally at the top of the router:
1. `middleware.RequestID` - Injects a request ID into the context of each request.
2. `middleware.RealIP` - Sets a real IP from headers.
3. `middleware.Logger` - Standard Chi request logging.
4. `middleware.Recoverer` - **CRITICAL:** Recovers from panics, logs the panic stack trace, and returns an HTTP 500 without crashing the server.
5. `middleware.Timeout(60 * time.Second)` - Prevent any single webhook from hanging the server indefinitely.

### Custom Middlewares (`api/internal/server/middleware.go`)
- We have custom middlewares: `AuthMiddleware`, `ServiceTokenMiddleware`, `TierLimitsMiddleware`, and `corsMiddleware`.
- Chi supports standard `func(http.Handler) http.Handler`, so these do not need logic changes.
- Note: Replace the custom `corsMiddleware` function in `router.go` with the official `github.com/go-chi/cors` package to ensure strict compliance.
- Update `AuthMiddleware` to extract URL parameters. Chi uses `chi.URLParam(r, "org")` instead of the Go 1.22 `r.PathValue("org")`. Ensure you update line 76 of `middleware.go`.

## 2. Success Criteria
1. `cd api && go test ./...` passes perfectly with zero routing errors.
2. The server successfully boots up with `make api` and serves the `/health` endpoint.
3. Code strictly adheres to idiomatic `chi` routing patterns.
