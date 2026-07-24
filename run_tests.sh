#!/bin/bash
set -e

echo "Starting Postgres..."
make postgres > /dev/null 2>&1

echo "Waiting for Postgres to be ready..."
for i in {1..30}; do
  if pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; then
    echo "Postgres is up!"
    break
  fi
  sleep 1
done
# Give it an extra few seconds to fully initialize
sleep 5

echo "Starting API Server..."
cd /home/krushna/Project/Substrate/api
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/substrate?sslmode=disable" \
REGISTRY_API_TOKEN="local-dev-token" \
INTERNAL_SERVICE_TOKEN="local-dev-token" \
JWT_SECRET="local-jwt-secret" \
GITHUB_CLIENT_ID="mock-client-id" \
GITHUB_CLIENT_SECRET="mock-client-secret" \
DASHBOARD_URL="http://localhost:5173" \
ENVIRONMENT="development" \
go run ./cmd/server/main.go > /home/krushna/Project/Substrate/api_manual.log 2>&1 &
API_PID=$!

echo "Waiting for API to be ready..."
for i in {1..30}; do
  if curl -s http://localhost:8090/health > /dev/null; then
    echo "API is up!"
    break
  fi
  sleep 1
done
sleep 2

echo "Running E2E tests..."
cd /home/krushna/Project/Substrate/scripts/e2e
export GITHUB_TOKEN=mock_token
go test -v -p 1 ./... || true

echo "Tests finished. Cleaning up..."
kill $API_PID
