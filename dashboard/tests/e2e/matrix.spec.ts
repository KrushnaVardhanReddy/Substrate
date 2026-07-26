import { test, expect } from '@playwright/test';

test.describe('Compatibility Matrix', () => {


	test('should render matrix and show breaking change alert on click', async ({ page }) => {
		// Navigate directly to the matrix page (SSR is disabled in tests)
		await page.goto('/org/testorg/matrix');

		// Wait for the matrix page to load
		await expect(page.locator('h1.page-title')).toContainText('Compatibility Matrix');

		// Assert that the matrix table renders
		const matrixTable = page.locator('.matrix-table');
		await expect(matrixTable).toBeVisible();

		// Ensure the provider name "users-api" is visible
		await expect(matrixTable.locator('.provider-name-cell', { hasText: 'users-api' })).toBeVisible();

		// Click the incompatible cell
		const incompatibleCell = page.locator('.danger-icon').first();
		await expect(incompatibleCell).toBeVisible();
		await incompatibleCell.click({ force: true });

		// Assert that the detail panel opens
		const detailPanel = page.locator('.cell-detail-panel');
		await expect(detailPanel).toBeVisible({ timeout: 10000 });

		// Assert the detail panel contains the "Breaking Change Detected" alert
		await expect(detailPanel.locator('strong')).toContainText('Breaking Change Detected');
	});
});
