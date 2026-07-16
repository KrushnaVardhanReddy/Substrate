# Spec: P12-T08 — V2.0 Production Cutover

## 1. Overview
Bundle the SvelteKit static assets and the Go WASM binary into the single Go binary deployment so the entire V2.0 platform ships as one self-contained Docker image.

## 2. Owner
**Jules** (Go Backend / DevOps)

## 3. Files Modified
- `Dockerfile` ← update build stages
- `Makefile` ← add `build-prod` target
- `.github/workflows/` ← gate Docker push on full test suite

## 4. Requirements

### Dockerfile — Multi-Stage Build
The final image must:
1. **Stage 1 (node-builder):** Run `npm run build --prefix dashboard` to produce `dashboard/build/`.
2. **Stage 2 (go-builder):** Run `go build -o /substrate .` with `CGO_ENABLED=0 GOOS=linux`.
3. **Stage 3 (final):** Copy the Go binary + `dashboard/build/` + `dashboard/static/engine.wasm` into the final slim image.

```dockerfile
# Stage 1: Build SvelteKit
FROM node:22-alpine AS node-builder
WORKDIR /app/dashboard
COPY dashboard/package*.json ./
RUN npm ci
COPY dashboard/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.23-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o substrate .

# Stage 3: Final image
FROM gcr.io/distroless/static-debian12
COPY --from=go-builder /app/substrate /substrate
COPY --from=node-builder /app/dashboard/build /static
COPY --from=node-builder /app/dashboard/static/engine.wasm /static/engine.wasm
EXPOSE 8090
ENTRYPOINT ["/substrate"]
```

### Makefile — `build-prod` Target
Add:
```makefile
build-prod:
	cd dashboard && npm ci && npm run build
	CGO_ENABLED=0 GOOS=linux go build -o substrate .

test-all:
	go test -race ./...
	cd dashboard && npx playwright test
```

### CI Workflow Update
In `.github/workflows/release.yml` (or equivalent):
1. Add a step that runs `make test-all` before `docker build`.
2. Gate the `docker push` step on `make test-all` succeeding.
3. Add `engine.wasm` to the build context (ensure it is not in `.dockerignore`).

## 5. Technical Constraints
- The Go binary MUST serve `dashboard/build/` as a static file tree at route `/`.
- Verify this by checking the existing static file server middleware in `main.go` or `api/server.go`.
- The final Docker image size MUST be < 100MB (use distroless base).
- `engine.wasm` must be accessible at `/engine.wasm` from the browser (same origin as the dashboard).

## 6. Success Criteria
- `docker build -t substrate:v2 .` completes successfully.
- `docker run -p 8090:8090 substrate:v2` serves `/` (returns 200 with SvelteKit HTML).
- `docker run -p 8090:8090 substrate:v2` serves `/engine.wasm` (returns 200 with `application/wasm` content-type).
- `docker run -p 8090:8090 substrate:v2` serves `/api/v1/events` (returns 200 with `text/event-stream`).
- Docker image size is under 100MB.
