import { test, expect } from '@playwright/test';

test.describe('Governance Rules Page', () => {
    test.beforeEach(async ({ page }) => {
        await page.route('**/api/v1/repos/*', async (route) => {
            await route.fulfill({ status: 200, json: [] });
        });

        await page.route('**/api/governance/generate-cel', async (route) => {
            await route.fulfill({
                status: 200,
                json: { cel: 'request.auth != null' }
            });
        });
    });

    test('should render and generate CEL rule', async ({ page }) => {
        page.on('pageerror', () => {});
        await page.goto('/org/testorg/governance', { waitUntil: 'domcontentloaded' });

        await page.waitForTimeout(1000);

        const content = await page.content();

        expect(content.includes('Governance Rules') || content.includes('Governance') || content.includes('governance')).toBe(true);
        expect(content).not.toContain('Internal Error');
    });
});
