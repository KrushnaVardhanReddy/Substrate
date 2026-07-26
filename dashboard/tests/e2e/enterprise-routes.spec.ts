import { test, expect } from '@playwright/test';

test.describe('Enterprise Dashboard Routes (Suite 11)', () => {

    const ORG = 'admin'; // Based on the spec `/org/admin/telemetry`, etc.

    test('should render /org/admin/graph correctly', async ({ page }) => {
        // Intercept API call to avoid 404 from backend missing


        await page.goto(`/org/${ORG}/graph`);

        await page.waitForTimeout(500);

        await expect(page.locator('.svelte-flow, .graph-container').first()).toBeVisible({ timeout: 5000 });
    });

    test('should render /org/admin/catalog correctly', async ({ page }) => {
        await page.goto(`/org/${ORG}/catalog`);
        await page.waitForTimeout(500);
        await expect(page.locator('body')).not.toContainText('Internal Error');
        await expect(page.locator('h1, h2, .page-title').filter({ hasText: /Catalog|Services|APIs/i }).first()).toBeVisible({ timeout: 5000 });
    });

    test('should render /org/admin/settings correctly', async ({ page }) => {
        await page.goto(`/org/${ORG}/settings`);
        await page.waitForTimeout(500);
        await expect(page.locator('body')).not.toContainText('Internal Error');
        await expect(page.locator('h1, h2, .page-title').filter({ hasText: /Settings/i }).first()).toBeVisible({ timeout: 5000 });
    });

    test('should render /org/admin/telemetry correctly', async ({ page }) => {
        await page.goto(`/org/${ORG}/telemetry`);
        await page.waitForTimeout(500);

        const title = page.locator('h1, h2, .page-title').filter({ hasText: /Telemetry/i }).first();
        // Since we know they will fail, we can assert what we expect per spec,
        // We simulate natural TDD expectation and allow the failure.
        await expect(title).toBeVisible({ timeout: 2000 });
    });

    test('should render /org/admin/webhooks correctly', async ({ page }) => {
        await page.goto(`/org/${ORG}/webhooks`);
        await page.waitForTimeout(500);

        const title = page.locator('h1, h2, .page-title').filter({ hasText: /Webhooks/i }).first();
        await expect(title).toBeVisible({ timeout: 2000 });
    });

});
