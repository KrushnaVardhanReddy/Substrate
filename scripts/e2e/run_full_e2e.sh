#!/bin/bash
set -e
set -o pipefail

echo "🚀 Starting Full-Stack E2E Test Harness (PGlite)..."

# Kill any zombie processes on ports before starting
echo "🧹 Clearing any zombie processes..."
fuser -k 8090/tcp 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
fuser -k 54320/tcp 2>/dev/null || true
fuser -k 5173/tcp 2>/dev/null || true
sleep 1

# Trap cleanup to run on exit or error
trap 'echo "🧹 Cleaning up background processes..."; kill $FRONTEND_PID $API_PID $ENGINE_PID $PGLITE_PID 2>/dev/null || true; fuser -k 8090/tcp 2>/dev/null || true; fuser -k 8080/tcp 2>/dev/null || true; fuser -k 54320/tcp 2>/dev/null || true; fuser -k 5173/tcp 2>/dev/null || true; echo "🧹 Tearing down Forgejo..."; podman stop forgejo >/dev/null 2>&1 || true; podman rm forgejo >/dev/null 2>&1 || true' EXIT

echo "🐙 Starting Forgejo Container..."
podman unshare rm -rf "$PWD/forgejo-data" 2>/dev/null || true
mkdir -p "$PWD/forgejo-data"
podman run --replace -d --name forgejo --network host -e USER_UID=1000 -e USER_GID=1000 -e GITEA__security__INSTALL_LOCK=true -e GITEA__database__DB_TYPE=sqlite3 -e GITEA__security__ALLOWED_HOST_LIST="*" -e GITEA__server__HTTP_PORT=3005 -e GITEA__server__SSH_PORT=2225 -e GITEA__security__ENABLE_BASIC_AUTHENTICATION=true -e GITEA__service__ENABLE_BASIC_AUTHENTICATION=true -v "$PWD/forgejo-data:/data" -v /etc/timezone:/etc/timezone:ro -v /etc/localtime:/etc/localtime:ro gitea/gitea:latest

echo "⏳ Waiting for Forgejo to initialize..."
for i in $(seq 1 30); do
  if curl -s http://127.0.0.1:3005/api/v1/version > /dev/null 2>&1; then
    echo "✅ Forgejo is ready!"
    break
  fi
  sleep 2
done

echo "👤 Creating Forgejo admin user via CLI..."
for i in $(seq 1 15); do
  if podman exec forgejo su git -c "gitea admin user create --admin --username adminuser --password 'Admin123!' --email admin@example.com --must-change-password=false"; then
    echo "✅ Admin user created!"
    break
  fi
  echo "Waiting for Gitea DB migrations to finish..."
  sleep 2
done

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

# 2. Start Go API Server
echo "⚙️ Starting Go API Server..."
cd api
export DATABASE_URL="postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
export JWT_SECRET="local-jwt-secret"
export GITHUB_CLIENT_ID="test-client"
export GITHUB_CLIENT_SECRET="test-secret"
export PORT="8090"
export REGISTRY_API_TOKEN="local-dev-token"
export INTERNAL_SERVICE_TOKEN="local-dev-token"
export PGLITE_PORT=54320
export GITHUB_API_URL="http://127.0.0.1:3005/api/v1"
export GITHUB_TOKEN="dummy"
export DASHBOARD_URL="http://localhost:5173"
export SKIP_MIGRATIONS="true"
export SKIP_RIVER="false"
ENVIRONMENT=development go run ./cmd/server > ../api.log 2>&1 &
API_PID=$!
cd ..

# 3. Start Go Engine
echo "⚙️ Starting Go Engine (Port 8080)..."
cd engine
export DATABASE_URL="postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
export API_URL="http://localhost:8090"
go run ./cmd/substrate serve > ../engine.log 2>&1 &
ENGINE_PID=$!
cd ..

# Wait for Go services
echo "⏳ Waiting for Go API to initialize..."
for i in $(seq 1 15); do
  if curl -s http://localhost:8090/health > /dev/null 2>&1; then
    echo "✅ Go API is ready!"
    break
  fi
  sleep 2
done

echo "⏳ Waiting for Go Engine to initialize..."
for i in $(seq 1 15); do
  if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Go Engine is ready!"
    break
  fi
  sleep 2
done

# 4. Start SvelteKit Frontend
echo "🖥️ Starting SvelteKit Frontend..."
cd dashboard
npm install --no-audit --no-fund --legacy-peer-deps > /dev/null 2>&1
export PUBLIC_API_URL="http://localhost:8090"
npm run dev -- --host > ../dashboard.log 2>&1 &
FRONTEND_PID=$!
cd ..

# Wait for Frontend
echo "⏳ Waiting for Frontend to initialize..."
for i in $(seq 1 15); do
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
rm -f e2e_test.logs
go test -v -p 1 ./... 2>&1 | tee e2e_test.logs
cd ../..

# Re-seed via API for Playwright UI tests (Go tests clean up the DB)
echo "🌱 Re-seeding via live API for UI tests..."
bash scripts/e2e/seed_via_api.sh

# Verify services are still up before Playwright
echo "🔍 Verifying services before Playwright run..."
curl -s http://localhost:8090/health > /dev/null 2>&1 && echo "  ✅ API :8090 OK" || echo "  ❌ API :8090 DOWN"
curl -s http://localhost:8080/health > /dev/null 2>&1 && echo "  ✅ Engine :8080 OK" || echo "  ❌ Engine :8080 DOWN"

# 5. Run Playwright UI Tests
echo "🧪 Running Playwright UI Tests..."
cd dashboard
npx playwright test
cd ..

echo "🎉 E2E Test Run Complete!"
