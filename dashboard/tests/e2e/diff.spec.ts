import { test, expect } from '@playwright/test';

test.describe('Diff Viewer & Sign Out E2E', () => {
    test('DiffViewer renders mock data and Sign Out works', async ({ page }) => {
        // Go to diff page
        await page.goto('/diff/123');

        // Wait for page to load
        await expect(page.locator('.diff-page')).toBeVisible();
        await expect(page.locator('h1.page-title')).toHaveText('Schema Diff Viewer');

        // Verify diff rows are rendered
        await expect(page.locator('.diff-row').first()).toBeVisible();
        await expect(page.locator('.diff-cell.left-content.delete').first()).toBeVisible();
        await expect(page.locator('.diff-cell.right-content.add').first()).toBeVisible();

        // Test Sign Out
        const signoutBtn = page.locator('.signout-btn');
        await expect(signoutBtn).toBeVisible();

        await signoutBtn.click();

        // Ensure redirected to /onboarding
        await expect(page).toHaveURL(/.*\/onboarding/);
    });
});
