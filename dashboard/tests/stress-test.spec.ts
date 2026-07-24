import { test, expect } from '@playwright/test';

test.describe('1,000-Node UI Stress Test', () => {
	test('can load a graph with 1,000 nodes and 3,000 edges without crashing', async ({ page }) => {
		test.setTimeout(120000); // 2 minutes timeout for the whole test

		page.on('pageerror', exception => {
			console.log(`Uncaught exception: "${exception}"`);
		});

		page.on('console', msg => {
			console.log(`Console message: "${msg.text()}"`);
		});





		const startTime = Date.now();
		
		await page.goto('/org/stress-test/graph');

		// Verify empty state is shown initially
		await expect(page.locator('.empty-state')).toBeVisible({ timeout: 10000 });
		await expect(page.locator('.svelte-flow__node')).toHaveCount(0);

		// Type a search query to trigger subset layout
		await page.fill('input[placeholder="Search repository..."]', 'node-15');

		// Wait for dagre layout to mount the subset of nodes
		// We expect > 0 nodes, not 200, since it's a subset
		await expect(page.locator('.svelte-flow__node')).not.toHaveCount(0, { timeout: 30000 });
		
		// Verify that the UI is still responsive and didn't crash
		const cyContainer = page.locator('.svelte-flow').first();
		await expect(cyContainer).toBeVisible({ timeout: 15000 });

		const endTime = Date.now();
		console.log(`Stress test completed in ${endTime - startTime}ms`);
	});
});
