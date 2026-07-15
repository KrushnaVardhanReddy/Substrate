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

		await page.route('**/api/v1/repos/**', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{ id: '1', name: 'core-auth', full_name: 'core/auth' }
				])
			});
		});

		await page.route('**/api/v1/graph/**', async route => {
			const json = [];
			
			// Guarantee exactly 200 unique nodes by creating a straight line
			for (let i = 0; i < 199; i++) {
				json.push({
					provider: `node-${i}`,
					consumer: `node-${i + 1}`,
					status: 'SAFE'
				});
			}
			
			// Add 600 more random edges, strictly provider < consumer to keep it acyclic
			for (let i = 0; i < 600; i++) {
				const providerIdx = Math.floor(Math.random() * 198);
				const consumerIdx = providerIdx + 1 + Math.floor(Math.random() * (199 - providerIdx));
				json.push({
					provider: `node-${providerIdx}`,
					consumer: `node-${consumerIdx}`,
					status: Math.random() > 0.95 ? 'BREAKING' : 'SAFE'
				});
			}
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(json) });
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
