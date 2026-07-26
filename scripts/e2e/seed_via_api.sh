#!/bin/bash
# ============================================================
# seed_via_api.sh — E2E Seed Script using Live API Endpoints
#
# This script seeds the full E2E test dataset by calling the
# REAL API endpoints instead of direct SQL inserts.
#
# What each endpoint does:
#   POST /api/v1/sync  → registers org + consumer_repo + provider_repo
#                        + contract + dependency edge in one shot
#
# Insurance policies have no HTTP POST endpoint (MCP-only), so those
# are still seeded via minimal SQL (just 2 rows).
#
# Usage: bash seed_via_api.sh
# Prerequisites: Go API must be running on :8090
# ============================================================

set -e

API="http://localhost:8090"
TOKEN="local-dev-token"

post_sync() {
  local payload="$1"
  local resp
  resp=$(curl -s -X POST "$API/api/v1/sync" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$payload" 2>&1) || {
    echo "  ⚠️  sync call failed: $resp"
    return 0  # don't abort — next syncs may still succeed
  }
  # Check if response contains an error field or status code
  if echo "$resp" | grep -q "error"; then
    echo "  ⚠️  sync call returned error: $resp"
    return 0
  fi
  echo "  ✅ $resp" | head -c 120
  echo ""
}

echo ""
echo "🌱 Seeding E2E data via live API..."

# ─── mcp-org ──────────────────────────────────────────────────────────────────
# frontend → backend (OpenAPI)
echo "📦 mcp-org: frontend → backend"
post_sync '{
  "installation_id": 123456,
  "org": "mcp-org",
  "consumer_repo": "mcp-org/frontend",
  "consumer_github_repo_id": 10102,
  "commit_sha": "sha-mcp-frontend-001",
  "dependencies": [{
    "provider_repo": "mcp-org/backend",
    "provider_github_repo_id": 10101,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Backend API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# discovery-test-repo → backend
echo "📦 mcp-org: discovery-test-repo → backend"
post_sync '{
  "installation_id": 123456,
  "org": "mcp-org",
  "consumer_repo": "mcp-org/discovery-test-repo",
  "consumer_github_repo_id": 10103,
  "commit_sha": "sha-mcp-discovery-001",
  "dependencies": [{
    "provider_repo": "mcp-org/backend",
    "provider_github_repo_id": 10101,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Backend API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# core-repo as standalone provider (registers it in the graph)
echo "📦 mcp-org: core-repo (standalone provider registration)"
post_sync '{
  "installation_id": 123456,
  "org": "mcp-org",
  "consumer_repo": "mcp-org/frontend",
  "consumer_github_repo_id": 10102,
  "commit_sha": "sha-mcp-core-001",
  "dependencies": [{
    "provider_repo": "mcp-org/core-repo",
    "provider_github_repo_id": 10104,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Core API\n  version: 1.0.0\npaths:\n  /health:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# ─── testorg ──────────────────────────────────────────────────────────────────
# payments-api → users-api
echo "📦 testorg: payments-api → users-api"
post_sync '{
  "installation_id": 234567,
  "org": "testorg",
  "consumer_repo": "testorg/payments-api",
  "consumer_github_repo_id": 20202,
  "commit_sha": "sha-testorg-payments-001",
  "dependencies": [{
    "provider_repo": "testorg/users-api",
    "provider_github_repo_id": 20201,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Users API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# api-gateway → users-api
echo "📦 testorg: api-gateway → users-api"
post_sync '{
  "installation_id": 234567,
  "org": "testorg",
  "consumer_repo": "testorg/api-gateway",
  "consumer_github_repo_id": 20203,
  "commit_sha": "sha-testorg-gateway-001",
  "dependencies": [{
    "provider_repo": "testorg/users-api",
    "provider_github_repo_id": 20201,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Users API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# ─── p3-org ───────────────────────────────────────────────────────────────────
# service-b → service-a
echo "📦 p3-org: service-b → service-a"
post_sync '{
  "installation_id": 345678,
  "org": "p3-org",
  "consumer_repo": "p3-org/service-b",
  "consumer_github_repo_id": 30302,
  "commit_sha": "sha-p3-sb-001",
  "dependencies": [{
    "provider_repo": "p3-org/service-a",
    "provider_github_repo_id": 30301,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Service A\n  version: 1.0.0\npaths:\n  /api:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# service-c → service-b (blast radius chain: a ← b ← c)
echo "📦 p3-org: service-c → service-b"
post_sync '{
  "installation_id": 345678,
  "org": "p3-org",
  "consumer_repo": "p3-org/service-c",
  "consumer_github_repo_id": 30303,
  "commit_sha": "sha-p3-sc-001",
  "dependencies": [{
    "provider_repo": "p3-org/service-b",
    "provider_github_repo_id": 30302,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Service B\n  version: 1.0.0\npaths:\n  /api:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# ─── admin org ────────────────────────────────────────────────────────────────
# repo-b → repo-a (blast radius chain for system-matrix tests)
echo "📦 admin: repo-b → repo-a"
post_sync '{
  "installation_id": 456789,
  "org": "admin",
  "consumer_repo": "admin/repo-b",
  "consumer_github_repo_id": 40402,
  "commit_sha": "sha-admin-rb-001",
  "dependencies": [{
    "provider_repo": "admin/repo-a",
    "provider_github_repo_id": 40401,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Repo A\n  version: 1.0.0\npaths:\n  /api:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

# repo-c → repo-b
echo "📦 admin: repo-c → repo-b"
post_sync '{
  "installation_id": 456789,
  "org": "admin",
  "consumer_repo": "admin/repo-c",
  "consumer_github_repo_id": 40403,
  "commit_sha": "sha-admin-rc-001",
  "dependencies": [{
    "provider_repo": "admin/repo-b",
    "provider_github_repo_id": 40402,
    "schema_type": "openapi",
    "spec_path": "openapi.yaml",
    "branch": "main",
    "raw_content": "openapi: 3.0.0\ninfo:\n  title: Repo B\n  version: 1.0.0\npaths:\n  /api:\n    get:\n      responses:\n        \"200\":\n          description: OK"
  }]
}'

echo ""
echo "⏳ Waiting 3s for River async jobs to process..."
sleep 3

# ─── Insurance policies (no POST HTTP endpoint — minimal SQL) ─────────────────
echo ""
echo "💰 Seeding insurance policies via SQL (no HTTP POST endpoint available)..."

DB_URL="postgres://postgres:postgres@localhost:54320/postgres"

psql "$DB_URL" <<'EOSQL'
-- Resolve org UUIDs from the names registered by sync above
DO $$
DECLARE
  mcp_id   UUID;
  test_id  UUID;
BEGIN
  SELECT id INTO mcp_id  FROM organizations WHERE github_org_name = 'mcp-org'  LIMIT 1;
  SELECT id INTO test_id FROM organizations WHERE github_org_name = 'testorg'  LIMIT 1;

  IF mcp_id IS NOT NULL THEN
    INSERT INTO insurance_policies (org_id, policy_limit_cents)
    VALUES (mcp_id, 5000000)   -- $50,000
    ON CONFLICT DO NOTHING;

    INSERT INTO insurance_claims (org_id, policy_id, github_pr_url, incident_date, status, amount_cents)
    SELECT mcp_id, id, 'https://github.com/mcp-org/repo/pull/10', NOW() - INTERVAL '7 days', 'APPROVED', 15000
    FROM insurance_policies WHERE org_id = mcp_id LIMIT 1
    ON CONFLICT DO NOTHING;
  END IF;

  IF test_id IS NOT NULL THEN
    INSERT INTO insurance_policies (org_id, policy_limit_cents)
    VALUES (test_id, 1000000)  -- $10,000
    ON CONFLICT DO NOTHING;

    INSERT INTO insurance_claims (org_id, policy_id, github_pr_url, incident_date, status, amount_cents)
    SELECT test_id, id, 'https://github.com/testorg/repo/pull/5', NOW() - INTERVAL '3 days', 'PENDING', 50000
    FROM insurance_policies WHERE org_id = test_id LIMIT 1
    ON CONFLICT DO NOTHING;
  END IF;
END$$;
EOSQL

echo "  ✅ Insurance policies seeded"

# ─── Verify graph API ──────────────────────────────────────────────────────────
echo ""
echo "🔍 Verifying graph edges..."
for ORG in mcp-org testorg p3-org admin; do
  COUNT=$(curl -sf "$API/api/v1/graph/$ORG" -H "Authorization: Bearer $TOKEN" 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d) if isinstance(d,list) else 0)" 2>/dev/null || echo "0")
  echo "  📊 $ORG: $COUNT dependency edges"
done

echo ""
echo "✅ API-based seeding complete!"
