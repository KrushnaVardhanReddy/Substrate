import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {


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
