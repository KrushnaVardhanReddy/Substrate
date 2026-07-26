#!/bin/bash
set -e

echo "🚀 Starting Full-Stack E2E Test Harness (PGlite)..."

# Kill any zombie processes on ports before starting
echo "🧹 Clearing any zombie processes..."
fuser -k 8090/tcp 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
fuser -k 54320/tcp 2>/dev/null || true
fuser -k 5173/tcp 2>/dev/null || true
sleep 1

# Trap cleanup to run on exit or error
trap 'echo "🧹 Cleaning up background processes..."; kill $FRONTEND_PID $API_PID $ENGINE_PID $PGLITE_PID 2>/dev/null || true; fuser -k 8090/tcp 2>/dev/null || true; fuser -k 8080/tcp 2>/dev/null || true; fuser -k 54320/tcp 2>/dev/null || true; fuser -k 5173/tcp 2>/dev/null || true' EXIT

# 1. Start PGlite Database
echo "📦 Starting PGlite Server..."
cd scripts/e2e
npm install --no-audit --no-fund --legacy-peer-deps > /dev/null 2>&1
node pglite_server.js > pglite.log 2>&1 &
PGLITE_PID=$!
cd ../..

# Wait for PGlite to be ready (health check)
echo "⏳ Waiting for PGlite to initialize..."
for i in $(seq 1 15); do
  if psql "postgres://postgres:postgres@localhost:54320/postgres" -c "SELECT 1" > /dev/null 2>&1; then
    echo "✅ PGlite is ready!"
    break
  fi
  sleep 1
done

# 2. Start Go API
echo "⚙️ Starting Go API Server..."
export DATABASE_URL="postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
export PORT="8090"
export REGISTRY_API_TOKEN="local-dev-token"
export INTERNAL_SERVICE_TOKEN="local-dev-token"
export JWT_SECRET="local-jwt-secret"
export GITHUB_CLIENT_ID="test-client"
export GITHUB_CLIENT_SECRET="test-secret"
export DASHBOARD_URL="http://localhost:5173"
export SKIP_MIGRATIONS="true"
export SKIP_RIVER="false"

cd api
ENVIRONMENT=development go run ./cmd/server > ../api.log 2>&1 &
API_PID=$!
cd ..

# 2.5 Start Go Engine (diff engine on port 8080)
echo "⚙️ Starting Go Engine (Port 8080)..."
cd engine
go run ./cmd/substrate serve > ../engine.log 2>&1 &
ENGINE_PID=$!
cd ..

# Wait for Go API to be ready (health check)
echo "⏳ Waiting for Go API to initialize..."
for i in $(seq 1 30); do
  if curl -s http://localhost:8090/health > /dev/null 2>&1; then
    echo "✅ Go API is ready!"
    break
  fi
  sleep 2
done

# Wait for Go Engine to be ready (health check on port 8080)
echo "⏳ Waiting for Go Engine to initialize..."
for i in $(seq 1 30); do
  if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Go Engine is ready!"
    break
  fi
  sleep 2
done

# 3. Start SvelteKit Frontend
echo "🖥️ Starting SvelteKit Frontend..."
cd dashboard
npm install --no-audit --no-fund --legacy-peer-deps > /dev/null 2>&1
npm run dev -- --port 5173 > frontend.log 2>&1 &
FRONTEND_PID=$!
cd ..

# Wait for Frontend to be ready
echo "⏳ Waiting for Frontend to initialize..."
for i in $(seq 1 30); do
  if curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo "✅ Frontend is ready!"
    break
  fi
  sleep 2
done

echo "✅ All tiers running! Executing Full E2E Test Suite..."

# Run All Go API Tests
echo "🧪 Running Go API Tests (All Phases)..."
cd scripts/e2e
go test -v -p 1 ./...
cd ../..

# Re-seed via API for Playwright UI tests (Go tests clean up the DB)
echo "🌱 Re-seeding via live API for UI tests..."
bash scripts/e2e/seed_via_api.sh

# Verify services are still up before Playwright
echo "🔍 Verifying services before Playwright run..."
curl -s http://localhost:8090/health > /dev/null 2>&1 && echo "  ✅ API :8090 OK" || echo "  ❌ API :8090 DOWN"
curl -s http://localhost:8080/health > /dev/null 2>&1 && echo "  ✅ Engine :8080 OK" || echo "  ❌ Engine :8080 DOWN"
curl -s http://localhost:5173 > /dev/null 2>&1 && echo "  ✅ Frontend :5173 OK" || echo "  ❌ Frontend :5173 DOWN"

# Run All Playwright UI Tests
echo "🧪 Running Playwright UI Tests (All Phases)..."
cd dashboard
npx playwright test tests/e2e --project=chromium
cd ..

# Trap handles cleanup
echo "🎉 E2E Test Run Complete!"
