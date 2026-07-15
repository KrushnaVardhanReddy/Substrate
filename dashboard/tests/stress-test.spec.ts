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
			
			// Guarantee exactly 1000 unique nodes by creating a straight line
			for (let i = 0; i < 999; i++) {
				json.push({
					provider: `node-${i}`,
					consumer: `node-${i + 1}`,
					status: 'SAFE'
				});
			}
			
			// Add 2000 more random edges, strictly provider < consumer to keep it acyclic
			for (let i = 0; i < 2000; i++) {
				const providerIdx = Math.floor(Math.random() * 998);
				// Consumer is strictly greater than provider
				const consumerIdx = providerIdx + 1 + Math.floor(Math.random() * (999 - providerIdx));
				json.push({
					provider: `node-${providerIdx}`,
					consumer: `node-${consumerIdx}`,
					status: Math.random() > 0.95 ? 'BREAKING' : 'SAFE'
				});
			}
			await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(json) });
		});

		const startTime = Date.now();
		
		const responsePromise = page.waitForResponse(response => response.url().includes('/api/v1/graph/'));
		await page.goto('/org/stress-test/graph');

		// Wait for network response
		await responsePromise;

		// Wait for dagre layout to mount all 1,000 nodes
		await expect(page.locator('.svelte-flow__node')).toHaveCount(1000, { timeout: 30000 });
		
		// Verify that the UI is still responsive and didn't crash
		const cyContainer = page.locator('.svelte-flow').first();
		await expect(cyContainer).toBeVisible({ timeout: 15000 });

		const endTime = Date.now();
		console.log(`Stress test completed in ${endTime - startTime}ms`);
	});
});
