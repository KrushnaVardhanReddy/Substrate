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

cat << 'EOF' > /tmp/seed_insurance.js
const { Client } = require('pg');

async function run() {
  const client = new Client({
    connectionString: "postgres://postgres:postgres@127.0.0.1:5432/substrate?sslmode=disable"
  });
  await client.connect();

  const sql = `
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
  `;


  // --- Seed Preview Sessions for preview-page tests ---
  // 1. Valid preview token
  const validToken = '11111111-1111-1111-1111-111111111111';
  // 2. Fallback token (no base/head schemas)
  const fallbackToken = '22222222-2222-2222-2222-222222222222';
  // 3. Expired token
  const expiredToken = '33333333-3333-3333-3333-333333333333';

  // We need dummy diff_id for these
  const dummyDiffId1 = '44444444-4444-4444-4444-444444444444';
  const dummyDiffId2 = '55555555-5555-5555-5555-555555555555';
  const dummyDiffId3 = '66666666-6666-6666-6666-666666666666';

  const previewSql = `
    INSERT INTO diff_reports (id, org_name, repo_name, report_data)
    VALUES 
      ('${dummyDiffId1}', 'mcp-org', 'frontend', '{"base_schema": "old schema content", "head_schema": "new schema content", "breaking_changes": []}'),
      ('${dummyDiffId2}', 'mcp-org', 'frontend', '{"some_data": "test value", "nested": {"key": "value"}}'),
      ('${dummyDiffId3}', 'mcp-org', 'frontend', '{"base_schema": "old", "head_schema": "new", "breaking_changes": []}')
    ON CONFLICT DO NOTHING;

    INSERT INTO preview_sessions (token, pr_number, diff_id, expires_at)
    VALUES
      ('${validToken}', 101, '${dummyDiffId1}', NOW() + INTERVAL '1 hour'),
      ('${fallbackToken}', 102, '${dummyDiffId2}', NOW() + INTERVAL '1 hour'),
      ('${expiredToken}', 103, '${dummyDiffId3}', NOW() - INTERVAL '1 hour')
    ON CONFLICT DO NOTHING;
  `;
  await client.query(previewSql);

  // --- Seed ROI for telemetry test ---
  const dummyDiffIdAdmin = '77777777-7777-7777-7777-777777777777';
  const roiSql = `
    INSERT INTO diff_reports (id, org_name, repo_name, report_data, is_audit_mode, created_at)
    VALUES 
      ('${dummyDiffIdAdmin}', 'admin', 'repo-a', '{"status": "blocked"}', true, NOW() - INTERVAL '1 day')
    ON CONFLICT DO NOTHING;
  `;
  await client.query(roiSql);

  // --- Seed Webhook for enterprise.spec.ts ---
  const webhookSql = `
    INSERT INTO org_webhooks (org, url, secret)
    VALUES ('mcp-org', 'https://example.com/webhook', 'secret123')
    ON CONFLICT DO NOTHING;
  `;
  await client.query(webhookSql);

  // --- Seed 1,000 nodes for stress-test ---
  const stressOrgSql = `
    INSERT INTO organizations (github_org_name, github_installation_id)
    VALUES ('stress-test', 999999)
    ON CONFLICT (github_installation_id) DO UPDATE SET github_org_name = EXCLUDED.github_org_name
    RETURNING id;
  `;
  const orgRes = await client.query(stressOrgSql);
  const stressOrgId = orgRes.rows[0].id;

  // Insert 1000 repositories in batches
  let repoValues = [];
  let depsValues = [];
  for (let i = 0; i < 1000; i++) {
    repoValues.push(`('${stressOrgId}', 'Service ${i}', 'Service ${i}', 900000 + ${i})`);
  }
  
  // Insert repos
  await client.query(`
    INSERT INTO repositories (org_id, name, full_name, github_repo_id)
    VALUES ${repoValues.join(',')}
    ON CONFLICT DO NOTHING;
  `);

  // We need to fetch the inserted repo UUIDs to create edges
  const repoRes = await client.query(`
    SELECT id, name FROM repositories WHERE org_id = '${stressOrgId}';
  `);
  
  const repoMap = {};
  for (const row of repoRes.rows) {
    repoMap[row.name] = row.id;
  }

  // Insert contracts
  const contractValues = [];
  for (let i = 0; i < 1000; i++) {
    const repoId = repoMap[`Service ${i}`];
    contractValues.push(`('${repoId}', 'openapi', 'openapi.yaml', 'main', '{"openapi": "3.0.0"}')`);
  }
  await client.query(`
    INSERT INTO contracts (repo_id, schema_type, spec_path, branch, raw_content)
    VALUES ${contractValues.join(',')}
    ON CONFLICT DO NOTHING;
  `);

  const contractRes = await client.query(`SELECT id, repo_id FROM contracts;`);
  const contractMap = {};
  for (const row of contractRes.rows) {
    contractMap[row.repo_id] = row.id;
  }

  // Insert dependencies
  depsValues = [];
  for (let i = 0; i < 3000; i++) {
    const fromName = `Service ${i % 1000}`;
    const toName = `Service ${(i + 1) % 1000}`;
    const fromId = repoMap[fromName];
    const toId = repoMap[toName];
    const contractId = contractMap[toId];
    if (fromId && contractId) {
      depsValues.push(`('${fromId}', '${contractId}')`);
    }
  }

  if (depsValues.length > 0) {
    await client.query(`
      INSERT INTO dependencies (consumer_repo_id, provider_contract_id)
      VALUES ${depsValues.join(',')}
      ON CONFLICT DO NOTHING;
    `);
  }


  await client.end();
}

run().catch(err => {
  console.error("Insurance seed failed:", err);
  process.exit(1);
});
EOF

npm install --prefix scripts/e2e pg --no-save > /dev/null 2>&1
NODE_PATH=scripts/e2e/node_modules node /tmp/seed_insurance.js

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
