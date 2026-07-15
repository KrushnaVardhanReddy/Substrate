import { test, expect } from '@playwright/test';

test.describe('Dependency Graph', () => {
	test.beforeEach(async ({ page }) => {
		await page.route('**/api/v1/repos/*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{ id: '1', name: 'core-auth', full_name: 'core/auth' }
				])
			});
		});

		await page.route('**/api/v1/graph/*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{ 
						provider: 'core/auth', 
						consumer: 'frontend/dashboard', 
						status: 'SAFE',
						provider_metadata: { type: 'database', team: 'Platform' },
						consumer_metadata: { type: 'frontend', team: 'Product' }
					}
				])
			});
		});
	});

	test('should render graph container and filter controls', async ({ page }) => {
		const responsePromise = page.waitForResponse('**/api/v1/graph/*');
		await page.goto('/org/testorg/graph');
		await responsePromise;

		await expect(page.locator('h1.page-title')).toContainText('Dependency Graph');
		await expect(page.locator('.filter-panel')).toBeVisible();
		await expect(page.locator('input[type="checkbox"]').first()).toBeVisible();
		await expect(page.locator('select.filter-select')).toBeVisible();
		await expect(page.locator('input.filter-input')).toBeVisible();

		// Check for Rotate and Export buttons
		await expect(page.locator('button[title="Export PNG"]')).toBeVisible();
		await expect(page.locator('button[title="Rotate Layout"]')).toBeVisible();

		const cyContainer = page.locator('.svelte-flow').first();
		await expect(cyContainer).toBeVisible();
	});

	test('should handle layout rotation, taxonomy badges, and deep links', async ({ page }) => {
		const responsePromise = page.waitForResponse('**/api/v1/graph/*');
		await page.goto('/org/testorg/graph');
		await responsePromise;

		// The graph should initially be empty due to Search-First model, but wait, the test currently triggers graph render without search in the app or the test passes search? Wait, the test mock data is rendered if search is typed, or if it bypasses search. Let's type in the search box to be safe.
		await page.fill('input.filter-input', 'core');
		
		await page.waitForSelector('.svelte-flow', { state: 'attached', timeout: 15000 });
		
		// 1. Check tinted icons logic based on taxonomy metadata
		await expect(page.locator('.icon-wrapper.database').first()).toBeVisible();
		await expect(page.locator('.icon-wrapper.frontend').first()).toBeVisible();

		// 2. Click the Rotate Layout button
		const rotateBtn = page.locator('button[title="Rotate Layout"]');
		await rotateBtn.click(); // Should change state to LR
		await rotateBtn.click(); // Should change state to TB

		// 3. Click the node to open detail panel
		await page.locator('.service-node-card.database').first().click();

		// 4. Verify Taxonomy Metadata in the detail panel
		const detailPanel = page.locator('.detail-panel');
		await expect(detailPanel).toBeVisible();
		await expect(detailPanel.locator('text=Taxonomy')).toBeVisible();
		await expect(detailPanel.locator('text=Platform')).toBeVisible();

		// 5. Verify Deep DX Links
		const logsLink = page.locator('a:has-text("View Logs")');
		const ideLink = page.locator('a:has-text("Open in IDE")');
		await expect(logsLink).toBeVisible();
		await expect(ideLink).toBeVisible();
		
		const hrefLogs = await logsLink.getAttribute('href');
		expect(hrefLogs).toContain('github.com');
		expect(hrefLogs).toContain('actions');

		const hrefIde = await ideLink.getAttribute('href');
		expect(hrefIde).toContain('vscode://');
	});

	test('should trigger PNG export', async ({ page }) => {
		const responsePromise = page.waitForResponse('**/api/v1/graph/*');
		await page.goto('/org/testorg/graph');
		await responsePromise;

		await page.fill('input.filter-input', 'core');
		await page.waitForSelector('.svelte-flow', { state: 'attached' });

		const exportBtn = page.locator('button[title="Export PNG"]');
		
		// Setup a listener for the download
		const downloadPromise = page.waitForEvent('download');
		await exportBtn.click();
		
		const download = await downloadPromise;
		expect(download.suggestedFilename()).toBe('substrate-graph.png');
	});
});
