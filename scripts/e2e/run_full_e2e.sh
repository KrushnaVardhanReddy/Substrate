#!/bin/bash
set -e

echo "🚀 Starting Full-Stack E2E Test Harness (PGlite)..."

# Trap cleanup to run on exit or error
trap 'echo "🧹 Cleaning up background processes..."; kill $FRONTEND_PID $API_PID $PGLITE_PID 2>/dev/null || true' EXIT

# 1. Start PGlite Database
echo "📦 Starting PGlite Server..."
cd scripts/e2e
npm install --no-audit --no-fund --legacy-peer-deps > /dev/null 2>&1
node pglite_server.js > pglite.log 2>&1 &
PGLITE_PID=$!
cd ../..

# Wait for PGlite to be ready
echo "⏳ Waiting for PGlite to initialize..."
sleep 3

# 2. Start Go API
echo "⚙️ Starting Go API Server..."
export DATABASE_URL="postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
export PORT="8090"
export REGISTRY_API_TOKEN="local-dev-token"
export JWT_SECRET="test-secret"
export GITHUB_CLIENT_ID="test-client"
export GITHUB_CLIENT_SECRET="test-secret"
export DASHBOARD_URL="http://localhost:5173"
export SKIP_MIGRATIONS="true"
export SKIP_RIVER="true"

cd api
go run ./cmd/server > ../api.log 2>&1 &
API_PID=$!
cd ..

# Wait for API to be ready
echo "⏳ Waiting for Go API to initialize..."
sleep 5

# 3. Start SvelteKit Frontend
echo "🖥️ Starting SvelteKit Frontend..."
cd dashboard
npm install --no-audit --no-fund --legacy-peer-deps > /dev/null 2>&1
npm run dev -- --port 5173 > frontend.log 2>&1 &
FRONTEND_PID=$!
cd ..

# Wait for Frontend to be ready
echo "⏳ Waiting for Frontend to initialize..."
sleep 5

echo "✅ All tiers running! Executing Phase 5 E2E Tests..."

# Run Phase 4 Go API Tests
echo "🧪 Running Go API Tests (Phase 4)..."
cd scripts/e2e
go test -v phase4_api_test.go
cd ../..

# Run Playwright UI Tests
echo "🧪 Running Playwright UI Tests (Phase 4)..."
cd dashboard
npx playwright test tests/e2e/ai_diff.spec.ts --project=chromium
cd ..

# Trap handles cleanup
echo "🎉 E2E Test Run Complete!"
