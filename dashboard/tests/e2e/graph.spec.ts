import { test, expect } from '@playwright/test';

test.describe('Dependency Graph', () => {


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
		await page.fill('input.filter-input', '/');
		
		await page.waitForSelector('.svelte-flow', { state: 'attached', timeout: 15000 });
		// Wait for the debounced search and graph render
		await page.waitForTimeout(1000);
		await page.waitForSelector('.service-node-card', { state: 'attached', timeout: 15000 });
		
		// 1. Check tinted icons logic based on taxonomy metadata
		await expect(page.locator('.icon-wrapper.database').first()).toBeVisible();
		await expect(page.locator('.icon-wrapper.frontend').first()).toBeVisible();

		// 2. Click the Rotate Layout button
		const rotateBtn = page.locator('button[title="Rotate Layout"]');
		await rotateBtn.click(); // Should change state to LR
		await rotateBtn.click(); // Should change state to TB

		// 3. Click the node to open detail panel
		await page.locator('.service-node-card.database').first().click({ force: true });

		// 4. Verify Taxonomy Metadata in the detail panel
		const detailPanel = page.locator('.detail-panel');
		await expect(detailPanel).toBeVisible();
		await expect(detailPanel.getByRole('heading', { name: 'Taxonomy' })).toBeVisible();
		await expect(detailPanel.getByText('Platform')).toBeVisible();

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

	test('should highlight blast radius on node click', async ({ page }) => {
		await page.goto('/org/testorg/graph');
		await page.waitForResponse('**/api/v1/graph/*');
		
		// Search-First mode: fill search to show nodes
		await page.fill('input.filter-input', '/');
		await page.waitForSelector('.service-node-card', { state: 'attached', timeout: 15000 });

		// Click a database node
		await page.locator('.service-node-card.database').first().click({ force: true });
		
		// Wait for reactivity
		await page.waitForTimeout(500);

		// Assertions
		await expect(page.locator('.detail-panel')).toBeVisible();
		await expect(page.locator('.service-node-card.origin').first()).toBeVisible();
		// We don't check for .faded because the mock only has 2 connected nodes.
	});

	test('should show edge tooltip on hover', async ({ page }) => {
		await page.goto('/org/testorg/graph');
		await page.waitForResponse('**/api/v1/graph/*');
		
		// Search-First mode: fill search to show nodes
		await page.fill('input.filter-input', '/');
		await page.waitForSelector('.svelte-flow__edge', { state: 'attached', timeout: 15000 });

		// Hover over the edge interaction path which has a wide stroke using dispatchEvent to bypass SVG bounding box issues
		await page.locator('.svelte-flow__edge-interaction').first().dispatchEvent('mouseenter');

		// Assert tooltip visibility
		await expect(page.locator('.tooltip-card')).toBeVisible({ timeout: 5000 });
	});
});
