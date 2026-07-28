import { test, expect } from '@playwright/test';

test.describe('API Keys Flow', () => {
	const orgName = 'test-org';

	test('can view API keys page', async ({ page }) => {
		// Mock the network calls to avoid hitting the actual backend which might not be running in this sandbox.
		await page.route(`**/api/v1/org/${orgName}/apikeys**`, async route => {
			if (route.request().method() === 'GET') {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify([{
						id: '1',
						org_id: 'org1',
						name: 'Existing Key',
						prefix: 'abcd...',
						created_at: new Date().toISOString()
					}])
				});
			} else if (route.request().method() === 'POST') {
				await route.fulfill({
					status: 201,
					contentType: 'application/json',
					body: JSON.stringify({
						key: {
							id: '2',
							org_id: 'org1',
							name: 'Playwright Test Key',
							prefix: 'efgh...',
							created_at: new Date().toISOString()
						},
						raw_token: '1234567890123456789012345678901234567890123456789012345678901234'
					})
				});
			} else if (route.request().method() === 'DELETE') {
				console.log('DELETE CALLED');
				await route.fulfill({
					status: 204
				});
			}
		});

		await page.goto(`/org/${orgName}/apikeys`);
		await expect(page.locator('h1')).toHaveText('API Keys');

		// Check generate key button
		const generateBtn = page.locator('button:has-text("Generate New Key")');
		await generateBtn.click();

		// Check modal
		await expect(page.locator('h3').filter({ hasText: 'Generate New API Key' })).toBeVisible();

		// Fill in key name
		await page.locator('input#key-name').fill('Playwright Test Key');
		await page.locator('button.btn-primary:has-text("Generate"):not(:has-text("New"))').click();

		// Should show raw key
		await expect(page.locator('h3').filter({ hasText: 'Your New API Key' })).toBeVisible();
		const rawKey = await page.locator('.key-display code').textContent();
		expect(rawKey).toBeTruthy();
		expect(rawKey?.length).toBe(64); // 32 bytes hex

		// Close modal
		await page.locator('button:has-text("I have copied my key")').click();

		// Check key is in list
		const tableRows = page.locator('.keys-table tbody tr');
		await expect(tableRows).toHaveCount(2); // Existing Key + Playwright Test Key
		await expect(tableRows.nth(0)).toContainText('Playwright Test Key');

		// Delete key
		await tableRows.nth(0).locator('button:has-text("Revoke")').click();
		await page.waitForTimeout(500); // Give svelte time to reactively remove the row
		await expect(tableRows).toHaveCount(1);
	});
});
