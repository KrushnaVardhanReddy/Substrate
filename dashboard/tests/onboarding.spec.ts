import { test, expect } from '@playwright/test';

test.describe('Onboarding E2E Flow', () => {
	test('user can complete zero-to-one onboarding journey', async ({ page }) => {
		// Mock the API responses needed for /org/demo/graph to load
		await page.route('**/api/v1/repos/*', async route => {
			await route.fulfill({ json: [] });
		});
		await page.route('**/api/v1/graph/*', async route => {
			await route.fulfill({ json: [] });
		});

		// 1. Visit /onboarding
		await page.goto('/onboarding');

		// 2. Fill in a mock GitHub token
		const tokenInput = page.getByTestId('github-token-input');
		await tokenInput.waitFor({ state: 'visible', timeout: 5000 });
		await tokenInput.fill('ghp_mock_12345');

		// 3. Click the "Connect" button
		const connectBtn = page.getByTestId('connect-github-btn');
		await connectBtn.click();

		// 4. Wait for the simulated "Scanning Repositories" progress bar to complete
		const enterDashboardBtn = page.getByTestId('enter-dashboard-btn');
		await expect(enterDashboardBtn).toBeVisible({ timeout: 10000 });

		// 5. Click Enter Dashboard to trigger redirect
		await enterDashboardBtn.click();

		// 6. Assert that the URL successfully redirects to /org/demo/graph
		await page.waitForURL('**/org/demo/graph', { timeout: 5000 });
		expect(page.url()).toContain('/org/demo/graph');

		// 7. Assert that the Svelte Flow canvas container is visible in the DOM
		const mainCanvas = page.locator('.main-canvas');
		await expect(mainCanvas).toBeVisible();
	});
});
