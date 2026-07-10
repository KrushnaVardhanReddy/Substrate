import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
	test.beforeEach(async ({ page }) => {
		await page.route('**/api/v1/repos/*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{ id: '1', name: 'core-auth', full_name: 'core/auth' },
					{ id: '2', name: 'frontend-dashboard', full_name: 'frontend/dashboard' }
				])
			});
		});

		await page.route('**/api/v1/graph/*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([])
			});
		});

		await page.route('**/api/v1/matrix/*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					providers: [],
					consumers: [],
					grid: []
				})
			});
		});
	});

	test('should navigate between different dashboard views', async ({ page }) => {
		// Navigate directly to the org overview page (SSR is disabled in tests)
		await page.goto('/org/testorg');

		// Check if we are on the dashboard
		await expect(page.locator('h1.page-title')).toContainText('Repositories');
		
		// Let the UI settle
		await page.waitForTimeout(500);
		
		// Let the UI settle
		await page.waitForTimeout(500);

		// Click the "Dependency Graph" link in the sidebar
		await page.getByText('Dependency Graph', { exact: true }).first().click();
		await page.waitForURL('**/org/*/graph');
		await expect(page).toHaveURL(/.*\/org\/.*?\/graph/);
		
		await page.waitForTimeout(500);

		// Click the "Compatibility Matrix" link in the sidebar
		await page.getByText('Compatibility Matrix', { exact: true }).first().click();
		await page.waitForURL('**/org/*/matrix');
		await expect(page).toHaveURL(/.*\/org\/.*?\/matrix/);
	});
});
