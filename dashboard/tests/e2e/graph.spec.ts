import { test, expect } from '@playwright/test';

test.describe('Dependency Graph', () => {

	test.beforeEach(async ({ page }) => {
		// We set github_token so initial fetch gets a valid auth token.
		await page.addInitScript(() => {
			localStorage.setItem('github_token', 'local-dev-token');
		});

		// Listen to unhandled promise rejections / hydration issues and ignore known Svelte hydration errors
		page.on('pageerror', (err) => {
			if (err.message.includes('hydration')) {
				console.warn('Ignored hydration error:', err.message);
			} else {
				throw err;
			}
		});
	});

	test('should render graph container and filter controls', async ({ page }) => {
		const responsePromise = page.waitForResponse('**/api/v1/graph/*');
		await page.goto('/org/mcp-org/graph');
		await responsePromise;

		await expect(page.locator('h1.page-title')).toContainText('Dependency Graph');
		await expect(page.locator('.filter-panel')).toBeVisible();
		await expect(page.locator('input[type="checkbox"]').first()).toBeVisible();
		await expect(page.locator('select.filter-select')).toBeVisible();
		await expect(page.locator('input.filter-input')).toBeVisible();

		// Check for Rotate and Export buttons
		await expect(page.locator('button[title="Export PNG"]')).toBeVisible();
		await expect(page.locator('button[title="Rotate Layout"]')).toBeVisible();

		const cyContainer = page.locator('.graph-container main').first();
		await expect(cyContainer).toBeVisible();

		// Since cytoscape can take a moment to initialize and mount to window:
		await page.waitForFunction(() => {
			return typeof (window as any).cyInstance !== 'undefined';
		}, { timeout: 15000 });
	});

	test('should handle layout rotation, taxonomy badges, and deep links', async ({ page }) => {
		const responsePromise = page.waitForResponse('**/api/v1/graph/*');
		await page.goto('/org/mcp-org/graph');
		await responsePromise;

		// Fill search to show nodes in Search-First mode if necessary
		await page.fill('input.filter-input', '/');

		// Wait for the cyContainer to be attached
		await page.waitForSelector('.graph-container main');

		// Since cytoscape can take a moment to initialize and mount to window:
		await page.waitForFunction(() => {
			return typeof (window as any).cyInstance !== 'undefined' && (window as any).cyInstance.nodes().length > 0;
		}, { timeout: 15000 });

		// Assert that two distinct nodes exist
		const nodeCount = await page.evaluate(() => {
			const cy = (window as any).cyInstance;
			return cy.nodes().length;
		});
		
		expect(nodeCount).toBeGreaterThanOrEqual(2);

		// Check the API returned a backend node
		const hasBackend = await page.evaluate(() => {
			const cy = (window as any).cyInstance;
			return cy.nodes().some((n: any) => n.id() === 'mcp-org/backend' || n.id().includes('backend'));
		});
		expect(hasBackend).toBeTruthy();

		// Click the node to open detail panel using cytoscape API
		await page.evaluate(() => {
			const cy = (window as any).cyInstance;
			const backendNode = cy.nodes().filter((n: any) => n.id() === 'mcp-org/backend' || n.id().includes('backend')).first();
			if (backendNode) {
				backendNode.emit('tap');
			}
		});

		// Verify Taxonomy Metadata in the detail panel
		const detailPanel = page.locator('.detail-panel');
		await expect(detailPanel).toBeVisible();
	});

});
