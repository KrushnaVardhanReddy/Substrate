import { test, expect } from '@playwright/test';

test.describe('Enterprise Dashboard', () => {
	test.beforeEach(async ({ page }) => {
		page.on('pageerror', (err) => {
			if (err.message.includes('hydration')) {
				console.warn('Hydration warning suppressed');
			} else {
				throw err;
			}
		});

		await page.goto('/org/mcp-org/enterprise');
	});

	test('should display webhooks tab and active webhook', async ({ page }) => {
		const webhooksTab = page.locator('button', { hasText: 'Webhooks' });
		await expect(webhooksTab).toBeVisible();
		await webhooksTab.click();

		const activeWebhook = page.locator('text=Active Webhook:');
		await expect(activeWebhook).toBeVisible();
	});

	test('should edit and save custom CEL rule', async ({ page }) => {
		const customRulesTab = page.locator('button', { hasText: 'Custom Rules' });
		await expect(customRulesTab).toBeVisible();
		await customRulesTab.click();

		const celTextarea = page.locator('textarea[name="cel-rule"]');
		await expect(celTextarea).toBeVisible();

		await celTextarea.fill('request.auth.claims.group == "admin"');

		const saveButton = page.locator('button:has-text("Save Rule")');
		await expect(saveButton).toBeVisible();
		await saveButton.click();

		const successMsg = page.locator('text="Rule saved successfully"');
		await expect(successMsg).toBeVisible();
	});

	test('should display drift report and resolve button', async ({ page }) => {
		const driftTab = page.locator('button', { hasText: 'Drift Detection' });
		await expect(driftTab).toBeVisible();
		await driftTab.click();

		const driftReport = page.locator('text="Drift Report"');
		await expect(driftReport).toBeVisible();

		const resolveButton = page.locator('button:has-text("Resolve")');
		await expect(resolveButton).toBeVisible();
	});
});
