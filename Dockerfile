# ─── Stage 1: Build SvelteKit frontend ────────────────────────────────────────
FROM node:22-alpine AS node-builder
WORKDIR /app/dashboard
COPY dashboard/package*.json ./
RUN npm ci --prefer-offline
COPY dashboard/ ./
RUN npm run build

# ─── Stage 2: Build Go binary ─────────────────────────────────────────────────
FROM golang:1.23-alpine AS go-builder
WORKDIR /app
COPY api/go.mod api/go.sum ./api/
WORKDIR /app/api
RUN go mod download
COPY api/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/substrate ./cmd/server

# ─── Stage 3: Final minimal image ────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12
COPY --from=go-builder /app/substrate /substrate
COPY --from=node-builder /app/dashboard/build /static
COPY --from=node-builder /app/dashboard/static/engine.wasm /static/engine.wasm
EXPOSE 8090
ENTRYPOINT ["/substrate"]
