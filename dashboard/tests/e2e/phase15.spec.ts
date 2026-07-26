import { test, expect } from '@playwright/test';

test.describe('Phase 15 Monetization E2E', () => {
	test.beforeEach(async ({ page }) => {
		// Mock layout requests that might cause unhandled promise rejections / timeouts

		// Setup local storage to bypass token parsing errors in layout
		await page.addInitScript(() => {
			const token = btoa(JSON.stringify({ orgs: { 'mcp-org': 'admin' } }));
			window.localStorage.setItem('github_token', `header.${token}.signature`);
		});

		// Ignore hydration errors in dev mode
		page.on('pageerror', (err) => {
			if (
				err.message.includes('hydration') ||
				err.message.includes('No matching export') ||
				err.message.includes('Svelte') ||
				err.message.includes('lifecycle_outside_component')
			) {
				return;
			}
			console.error(err);
		});
	});

	test('should render an active policy and claims history', async ({ page }) => {
		await page.goto('/org/mcp-org/settings/insurance');

		await expect(page.getByText('Policy ID:')).toBeVisible();
		await expect(page.getByText('Limit:')).toBeVisible();
		await expect(page.getByText('$50000.00')).toBeVisible();

		await expect(page.getByRole('heading', { name: 'Claim History' })).toBeVisible();
		await expect(page.getByText('$150.00')).toBeVisible();
		await expect(page.getByText('APPROVED')).toBeVisible();
	});

	test('should successfully file a new claim', async ({ page }) => {
		await page.goto('/org/mcp-org/settings/insurance');

		await expect(page.getByText('Policy ID:')).toBeVisible();

		await page.fill('#prUrl', 'https://github.com/foo/bar/pull/2');
		await page.fill('#incidentDate', '2023-10-15');
		await page.fill('#amount', '150.50');

		await page.click('button:has-text("File Claim")');

		await expect(page.getByText('$150.50')).toBeVisible();
		await expect(page.getByText('PENDING')).toBeVisible();
	});
});
