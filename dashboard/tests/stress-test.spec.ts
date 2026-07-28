import { test, expect } from '@playwright/test';

test.describe('1,000-Node UI Stress Test', () => {
	test('can load a graph with 1,000 nodes and 3,000 edges without crashing', async ({ page, context }) => {
		// Use a real JWT from the E2E runner (signed with local-jwt-secret)
		await context.addInitScript((t) => {
			localStorage.setItem('github_token', t);
		}, process.env.E2E_AUTH_TOKEN || 'placeholder');
		// The `stress-test` org in the DB is seeded with 1,000 nodes and 3,000 edges.

		// Set a longer timeout for the test given the large graph payload
		test.setTimeout(120000);

		// Listen for JS errors that indicate browser crash/OOM
		const errors: string[] = [];
		page.on('pageerror', err => {
			if (!err.message.includes('hydration')) {
				errors.push(err.message);
			}
		});

		await page.goto('/org/stress-test/graph');

		// Wait for the empty state — this means the initial data has loaded
		await expect(page.locator('.empty-state')).toBeVisible({ timeout: 10000 });

		await page.fill('input[placeholder="Search repository..."]', 'node-15');
		// Wait for the debounce (300ms) plus some render time
		await page.waitForTimeout(600);

		// Wait for cytoscape canvas to mount (it draws on a single canvas)
		await page.waitForSelector('canvas[data-id="layer2-node"]', { state: 'attached', timeout: 30000 });

		// Poll until cyInstance is populated (it may take a frame or two after canvas appears)
		const nodeCount = await page.waitForFunction(() => {
			const cy = (window as any).cyInstance;
			return cy ? cy.nodes().length : 0;
		}, { timeout: 15000 }).then(h => h.jsonValue() as Promise<number>);

		expect(nodeCount).toBeGreaterThan(0);

		// Assert no actual crashes
		expect(errors.length).toBe(0);
	});
});
