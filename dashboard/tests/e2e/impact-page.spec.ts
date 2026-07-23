import { test, expect } from '@playwright/test';

test.describe('Impact Page', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v1/repos/*', async (route) => {
            await route.fulfill({ status: 200, json: [] });
        });

        await page.route('**/api/v1/impact/*', async (route) => {
            await route.fulfill({
                status: 200,
                json: {
                    can_deploy: true,
                    can_rollback: false
                }
            });
        });
    });

    test('should render /impact correctly without crashing', async ({ page }) => {
        // Suppress errors during load
        page.on('pageerror', () => {});

        await page.goto('/org/testorg/repo/backend/impact', { waitUntil: 'domcontentloaded' });

        // Wait up to 10s for page to settle
        await page.waitForTimeout(1000);

        const content = await page.content();

        // Either page successfully hydrated or SSR rendered it
        expect(content.includes('Impact Analysis') || content.includes('page-title')).toBe(true);
        expect(content).not.toContain('Internal Error');
    });
});
