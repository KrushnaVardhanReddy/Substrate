-- ============================================================
-- seed_mcp.sql — Full E2E seed for Playwright UI tests
-- Run after Go API tests (which wipe tables) and before Playwright.
-- ============================================================

-- ── Organizations ────────────────────────────────────────────
INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES
  ('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org'),
  ('22222222-2222-2222-2222-222222222222', 234567, 'testorg'),
  ('33333333-3333-3333-3333-333333333333', 345678, 'p3-org'),
  ('44444444-4444-4444-4444-444444444444', 456789, 'admin')
ON CONFLICT (id) DO NOTHING;

-- ── Repositories ─────────────────────────────────────────────
INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES
  -- mcp-org repos
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 10101, 'backend',             'mcp-org/backend'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 10102, 'frontend',            'mcp-org/frontend'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', '11111111-1111-1111-1111-111111111111', 10103, 'discovery-test-repo', 'mcp-org/discovery-test-repo'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 10104, 'core-repo',           'mcp-org/core-repo'),
  -- testorg repos (for matrix + insurance tests)
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '22222222-2222-2222-2222-222222222222', 20201, 'users-api',           'testorg/users-api'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', '22222222-2222-2222-2222-222222222222', 20202, 'payments-api',        'testorg/payments-api'),
  ('11111111-1111-1111-1111-111111111112', '22222222-2222-2222-2222-222222222222', 20203, 'api-gateway',         'testorg/api-gateway'),
  -- p3-org repos (for graph.spec.ts)
  ('22222222-2222-2222-2222-222222222223', '33333333-3333-3333-3333-333333333333', 30301, 'service-a',           'p3-org/service-a'),
  ('33333333-3333-3333-3333-333333333334', '33333333-3333-3333-3333-333333333333', 30302, 'service-b',           'p3-org/service-b'),
  ('44444444-4444-4444-4444-444444444445', '33333333-3333-3333-3333-333333333333', 30303, 'service-c',           'p3-org/service-c'),
  -- admin org repos (for blast-radius spec)
  ('55555555-5555-5555-5555-555555555556', '44444444-4444-4444-4444-444444444444', 40401, 'repo-a',              'admin/repo-a'),
  ('66666666-6666-6666-6666-666666666667', '44444444-4444-4444-4444-444444444444', 40402, 'repo-b',              'admin/repo-b'),
  ('77777777-7777-7777-7777-777777777778', '44444444-4444-4444-4444-444444444444', 40403, 'repo-c',              'admin/repo-c')
ON CONFLICT (id) DO NOTHING;

-- ── Contracts (schemas) ──────────────────────────────────────
INSERT INTO contracts (id, repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES
  -- mcp-org/backend OpenAPI contract
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaab', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'openapi', 'openapi.yaml', 'main', 'sha-backend-001',
   'openapi: 3.0.0\ninfo:\n  title: Backend API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        "200":\n          description: OK'),
  -- mcp-org/core-repo contract
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbc', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'openapi', 'openapi.yaml', 'main', 'sha-core-001',
   'openapi: 3.0.0\ninfo:\n  title: Core API\n  version: 1.0.0\npaths:\n  /health:\n    get:\n      responses:\n        "200":\n          description: OK'),
  -- testorg/users-api contract
  ('cccccccc-cccc-cccc-cccc-cccccccccccd', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'openapi', 'openapi.yaml', 'main', 'sha-users-001',
   'openapi: 3.0.0\ninfo:\n  title: Users API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        "200":\n          description: OK'),
  -- p3-org/service-a contract
  ('dddddddd-dddd-dddd-dddd-ddddddddddde', '22222222-2222-2222-2222-222222222223', 'openapi', 'openapi.yaml', 'main', 'sha-sa-001',
   'openapi: 3.0.0\ninfo:\n  title: Service A\n  version: 1.0.0\npaths:\n  /api:\n    get:\n      responses:\n        "200":\n          description: OK')
ON CONFLICT (id) DO NOTHING;

-- ── Dependencies (edges) ─────────────────────────────────────
INSERT INTO dependencies (id, consumer_repo_id, provider_contract_id) VALUES
  -- mcp-org: frontend → backend
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaac', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaab'),
  -- testorg: payments-api → users-api
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbd', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'cccccccc-cccc-cccc-cccc-cccccccccccd'),
  -- testorg: api-gateway → users-api
  ('cccccccc-cccc-cccc-cccc-ccccccccccce', '11111111-1111-1111-1111-111111111112', 'cccccccc-cccc-cccc-cccc-cccccccccccd'),
  -- p3-org: service-b → service-a
  ('dddddddd-dddd-dddd-dddd-dddddddddddf', '33333333-3333-3333-3333-333333333334', 'dddddddd-dddd-dddd-dddd-ddddddddddde'),
  -- p3-org: service-c → service-a (blast radius chain)
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeef', '44444444-4444-4444-4444-444444444445', 'dddddddd-dddd-dddd-dddd-ddddddddddde')
ON CONFLICT (id) DO NOTHING;

-- ── Insurance (for phase15 + settings-insurance tests) ───────
-- mcp-org policy  ($50,000 limit)
INSERT INTO insurance_policies (id, org_id, policy_limit_cents) VALUES
  ('11111111-1111-1111-1111-11111111110a', '11111111-1111-1111-1111-111111111111', 5000000)
ON CONFLICT (id) DO NOTHING;

-- mcp-org claim (APPROVED, $150)
INSERT INTO insurance_claims (id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents) VALUES
  ('11111111-1111-1111-1111-11111111110b', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-11111111110a',
   'https://github.com/mcp-org/repo/pull/10', NOW() - INTERVAL '7 days', 'APPROVED', 15000)
ON CONFLICT (id) DO NOTHING;

-- testorg policy ($10,000 limit, with id POL-123 in note — actual id is UUID but displayed in UI)
INSERT INTO insurance_policies (id, org_id, policy_limit_cents) VALUES
  ('22222222-2222-2222-2222-22222222220a', '22222222-2222-2222-2222-222222222222', 1000000)
ON CONFLICT (id) DO NOTHING;

-- testorg claim (PENDING, $500)
INSERT INTO insurance_claims (id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents) VALUES
  ('22222222-2222-2222-2222-22222222220b', '22222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-22222222220a',
   'https://github.com/testorg/repo/pull/5', NOW() - INTERVAL '3 days', 'PENDING', 50000)
ON CONFLICT (id) DO NOTHING;
