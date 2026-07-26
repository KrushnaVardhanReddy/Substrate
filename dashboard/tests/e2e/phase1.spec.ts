import { test, expect } from '@playwright/test';

test.describe('Phase 1 Core Diff Engine', () => {
    test.beforeEach(async ({ page }) => {
        await page.addInitScript(() => {
            window.localStorage.setItem('auth_token', 'local-dev-token');
        });

        page.on('pageerror', (err) => {
            if (err.message.includes('hydration')) {
                console.warn('Hydration error suppressed', err);
            } else {
                throw err;
            }
        });
    });

    test('should render Repositories list correctly', async ({ page }) => {
        await page.goto('/org/mcp-org', { timeout: 10000 });
        await expect(page.locator('h1.page-title')).toContainText('Repositories', { timeout: 5000 });
        await expect(page.locator('body')).toContainText('core-repo', { timeout: 3000 });
    });

    test('should render graph and cytoscape instance for core-repo', async ({ page }) => {
        const responsePromise = page.waitForResponse('**/api/v1/graph/*', { timeout: 5000 });
        await page.goto('/org/mcp-org/graph', { timeout: 10000 });
        await responsePromise;

        // Wait for graph container and search functionality to be visible
        await expect(page.locator('.filter-panel')).toBeVisible({ timeout: 5000 });

        await page.fill('input.filter-input', '/');

        await page.waitForFunction(() => (window as any).cyInstance !== undefined && (window as any).cyInstance !== null, { timeout: 3000 });
        await expect(page.locator('.main-canvas')).toBeVisible({ timeout: 2000 });

        // Interact via cytoscape instance as required by rules
        await page.evaluate(() => {
            const cy = (window as any).cyInstance;
            if (cy) {
                const node = cy.nodes().first();
                if (node) {
                    node.emit('tap');
                }
            }
        });

        await page.waitForFunction(() => (window as any).cyInstance.nodes().length > 0, { timeout: 2000 });
    });
});
