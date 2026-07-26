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

# Copy the frontend built assets into the Go source tree before building
# so that //go:embed can bundle them into the single binary
RUN mkdir -p /app/api/internal/server/static
COPY --from=node-builder /app/dashboard/build/ /app/api/internal/server/static/
# Copy the WASM binary as well, since it's loaded from root
COPY --from=node-builder /app/dashboard/static/engine.wasm /app/api/internal/server/static/engine.wasm

WORKDIR /app/api
RUN go mod download
# Build static binary (no CGO for distroless compatibility)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/substrate ./cmd/server
# Output: /app/substrate

# ─── Stage 3: Final minimal image ────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12
# Copy the single compiled Go API binary (which now contains the embedded UI)
COPY --from=go-builder /app/substrate /substrate

EXPOSE 8090
ENTRYPOINT ["/substrate"]
