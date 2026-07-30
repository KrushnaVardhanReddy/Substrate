import { test, expect } from '@playwright/test';

// Use process.env variables specifically for Playwright tests when running in CI without .env
const testOrgName = process.env.PUBLIC_ORG_NAME || 'e2e-org';

test.describe('Phase 16 UI Validation', () => {
	test('navigates to catalog page and tests guides and sandbox', async ({ page }) => {
		// Set local dev token in localStorage so fetch gets authorized
		await page.addInitScript(() => {
			window.localStorage.setItem('token', 'local-dev-token');
		});

		// Navigate to catalog page
		await page.goto('/org/e2e-org/catalog/guide-api-repo');

		// Wait for the guides tab to be visible and click it
		const guidesTab = page.locator('#guides-tab');
		await expect(guidesTab).toBeVisible({ timeout: 10000 });
		await guidesTab.click();

		// Assert "Authentication Guide" is visible
		const authGuide = page.locator('.guide-link', { hasText: 'Authentication Guide' });
		await expect(authGuide).toBeVisible({ timeout: 10000 });
		await authGuide.click();

		// Check the content renders
		const guideContent = page.locator('.guide-content');
		await expect(guideContent).toContainText('Use Bearer tokens');

		// Try a sandbox request
		const sandboxToggle = page.locator('.sandbox-toggle');
		await expect(sandboxToggle).toBeVisible();
		await sandboxToggle.click();

		// Check request builder is visible
		await expect(page.locator('.request-builder')).toBeVisible();

		const pathInput = page.locator('.path-input');
		await pathInput.fill('/test');

		const sendBtn = page.locator('.btn-send');
		await sendBtn.click();

		// Assert `#sandbox-response` is visible
		const sandboxResponse = page.locator('#sandbox-response');
		await expect(sandboxResponse).toBeVisible({ timeout: 10000 });
	});
});
