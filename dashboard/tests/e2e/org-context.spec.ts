import { test, expect } from '@playwright/test';

const testOrgName = process.env.PUBLIC_ORG_NAME || 'mcp-org';

test.describe('UI Org Context & API Keys', () => {

	test('1. API Keys View (Org Scoped)', async ({ page }) => {
		await page.goto(`/org/${testOrgName}/apikeys`);

		// Assert Page header displays "API Keys"
		await expect(page.locator('h1.page-title')).toContainText('API Keys');

		// Assert UI contains a "Generate New Key" button
		const generateBtn = page.getByRole('button', { name: 'Generate New Key' });
		await expect(generateBtn).toBeVisible();

		// Assert Table or empty state renders correctly
		const emptyState = page.locator('.empty-state');
		await expect(emptyState).toBeVisible();
		await expect(emptyState).toContainText('No API keys generated yet.');

		// Assert Sidebar .active item is correctly set on "API Keys"
		const activeSidebarItem = page.locator('.sidebar .nav-item.active');
		await expect(activeSidebarItem).toHaveText('API Keys');
	});

	test('2. Context Persistence (Marketplace)', async ({ page }) => {
		await page.goto(`/org/${testOrgName}`);

		// Click "Marketplace" in the sidebar
		await page.getByRole('link', { name: 'Marketplace' }).click();

		// Assert URL is now /org/mcp-org/marketplace
		await expect(page).toHaveURL(new RegExp(`/org/${testOrgName}/marketplace`));

		// Assert the sidebar still displays the full suite of organization links
		const sidebar = page.locator('.sidebar');
		await expect(sidebar).toContainText('Repositories');
		await expect(sidebar).toContainText('Dependency Graph');
		await expect(sidebar).toContainText('API Keys');
	});

	test('3. Context Persistence (AI Playground)', async ({ page }) => {
		await page.goto(`/org/${testOrgName}`);

		// Click "✨ AI Playground" in the sidebar
		await page.getByRole('link', { name: '✨ AI Playground' }).click();

		// Assert URL is now /org/mcp-org/playground
		await expect(page).toHaveURL(new RegExp(`/org/${testOrgName}/playground`));

		// Assert the sidebar still displays the full suite of organization links
		const sidebar = page.locator('.sidebar');
		await expect(sidebar).toContainText('Repositories');
		await expect(sidebar).toContainText('Dependency Graph');
		await expect(sidebar).toContainText('API Keys');
	});
});
