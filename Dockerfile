# ─── Stage 1: Build SvelteKit frontend ────────────────────────────────────────
FROM node:22-alpine AS node-builder
WORKDIR /app/dashboard
COPY dashboard/package*.json ./
RUN npm ci --prefer-offline
COPY dashboard/ ./
RUN npm run build
# Output: /app/dashboard/build/

# ─── Stage 2: Build Go binary ─────────────────────────────────────────────────
FROM docker.io/library/golang:alpine AS go-builder
WORKDIR /app
COPY engine/ ./engine/
COPY api/ ./api/
WORKDIR /app/api
RUN go mod download
# Build static binary (no CGO for distroless compatibility)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/substrate ./cmd/server
# Output: /app/substrate

# ─── Stage 3: Final minimal image ────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12
# Copy the compiled Go API binary
COPY --from=go-builder /app/substrate /substrate
# Copy SvelteKit built static assets to /static (served at "/" by the Go binary)
COPY --from=node-builder /app/dashboard/build /static
# Copy the Go WASM binary (built separately and checked into dashboard/static/)
COPY --from=node-builder /app/dashboard/static/engine.wasm /static/engine.wasm
EXPOSE 8090
ENTRYPOINT ["/substrate"]
