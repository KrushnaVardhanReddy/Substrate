import { test, expect } from '@playwright/test';

test.describe('Phase 2 E2E: GitHub App UI', () => {

	test('should render the correct state for repositories in the organization', async ({ page, request }) => {
		// Mock token setup if needed - in this environment, requests rely on PUBLIC_API_TOKEN etc.
		// Check if backend API is up. Since this is an E2E test, if local API is unreachable, we skip.
		try {
			const res = await request.get('http://localhost:8090/health', { timeout: 2000 });
			if (!res.ok()) {
				test.skip(true, 'Live infrastructure unreachable (API). Skipping E2E test.');
			}
		} catch (e) {
			test.skip(true, 'Live infrastructure unreachable (API). Skipping E2E test.');
		}

		const responsePromise = page.waitForResponse('**/api/v1/repos/*');

		await page.goto('/org/mcp-org');

		await responsePromise;

		// The page should be displaying "Repositories"
		await expect(page.locator('h1.page-title')).toContainText('Repositories');
		await expect(page.locator('.page-subtitle')).toContainText('Manage and view schemas for all registered repositories.');

		// The table should be visible and contain the repositories we seeded
		const table = page.locator('table');
		await expect(table).toBeVisible();

		// Ensure 'backend' and 'frontend' are rendered in the table rows
		await expect(page.locator('td', { hasText: 'mcp-org/backend' }).first()).toBeVisible();
		await expect(page.locator('td', { hasText: 'mcp-org/frontend' }).first()).toBeVisible();

		// The table should have columns for Repository Name, Schema Type, Last Synced, Status
		await expect(page.locator('th', { hasText: 'Repository Name' })).toBeVisible();
		await expect(page.locator('th', { hasText: 'Schema Type' })).toBeVisible();
	});

});
