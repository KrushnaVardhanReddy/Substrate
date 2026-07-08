FROM docker.io/library/golang:alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app
# We assume the action context is the root of the repository.
# We copy the engine directory and build it.
COPY engine/ ./engine/
WORKDIR /app/engine
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /substrate cmd/substrate/main.go

FROM docker.io/library/alpine:latest
RUN apk --no-cache add ca-certificates bash git

COPY --from=builder /substrate /usr/local/bin/substrate
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
