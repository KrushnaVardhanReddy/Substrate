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

		// 2. Select GitHub provider
		const githubBtn = page.getByTestId('provider-github-btn');
		await githubBtn.waitFor({ state: 'visible', timeout: 5000 });
		await githubBtn.click();

		// 3. Fill in a mock GitHub token
		const tokenInput = page.getByTestId('token-input');
		await tokenInput.waitFor({ state: 'visible', timeout: 5000 });
		await tokenInput.fill('ghp_mock_12345');

		// 4. Click the "Connect" button
		const connectBtn = page.getByTestId('connect-provider-btn');
		await connectBtn.click();

		// 5. Wait for the simulated "Scanning Repositories" progress bar to complete
		const enterDashboardBtn = page.getByTestId('enter-dashboard-btn');
		await expect(enterDashboardBtn).toBeVisible({ timeout: 10000 });

		// 6. Click Enter Dashboard to trigger redirect
		await enterDashboardBtn.click();

		// 7. Assert that the URL successfully redirects to /org/*/graph
		await page.waitForURL('**/org/*/graph', { timeout: 5000 });
		expect(page.url()).toMatch(/\/org\/.*\/graph/);

		// 8. Assert that the Svelte Flow canvas container is visible in the DOM
		const mainCanvas = page.locator('.main-canvas');
		await expect(mainCanvas).toBeVisible();
	});
});
