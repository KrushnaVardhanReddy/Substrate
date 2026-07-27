import { test, expect } from '@playwright/test';

// Use process.env variables specifically for Playwright tests when running in CI without .env
const testOrgName = process.env.PUBLIC_ORG_NAME || 'myorg';

test('dashboard loads and displays the seeded data', async ({ page }) => {
	await page.goto('/');

	// Verify sidebar contains the dynamic repo from seeded data
	await expect(page.locator('.sidebar')).toContainText('mcp-org/frontend', { timeout: 10000 });

	await expect(page.locator('header')).toBeVisible();

	// Verify main content
	await expect(page.locator('h1.page-title')).toHaveText('Dependency Graph');
	
	// Since the default org is 'myorg' which has no graph data seeded, it should show empty state
	await expect(page.locator('.empty-state').first()).toContainText('No dependencies mapped');
	await expect(page.locator('.empty-state').nth(1)).toContainText('No recent schema changes');
});
