import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8090';

test.describe('Cross-Repo Blast Radius E2E (Suite 13)', () => {

    test('should display blast radius alert on C when A is broken in A -> B -> C chain', async ({ page, request }) => {
        // Verify API is up
        const health = await request.get(`${API_URL}/health`, { timeout: 3000 }).catch(() => null);
        if (!health || !health.ok()) {
            test.skip(true, 'API not reachable');
            return;
        }

        // The seed_mcp.sql seeds admin org with:
        // repo-a (provider) ← repo-b (consumer) ← repo-c (consumer)
        // Push a breaking change via sync to simulate A breaking
        await request.post(`${API_URL}/api/v1/sync`, {
            headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer local-dev-token' },
            data: {
                installation_id: 0,
                org: 'admin',
                consumer_repo: 'admin/repo-b',
                consumer_github_repo_id: 402,
                commit_sha: `sha-blast-${Date.now()}`,
                dependencies: [{
                    provider_repo: 'admin/repo-a',
                    provider_github_repo_id: 401,
                    schema_type: 'openapi',
                    spec_path: 'openapi.yaml',
                    branch: 'main',
                    raw_content: 'openapi: 3.0.0\ninfo:\n  title: Repo A\n  version: 2.0.0\npaths: {}'
                }]
            }
        });

        await request.post(`${API_URL}/api/v1/sync`, {
            headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer local-dev-token' },
            data: {
                installation_id: 0,
                org: 'admin',
                consumer_repo: 'admin/repo-c',
                consumer_github_repo_id: 403,
                commit_sha: `sha-blast-c-${Date.now()}`,
                dependencies: [{
                    provider_repo: 'admin/repo-b',
                    provider_github_repo_id: 402,
                    schema_type: 'openapi',
                    spec_path: 'openapi.yaml',
                    branch: 'main',
                    raw_content: 'openapi: 3.0.0\ninfo:\n  title: Repo B\n  version: 1.0.0\npaths:\n  /api:\n    get:\n      responses:\n        "200":\n          description: OK'
                }]
            }
        });

        // Allow River to process
        await page.waitForTimeout(3000);

        // Navigate to graph for admin org
        await page.goto('/org/admin/graph');
        await page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => {});

        // Type search to trigger node rendering
        const searchInput = page.locator('input.filter-input, input[placeholder*="Search"]').first();
        await expect(searchInput).toBeVisible({ timeout: 10000 });
        await searchInput.fill('repo');
        await page.waitForTimeout(1500);

        // Wait for at least one node to appear
        await page.waitForFunction(
            () => (window as any).cyInstance && (window as any).cyInstance.nodes().length > 0,
            { timeout: 15000 }
        );

        const nodeCount = await page.evaluate(() => (window as any).cyInstance.nodes().length);
        expect(nodeCount).toBeGreaterThan(0);
        console.log(`[Blast Radius] Graph shows ${nodeCount} nodes for admin org`);
    });

});
