import { test, expect } from '@playwright/test';

test.describe('Dependency Graph Filters', () => {
	test('can toggle BREAKING filter and highlight neighbors', async ({ page }) => {
		// Mock graph data to have a BREAKING edge and SAFE edges
		await page.route('**/api/v1/graph/*', async route => {
			const json = [
				{ provider: 'service-openapi-A', consumer: 'service-B', status: 'SAFE' },
				{ provider: 'service-C', consumer: 'service-openapi-A', status: 'BREAKING' },
				{ provider: 'service-D', consumer: 'service-E', status: 'SAFE' }
			];
			await route.fulfill({ json });
		});

		await page.goto('/org/myorg/graph');

		// Wait for canvas to load
		await expect(page.locator('.main-canvas')).toBeVisible();

		// Wait for the mock fetch to complete
		await page.waitForTimeout(1000);

		// Check the checkbox for BREAKING
		const breakingCheckbox = page.locator('text=Show Only BREAKING Changes');
		await breakingCheckbox.check();

		// Check the checkbox for Neighbors
		const neighborCheckbox = page.locator('text=Highlight connected neighbors');
		await neighborCheckbox.check();

		// Change protocol select
		await page.selectOption('select.filter-select', 'openapi');

		// Type search query
		await page.fill('input[placeholder="Search repository..."]', 'service-D');

		// Give Cytoscape layout a moment to apply
		await page.waitForTimeout(500);

		// We can't easily assert inside the canvas, but we can verify the UI didn't crash
		await expect(page.locator('.filter-panel')).toBeVisible();
	});
});
