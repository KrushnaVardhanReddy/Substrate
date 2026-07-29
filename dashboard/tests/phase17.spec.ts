import { test, expect } from '@playwright/test';

// Use process.env variables specifically for Playwright tests when running in CI without .env
const testOrgName = process.env.PUBLIC_ORG_NAME || 'mcp-org';

test.describe('Phase 17 UI Validation', () => {
	test('navigates to /org/mgr-org and asserts "Health Grade" column', async ({ page }) => {
		await page.goto('/org/mgr-org');
		await expect(page.locator('th', { hasText: 'Health Grade' })).toBeVisible();
	});

	test('asserts payments-api shows a coloured grade badge and navigates to scorecard', async ({ page }) => {
		await page.goto('/org/mgr-org');
		// Wait for the repo list to load
		await expect(page.locator('td', { hasText: 'payments-api' })).toBeVisible({ timeout: 10000 });

		// Assert grade badge is visible
		const gradeBadge = page.locator('[data-testid="health-grade-badge"]').first();
		await expect(gradeBadge).toBeVisible();

		// Click the grade badge
		await gradeBadge.click();

		// Assert it navigates to scorecard detail view with the breakdown chart
		await expect(page).toHaveURL(/\/org\/mgr-org\/scorecard\/payments-api.*/);
		await expect(page.locator('[data-testid="breakdown-chart"]')).toBeVisible();
	});

	test('navigates to /org/mgr-org/graph and asserts Event Timeline panel lists the deployment event', async ({ page }) => {
		// Set local dev token in localStorage so fetch gets authorized
		await page.addInitScript(() => {
			window.localStorage.setItem('local-dev-token', 'local-dev-token');
		});

		await page.goto('/org/mgr-org/graph');

		// Wait for the timeline panel to appear
		const timelinePanel = page.locator('.timeline-panel');
		await expect(timelinePanel).toBeVisible({ timeout: 10000 });
		await expect(timelinePanel.locator('.timeline-title')).toHaveText('Event Timeline');

		// Check the event list for the "deployed v1.0.0" description (as seeded by the go script)
		const eventDesc = timelinePanel.locator('.event-desc', { hasText: 'deployed v1.0.0' }).first();
		await expect(eventDesc).toBeVisible({ timeout: 10000 });
	});
});
