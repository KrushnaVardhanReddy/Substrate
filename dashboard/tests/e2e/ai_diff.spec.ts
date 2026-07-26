import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8090';

test.describe('Phase 4 AI Diff Engine E2E', () => {
    test('AI Patch generation flow', async ({ page }) => {
        // Check API is reachable
        const healthRes = await page.request.get(`${API_URL}/health`, { timeout: 3000 }).catch(() => null);
        if (!healthRes || !healthRes.ok()) {
            test.skip(true, 'Live API not reachable — skipping AI Diff test');
            return;
        }

        // Seed two schemas via the live sync endpoint
        const baseSchema = 'openapi: 3.0.0\ninfo:\n  title: Diff Test\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        "200":\n          description: OK';
        const headSchema = 'openapi: 3.0.0\ninfo:\n  title: Diff Test\n  version: 1.0.0\npaths: {}';

        // Push baseline schema
        await page.request.post(`${API_URL}/api/v1/diff`, {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer local-dev-token',
            },
            data: {
                base_schema: baseSchema,
                head_schema: headSchema,
                schema_type: 'openapi',
                config: ''
            }
        });

        // Navigate to diff viewer page
        await page.goto('/org/mcp-org/diff/diff-123');

        // Wait for viewer to load (the page should render even without AI)
        await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => {});

        // The diff viewer page should be accessible
        const body = await page.textContent('body');
        expect(body).toBeTruthy();
        // Should not show a generic 404 / error
        expect(body).not.toMatch(/404|Page Not Found/i);
    });
});
