# Stage 1: Build SvelteKit dashboard
FROM docker.io/library/node:20-alpine AS node-builder
WORKDIR /app
COPY dashboard/package*.json ./dashboard/
WORKDIR /app/dashboard
RUN npm ci
COPY dashboard/ ./
RUN npm run build

# Stage 2: Build Go API server
FROM docker.io/library/golang:1.22-alpine AS go-builder
WORKDIR /app
# Install necessary tools
RUN apk add --no-cache gcc musl-dev git
# Copy API source code
COPY api/go.mod api/go.sum ./api/
WORKDIR /app/api
RUN go mod download
COPY api/ ./
# Copy dashboard build output into the API directory for //go:embed
COPY --from=node-builder /app/dashboard/build ./ui/build
# Build the single binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /substrate-server ./cmd/server/main.go

# Stage 3: Final Image
FROM docker.io/library/alpine:latest
# Or distroless, but using alpine to match current style
RUN apk --no-cache add ca-certificates bash
WORKDIR /app
COPY --from=go-builder /substrate-server /usr/local/bin/substrate-server

EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/substrate-server"]
