# ─── Stage 1: Build Everything via Make ────────────────────────────────────────
FROM docker.io/library/golang:1.23-alpine AS builder

# Install Node.js, npm, and make
RUN apk add --no-cache nodejs npm make git

WORKDIR /app
COPY . .

# Run the full production build (WASM -> SvelteKit -> Go embedded binary)
RUN make build-prod

# ─── Stage 2: Final minimal image ────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

# The single Go binary now contains the entire dashboard and WASM engine inside it!
COPY --from=builder /app/substrate /substrate

EXPOSE 8090
ENTRYPOINT ["/substrate"]
