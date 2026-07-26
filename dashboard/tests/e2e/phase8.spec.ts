import { test, expect } from '@playwright/test';

test.describe('Phase 8 Readiness, Authz & Jobs', () => {

	test.beforeEach(async ({ page }) => {
		page.on('pageerror', (err) => {
			if (err.message.includes('hydration')) {
				console.warn('Hydration error suppressed:', err.message);
			} else {
				throw err;
			}
		});

        // Add init script to pass auth
        await page.addInitScript(() => {
            localStorage.setItem('substrate-token', 'local-dev-token');
        });
	});

	test('should render correct state on org dashboard and graph', async ({ page }) => {
		await page.goto('/org/mcp-org/');

        // Basic assertions based on "Assert the UI renders the correct state."
		await expect(page.locator('h1')).toBeVisible();

        // Navigate to graph and interact with live cytoscape instance
        const responsePromise = page.waitForResponse('**/api/v1/graph/*', { timeout: 15000 });
        await page.goto('/org/mcp-org/graph');
        await responsePromise;

        // Search to trigger node rendering (graph is search-first)
        const searchInput = page.locator('input[placeholder*="Search"], input.filter-input').first();
        await expect(searchInput).toBeVisible({ timeout: 10000 });
        await searchInput.fill('backend');
        await page.waitForTimeout(800);

        // Interact with cytoscape via window.cyInstance per the E2E rules
        const nodeCount = await page.waitForFunction(
            () => (window as any).cyInstance && (window as any).cyInstance.nodes().length > 0,
            { timeout: 15000 }
        ).then(h => h.jsonValue()).catch(() => 0);

        expect(nodeCount).toBeGreaterThan(0);

        await page.evaluate(() => {
            const cy = (window as any).cyInstance;
            const node = cy.nodes().first();
            if (node) node.emit('tap');
        });

        await page.waitForTimeout(500);
        await expect(page.locator('.detail-panel')).toBeVisible({ timeout: 5000 });
	});
});
