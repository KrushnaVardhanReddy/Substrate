FROM golang:alpine AS builder

WORKDIR /app
# We assume the action context is the root of the repository.
# We copy the engine directory and build it.
COPY engine/ ./engine/
WORKDIR /app/engine
RUN go mod download
RUN go build -o /substrate cmd/substrate/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates bash

COPY --from=builder /substrate /usr/local/bin/substrate
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
