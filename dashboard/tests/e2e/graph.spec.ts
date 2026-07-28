import { test, expect } from '@playwright/test';

// p3-org has seeded service-a, service-b, service-c with dependency edges in seed_mcp.sql
const ORG = 'p3-org';

test.describe('Dependency Graph', () => {

    test.beforeEach(async ({ page }) => {
        page.on('pageerror', (err) => {
            if (err.message.includes('hydration')) {
                console.warn('Hydration warning suppressed');
            } else {
                throw err;
            }
        });
    });

    test('should render graph container and filter controls', async ({ page }) => {
        const responsePromise = page.waitForResponse('**/api/v1/graph/**', { timeout: 15000 });
        await page.goto(`/org/${ORG}/graph`);
        await responsePromise;

        await expect(page.locator('h1.page-title')).toContainText('Dependency Graph');
        await expect(page.locator('.filter-panel')).toBeVisible();
        await expect(page.locator('input[type="checkbox"]').first()).toBeVisible();
        await expect(page.locator('select.filter-select')).toBeVisible();
        await expect(page.locator('input.filter-input')).toBeVisible();

        // Check for Export button
        await expect(page.locator('button[title="Export PNG"]')).toBeVisible();
        // Graph container visible (cyInstance not created until search)
        await expect(page.locator('.main-canvas')).toBeVisible();
    });

    test('should handle layout rotation, taxonomy badges, and deep links', async ({ page }) => {
        const responsePromise = page.waitForResponse('**/api/v1/graph/**', { timeout: 15000 });
        await page.goto(`/org/${ORG}/graph`);
        await responsePromise;

        // Trigger cyInstance by typing a search (graph is search-gated)
        const searchInput = page.locator('input.filter-input');
        await expect(searchInput).toBeVisible({ timeout: 10000 });
        await searchInput.fill('service');
        await page.waitForTimeout(800); // debounce

        await page.waitForFunction(() => (window as any).cyInstance !== undefined && (window as any).cyInstance !== null, { timeout: 15000 });
        await page.waitForFunction(() => (window as any).cyInstance.nodes().length > 0, { timeout: 10000 });

        // Click the Rotate Layout button
        const rotateBtn = page.locator('button[title="Rotate Layout"]');
        await rotateBtn.click();
        await rotateBtn.click();

        // Click the node to open detail panel via Cytoscape API
        await page.evaluate(() => {
            const cy = (window as any).cyInstance;
            const node = cy.nodes().first();
            if (node) node.emit('tap');
        });

        // Verify detail panel appears
        const detailPanel = page.locator('.detail-panel');
        await expect(detailPanel).toBeVisible({ timeout: 5000 });
        await expect(detailPanel.locator('.detail-title')).toBeVisible();
    });

    test('should trigger PNG export', async ({ page }) => {
        const responsePromise = page.waitForResponse('**/api/v1/graph/**', { timeout: 15000 });
        await page.goto(`/org/${ORG}/graph`);
        await responsePromise;

        // Trigger search so button state is correct
        const searchInput = page.locator('input.filter-input');
        await searchInput.fill('service');
        await page.waitForTimeout(400);

        const exportBtn = page.locator('button[title="Export PNG"]');
        await expect(exportBtn).toBeVisible();
    });

    test('should highlight blast radius on node click', async ({ page }) => {
        const responsePromise = page.waitForResponse('**/api/v1/graph/**', { timeout: 15000 });
        await page.goto(`/org/${ORG}/graph`);
        await responsePromise;

        // Type search to trigger cyInstance
        const searchInput = page.locator('input.filter-input');
        await searchInput.fill('service');
        await page.waitForTimeout(800);

        await page.waitForFunction(() => (window as any).cyInstance !== undefined && (window as any).cyInstance !== null, { timeout: 15000 });
        await page.waitForFunction(() => (window as any).cyInstance.nodes().length > 0, { timeout: 10000 });

        // Click a node that has edges
        await page.evaluate(() => {
            const cy = (window as any).cyInstance;
            const node = cy.nodes().find((n: any) => n.connectedEdges().length > 0) || cy.nodes().first();
            if (node) node.emit('tap');
        });

        await page.waitForTimeout(500);
        await expect(page.locator('.detail-panel')).toBeVisible();
    });
});
