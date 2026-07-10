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

	test('should render graph nodes and display node details on click', async ({ page }) => {
		// Navigate directly to the graph page (SSR is disabled in tests)
		await page.goto('/org/testorg/graph');

		// Wait for the graph page to load
		await expect(page.locator('h1.page-title')).toContainText('Dependency Graph');

		// Ensure the node card for "core/auth" is visible
		const authNode = page.locator('.node-card', { hasText: 'core/auth' }).first();
		await expect(authNode).toBeVisible();

		// Click the node card
		await authNode.click({ force: true });

		// Assert that the detail panel becomes visible
		const detailPanel = page.locator('aside.detail-panel');
		await expect(detailPanel).toBeVisible({ timeout: 10000 });

		// Verify the detail panel displays "core/auth"
		await expect(detailPanel.locator('.detail-title')).toContainText('core/auth');
	});
});
