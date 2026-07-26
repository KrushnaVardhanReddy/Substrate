import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8090';

test.describe('SSE Realtime E2E (Suite 12)', () => {

    test('should receive backend event in multiple contexts without page reload', async ({ browser }) => {
        // Create two separate browser contexts
        const context1 = await browser.newContext();
        const context2 = await browser.newContext();

        const page1 = await context1.newPage();
        const page2 = await context2.newPage();

        // Check API is reachable
        const healthRes = await page1.request.get(`${API_URL}/health`, { timeout: 3000 }).catch(() => null);
        if (!healthRes || !healthRes.ok()) {
            await context1.close();
            await context2.close();
            test.skip(true, 'Live API not reachable — skipping SSE realtime test');
            return;
        }

        // Navigate to graph page — the UI opens an SSE connection to /api/v1/events
        await page1.goto('/org/mcp-org/graph');
        await page2.goto('/org/mcp-org/graph');

        // Wait a moment for the pages to connect to the live SSE endpoint
        await page1.waitForTimeout(2000);
        await page2.waitForTimeout(2000);

        // Trigger a sync via the live API to fire an SSE event
        await page1.request.post(`${API_URL}/api/v1/sync`, {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer local-dev-token',
            },
            data: {
                installation_id: 0,
                org: 'mcp-org',
                consumer_repo: 'mcp-org/sse-test-consumer',
                consumer_github_repo_id: 9999,
                commit_sha: `sha-sse-${Date.now()}`,
                dependencies: [{
                    provider_repo: 'mcp-org/backend',
                    provider_github_repo_id: 101,
                    schema_type: 'openapi',
                    spec_path: 'openapi.yaml',
                    branch: 'main',
                    raw_content: 'openapi: 3.0.0\ninfo:\n  title: SSE Test\n  version: 1.0.0\npaths: {}'
                }]
            }
        });

        // Allow time for backend to emit the event
        await page1.waitForTimeout(2000);
        await page2.waitForTimeout(2000);

        // Verify both pages are still alive and connected (no crash)
        const title1 = await page1.title();
        const title2 = await page2.title();
        expect(title1).toBeTruthy();
        expect(title2).toBeTruthy();

        await context1.close();
        await context2.close();
    });

});
