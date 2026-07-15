import { test, expect } from '@playwright/test';

// Use process.env variables specifically for Playwright tests when running in CI without .env
const testOrgName = process.env.PUBLIC_ORG_NAME || 'myorg';

test('dashboard loads and displays the empty state', async ({ page }) => {
	// Mock the API response
	await page.route('**/api/v1/repos/*', async route => {
		const json = [{ id: '1', name: 'repo-1', full_name: 'org/repo-1' }];
		await route.fulfill({ json });
	});

	await page.goto('/');

	// Verify sidebar contains the dynamic repo
	await expect(page.locator('.sidebar')).toContainText('org/repo-1');

	await expect(page.locator('header')).toBeVisible();

	// Verify main content empty state
	await expect(page.locator('h1.page-title')).toHaveText('Dependency Graph');
	await expect(page.locator('.empty-state').first()).toContainText('No dependencies mapped yet');
	await expect(page.locator('.empty-state').nth(1)).toContainText('No recent schema changes detected.');
});
