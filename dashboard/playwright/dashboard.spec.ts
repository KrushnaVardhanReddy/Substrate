import { test, expect } from '@playwright/test';

test('Dashboard UI skeleton test', async ({ page }) => {
    await page.goto('/');
    // The page doesn't have a <title> element set by SvelteKit currently,
    // so we verify the main content is loaded instead.
    await expect(page.locator('text=Dependency Graph').first()).toBeVisible();
});