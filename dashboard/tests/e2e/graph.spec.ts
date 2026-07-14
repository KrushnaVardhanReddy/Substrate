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
					{ provider: 'core/auth', consumer: 'frontend/dashboard', status: 'SAFE' }
				])
			});
		});
	});

	test('should render graph container and filter controls', async ({ page }) => {
		// Navigate directly to the graph page (SSR is disabled in tests)
		await page.goto('/org/testorg/graph');

		// Wait for the graph page to load
		await expect(page.locator('h1.page-title')).toContainText('Dependency Graph');

		// Check for Graph Controls (filter panel)
		await expect(page.locator('.filter-panel')).toBeVisible();

		// Check that the checkbox, select, and input are rendered
		await expect(page.locator('input[type="checkbox"]').first()).toBeVisible();
		await expect(page.locator('select.filter-select')).toBeVisible();
		await expect(page.locator('input.filter-input')).toBeVisible();

		// Check for the Cytoscape container (canvas is inside)
		const cyContainer = page.locator('.main-canvas div').first();
		await expect(cyContainer).toBeVisible();
	});

	test('should render graph nodes and intercept network responses', async ({ page }) => {
		// Navigate directly to the graph page (SSR is disabled in tests)
		const responsePromise = page.waitForResponse('**/api/v1/graph/*');
		await page.goto('/org/testorg/graph');

		// Wait for the graph page to load
		await expect(page.locator('h1.page-title')).toContainText('Dependency Graph');

		// Wait for network response
		await responsePromise;

		// Ensure the graph container is visible
		const cyContainer = page.locator('.main-canvas div').first();
		await expect(cyContainer).toBeVisible();
	});
});
