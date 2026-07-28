import { test, expect } from '@playwright/test';

test.describe('Phase 11: Graph UI & Visual Studio', () => {

	test.beforeEach(async ({ page }) => {
		// Suppress Svelte dev-mode hydration errors which can break Playwright
		page.on('pageerror', (err) => {
			if (err.message.includes('hydration')) {
				console.warn('Hydration error suppressed:', err.message);
			} else {
				throw err;
			}
		});

		// Provide local-dev-token to avoid unauthorized responses
		await page.addInitScript(() => {
			window.localStorage.setItem('auth_token', 'local-dev-token');
		});
	});

	test('should render graph container and interact with nodes', async ({ page }) => {
		const responsePromise = page.waitForResponse('**/api/v1/graph/*', { timeout: 30000 });
		await page.goto('/org/mcp-org/graph');
		await responsePromise;

		// Assert page renders correctly
		await expect(page.locator('h1.page-title')).toContainText('Dependency Graph');

		// The graph is search-gated: cyInstance is null until a search query is typed
		const searchInput = page.locator('input.filter-input, input[placeholder*="Search"]').first();
		await expect(searchInput).toBeVisible({ timeout: 10000 });
		await searchInput.fill('backend');
		await page.waitForTimeout(800); // debounce

		// Wait for cytoscape instance to be ready (set on window after search)
		await page.waitForFunction(() => (window as any).cyInstance !== undefined && (window as any).cyInstance !== null, { timeout: 15000 });

		// Ensure graph is populated
		await page.waitForFunction(() => (window as any).cyInstance.nodes().length > 0, { timeout: 10000 });

		// Click the node via Cytoscape API
		await page.evaluate(() => {
			const cy = (window as any).cyInstance;
			const node = cy.nodes().first();
			if (node) {
				node.emit('tap');
			}
		});

		// Check the detail panel
		const detailPanel = page.locator('.detail-panel');
		await expect(detailPanel).toBeVisible({ timeout: 5000 });
		await expect(detailPanel.locator('.detail-title')).toBeVisible();
	});

});
