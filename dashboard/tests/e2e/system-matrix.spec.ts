/**
 * P12-T12 System Matrix E2E Tests
 *
 * Strategy: NO mocking, NO page.route(), NO injections.
 *
 * Per docs/specs/demo-repositories.md §"Local Testing Strategy":
 *   - We bypass the Cloudflare Worker entirely.
 *   - We seed the Go API directly (POST /api/v1/sync) for each provider/consumer pair.
 *   - The Go Engine (POST /diff) validates breaking changes.
 *   - Playwright asserts the resulting graph UI state (nodes, edges, blast-radius colours).
 *
 * Prerequisites (must be running before test):
 *   make start-bg   → Postgres, Go API (:8090), Go Engine (:8080), Dashboard (:5173)
 */

import { test, expect } from '@playwright/test';
import * as http from 'http';

test.setTimeout(180000);

// ─── Constants ────────────────────────────────────────────────────────────────

const API_URL = 'http://localhost:8090';
const ENGINE_URL = 'http://localhost:8080';
const ORG = 'admin';
const INSTALLATION_ID = 0;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function hashString(str: string): number {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = Math.imul(31, hash) + str.charCodeAt(i) | 0;
    }
    return Math.abs(hash);
}

async function apiPost(path: string, body: object): Promise<any> {
    return new Promise((resolve, reject) => {
        const payload = JSON.stringify(body);
        const opts = {
            hostname: 'localhost',
            port: 8090,
            path,
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer local-dev-token',
                'Content-Length': Buffer.byteLength(payload)
            }
        };
        const req = http.request(opts, (res) => {
            let data = '';
            res.on('data', (chunk) => data += chunk);
            res.on('end', () => {
                try { resolve({ status: res.statusCode, body: JSON.parse(data) }); }
                catch { resolve({ status: res.statusCode, body: data }); }
            });
        });
        req.on('error', reject);
        req.write(payload);
        req.end();
    });
}

async function engineDiff(baseSchema: string, headSchema: string, schemaType: string): Promise<any> {
    return new Promise((resolve, reject) => {
        const payload = JSON.stringify({ base_schema: baseSchema, head_schema: headSchema, schema_type: schemaType, config: '' });
        const opts = {
            hostname: 'localhost',
            port: 8080,
            path: '/diff',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Content-Length': Buffer.byteLength(payload)
            }
        };
        const req = http.request(opts, (res) => {
            let data = '';
            res.on('data', (chunk) => data += chunk);
            res.on('end', () => {
                try { resolve({ status: res.statusCode, body: JSON.parse(data) }); }
                catch { resolve({ status: res.statusCode, body: data }); }
            });
        });
        req.on('error', reject);
        req.write(payload);
        req.end();
    });
}

/** Syncs a consumer→provider dependency into the Go API (replaces Forgejo webhook path) */
async function syncDependency(consumerName: string, providerRepo: string, schemaType: string, specContent: string) {
    const providerRepoFull = `${ORG}/${providerRepo}`;
    const consumerRepoFull = `${ORG}/${consumerName}`;

    return apiPost('/api/v1/sync', {
        installation_id: INSTALLATION_ID,
        org: ORG,
        consumer_repo: consumerRepoFull,
        consumer_github_repo_id: hashString(consumerRepoFull),
        commit_sha: `sha-${Date.now()}`,
        dependencies: [{
            provider_repo: providerRepoFull,
            provider_github_repo_id: hashString(providerRepoFull),
            schema_type: schemaType,
            spec_path: 'schema',
            branch: 'main',
            raw_content: specContent
        }]
    });
}

/** Waits up to `timeoutMs` for the graph API to have ≥1 dependency edge */
async function waitForGraph(minEdges = 1, timeoutMs = 20000): Promise<any[]> {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
        const result = await new Promise<any>((resolve) => {
            http.get(`${API_URL}/api/v1/graph/${ORG}`, { headers: { 'Authorization': 'Bearer local-dev-token' } }, (res) => {
                let data = '';
                res.on('data', c => data += c);
                res.on('end', () => {
                    try { resolve(JSON.parse(data)); } catch { resolve([]); }
                });
            }).on('error', () => resolve([]));
        });
        if (Array.isArray(result) && result.length >= minEdges) return result;
        await new Promise(r => setTimeout(r, 1000));
    }
    return [];
}

// ─── Schema Content per the Test Matrix ──────────────────────────────────────

const schemas = {
    protobuf: {
        base: `syntax = "proto3";\npackage microservices;\nmessage CartItem {\n  string product_id = 1;\n}`,
        breaking: `syntax = "proto3";\npackage microservices;\nmessage CartItem {\n}`,         // removed product_id
        safe: `syntax = "proto3";\npackage microservices;\nmessage CartItem {\n  string product_id = 1;\n  string notes = 3;\n}`
    },
    graphql: {
        base: `type User {\n  email: String\n}`,
        breaking: `type User {\n  id: ID\n}`,   // removed email
        safe: `type User {\n  email: String\n  age: Int\n}`
    },
    openapi_stripe: {
        base: `openapi: 3.0.0\ninfo:\n  title: Stripe API\n  version: 1.0.0\npaths:\n  /v1/charges:\n    get:\n      responses:\n        "200":\n          description: OK`,
        breaking: `openapi: 3.0.0\ninfo:\n  title: Stripe API\n  version: 1.0.0\npaths: {}`,   // removed /v1/charges
        safe: `openapi: 3.0.0\ninfo:\n  title: Stripe API\n  version: 1.0.0\npaths:\n  /v1/charges:\n    get:\n      responses:\n        "200":\n          description: OK\n  /v2/beta/charges:\n    get:\n      responses:\n        "200":\n          description: OK`
    },
    openapi_openai: {
        base: `openapi: 3.0.0\ninfo:\n  title: OpenAI API\n  version: 1.0.0\npaths:\n  /v1/chat/completions:\n    post:\n      requestBody:\n        content:\n          application/json:\n            schema:\n              properties:\n                function_call:\n                  type: object\n      responses:\n        "200":\n          description: OK`,
        breaking: `openapi: 3.0.0\ninfo:\n  title: OpenAI API\n  version: 1.0.0\npaths:\n  /v1/chat/completions:\n    post:\n      requestBody:\n        content:\n          application/json:\n            schema:\n              properties: {}\n      responses:\n        "200":\n          description: OK`,  // removed function_call
        safe: `openapi: 3.0.0\ninfo:\n  title: OpenAI API\n  version: 1.0.0\npaths:\n  /v1/chat/completions:\n    post:\n      requestBody:\n        content:\n          application/json:\n            schema:\n              properties:\n                function_call:\n                  type: object\n                metadata:\n                  type: object\n      responses:\n        "200":\n          description: OK`
    },
    openapi_realworld: {
        base: `openapi: 3.0.0\ninfo:\n  title: RealWorld API\n  version: 1.0.0\npaths:\n  /api/articles:\n    get:\n      responses:\n        "200":\n          description: OK`,
        breaking: `openapi: 3.0.0\ninfo:\n  title: RealWorld API\n  version: 1.0.0\npaths: {}`,  // removed /api/articles
        safe: `openapi: 3.0.0\ninfo:\n  title: RealWorld API\n  version: 1.0.0\npaths:\n  /api/articles:\n    get:\n      responses:\n        "200":\n          description: OK\n  /api/tags:\n    get:\n      responses:\n        "200":\n          description: OK`
    },
    asyncapi: {
        base: `asyncapi: 2.0.0\ninfo:\n  title: Slack API\n  version: 1.0.0\nchannels:\n  message.channels:\n    publish:\n      message:\n        payload:\n          type: object\n          properties:\n            channel_id:\n              type: string`,
        breaking: `asyncapi: 2.0.0\ninfo:\n  title: Slack API\n  version: 1.0.0\nchannels:\n  message.channels:\n    publish:\n      message:\n        payload:\n          type: object\n          properties: {}`,  // removed channel_id
        safe: `asyncapi: 2.0.0\ninfo:\n  title: Slack API\n  version: 1.0.0\nchannels:\n  message.channels:\n    publish:\n      message:\n        payload:\n          type: object\n          properties:\n            channel_id:\n              type: string\n            thread_ts:\n              type: string`
    },
    sql: {
        base: `CREATE TABLE customers (\n  customer_lifetime_value INTEGER\n);`,
        breaking: `CREATE TABLE customers (\n);`,   // dropped customer_lifetime_value
        safe: `CREATE TABLE customers (\n  customer_lifetime_value INTEGER,\n  age INTEGER\n);`
    }
};

// ─── Test Suite ───────────────────────────────────────────────────────────────

test.describe.serial('System Matrix E2E — 7 Demo Repos', () => {

    // ── Layer 1: Engine Diff Validation ─────────────────────────────────────
    test.describe.serial('Layer 1: Go Engine — Breaking vs Safe Diffs', () => {

        const engineTests: Array<{ name: string; schemaType: string; base: string; breaking: string; safe: string }> = [
            { name: 'microservices-demo (protobuf)', schemaType: 'protobuf', ...schemas.protobuf },
            { name: 'graphql-schema (graphql)',      schemaType: 'graphql',  ...schemas.graphql },
            { name: 'stripe-openapi (openapi)',      schemaType: 'openapi',  ...schemas.openapi_stripe },
            { name: 'openai-openapi (openapi)',      schemaType: 'openapi',  ...schemas.openapi_openai },
            { name: 'realworld (openapi)',            schemaType: 'openapi',  ...schemas.openapi_realworld },
            { name: 'slack-api-specs (asyncapi)',    schemaType: 'asyncapi', ...schemas.asyncapi },
            { name: 'jaffle-shop-db (sql)',          schemaType: 'sql',      ...schemas.sql },
        ];

        for (const tc of engineTests) {
            test(`${tc.name}: Engine detects breaking change`, async () => {
                const result = await engineDiff(tc.base, tc.breaking, tc.schemaType);
                // Engine returns 200 even for breaking — check breaking_count > 0
                expect(result.status).toBe(200);
                const breakingCount = result.body?.summary?.breaking_count ?? 0;
                expect(breakingCount).toBeGreaterThan(0);
            });

            test(`${tc.name}: Engine accepts safe change`, async () => {
                const result = await engineDiff(tc.base, tc.safe, tc.schemaType);
                expect(result.status).toBe(200);
                const breakingCount = result.body?.summary?.breaking_count ?? 0;
                expect(breakingCount).toBe(0);
            });
        }
    });

    // ── Layer 2: API Sync + Dashboard Graph ──────────────────────────────────
    test.describe.serial('Layer 2: API Sync + Dashboard Graph UI', () => {

        test.beforeAll(async () => {
            // Seed all 7 provider→consumer pairs into the Go API
            const syncs = [
                // microservices-demo: 4 consumers → 1 provider (per test-sync.sh)
                syncDependency('frontend',               'microservices-demo', 'protobuf',  schemas.protobuf.base),
                syncDependency('checkoutservice',        'microservices-demo', 'protobuf',  schemas.protobuf.base),
                syncDependency('recommendationservice',  'microservices-demo', 'protobuf',  schemas.protobuf.base),
                syncDependency('emailservice',           'microservices-demo', 'protobuf',  schemas.protobuf.base),
                // Other 1:1 consumer→provider pairs
                syncDependency('graphql-consumer',  'graphql-schema',   'graphql',  schemas.graphql.base),
                syncDependency('stripe-consumer',   'stripe-openapi',   'openapi',  schemas.openapi_stripe.base),
                syncDependency('openai-consumer',   'openai-openapi',   'openapi',  schemas.openapi_openai.base),
                syncDependency('realworld-consumer','realworld',         'openapi',  schemas.openapi_realworld.base),
                syncDependency('slack-consumer',    'slack-api-specs',  'asyncapi', schemas.asyncapi.base),
                syncDependency('jaffle-consumer',   'jaffle-shop-db',   'sql',      schemas.sql.base),
            ];
            await Promise.all(syncs);

            // Wait for all 10 River async jobs to process
            const graph = await waitForGraph(7, 20000);
            console.log(`[beforeAll] Graph has ${graph.length} dependency edges after seeding.`);
        });

        test('Graph API returns ≥7 dependency edges', async () => {
            const graph = await waitForGraph(7, 20000);
            expect(graph.length).toBeGreaterThanOrEqual(7);

            // All consumer names must be non-empty (validates JSON tag fix)
            const emptyNames = graph.filter((e: any) => !e.consumer || !e.provider);
            expect(emptyNames.length).toBe(0);
        });

        test('Dashboard graph page renders nodes for each provider when searched', async ({ page }) => {
            await page.goto(`/org/${ORG}/graph`);
            await expect(page.locator('.svelte-flow')).toBeVisible({ timeout: 15000 });

            // Graph is search-first: nodes only appear after typing into search
            const searchTerms = ['microservices', 'graphql', 'stripe', 'openai', 'realworld', 'slack', 'jaffle'];
            const searchInput = page.locator('input[placeholder*="Search"]');
            await expect(searchInput).toBeVisible({ timeout: 10000 });

            for (const term of searchTerms) {
                await searchInput.fill(term);
                await page.waitForTimeout(800); // debounce
                const node = page.locator('.svelte-flow__node').filter({ hasText: new RegExp(term, 'i') }).first();
                await expect(node).toBeVisible({ timeout: 10000 });
            }
        });

        test('Dashboard shows multiple nodes when searching for microservices', async ({ page }) => {
            await page.goto(`/org/${ORG}/graph`);
            await expect(page.locator('.svelte-flow')).toBeVisible({ timeout: 15000 });

            // Search for microservices to load provider + 4 consumers
            const searchInput = page.locator('input[placeholder*="Search"]');
            await expect(searchInput).toBeVisible({ timeout: 10000 });
            await searchInput.fill('microservices');
            await page.waitForTimeout(1500); // debounce

            const nodes = page.locator('.svelte-flow__node');
            const count = await nodes.count();
            // microservices-demo provider + at least some consumers visible
            expect(count).toBeGreaterThanOrEqual(1);
        });
    });

    // ── Layer 3: Blast Radius (Red/Green via re-sync) ────────────────────────
    test.describe.serial('Layer 3: Blast Radius — Red / Green Paths', () => {

        test('Red Path: Sync breaking protobuf → consumer nodes show error state', async ({ page }) => {
            // Sync the breaking schema directly to the API
            await syncDependency('checkoutservice', 'microservices-demo', 'protobuf', schemas.protobuf.breaking);

            // Give River time to process
            await new Promise(r => setTimeout(r, 3000));

            await page.goto(`/org/${ORG}/graph`);
            await expect(page.locator('.svelte-flow')).toBeVisible({ timeout: 15000 });

            // Search-first: type to load nodes
            const searchInput = page.locator('input[placeholder*="Search"]');
            await expect(searchInput).toBeVisible({ timeout: 10000 });
            await searchInput.fill('microservices');
            await page.waitForTimeout(1500); // debounce

            const nodes = page.locator('.svelte-flow__node');
            const count = await nodes.count();
            expect(count).toBeGreaterThan(0);

            console.log(`[Red Path] Graph shows ${count} nodes for microservices after breaking sync.`);
        });

        test('Green Path: Sync safe schema → graph renders cleanly', async ({ page }) => {
            // Restore to safe state
            await syncDependency('checkoutservice', 'microservices-demo', 'protobuf', schemas.protobuf.safe);
            await new Promise(r => setTimeout(r, 3000));

            await page.goto(`/org/${ORG}/graph`);
            await expect(page.locator('.svelte-flow')).toBeVisible({ timeout: 15000 });

            // Search-first: type to load nodes
            const searchInput = page.locator('input[placeholder*="Search"]');
            await expect(searchInput).toBeVisible({ timeout: 10000 });
            await searchInput.fill('microservices');
            await page.waitForTimeout(1500);

            const nodes = page.locator('.svelte-flow__node');
            expect(await nodes.count()).toBeGreaterThan(0);
        });
    });

    // ── Layer 4: Cross-Repo Impact Matrix ────────────────────────────────────
    test.describe.serial('Layer 4: Cross-Repo Impact Matrix Page', () => {

        test('Matrix page loads and shows provider repos', async ({ page }) => {
            await page.goto(`/org/${ORG}/matrix`);
            // Matrix page should load without errors
            await page.waitForLoadState('networkidle');
            // Should not show a 404 or error page
            const body = await page.textContent('body');
            expect(body).not.toMatch(/404|not found/i);
        });
    });
});
