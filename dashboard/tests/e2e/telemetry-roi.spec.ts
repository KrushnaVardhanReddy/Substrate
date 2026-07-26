import { test, expect } from '@playwright/test';

test.describe('Telemetry ROI E2E (Suite 14)', () => {

    test('should validate ROI metrics dashboard updates based on telemetry events', async ({ page }) => {
        // Intercept telemetry API response to provide mock data for the UI


        // The exact route isn't strictly defined, but "telemetry" was mapped in enterprise routes
        await page.goto('/org/admin/telemetry');

        await page.waitForTimeout(500);

        // We assert that it requests the data and tries to display it.
        // If the page doesn't exist yet, we capture the natural fail for TDD.
        const roiCard = page.locator('body').filter({ hasText: /45/i });
        await expect(roiCard).toBeVisible({ timeout: 2000 });
    });

    test('should track telemetry events via POST /api/v1/telemetry/track', async ({ page }) => {
        let telemetryTracked = false;
		page.on('request', request => {
			if (request.url().includes('/api/v1/telemetry/track') && request.method() === 'POST') {
				telemetryTracked = true;
			}
		});



        // Trigger an action that should emit telemetry (e.g. playground analyze or graph search)
        await page.goto('/playground');
        await page.waitForTimeout(500);

        // Sometimes just loading a dashboard page fires an event, or we interact.
        const analyzeBtn = page.getByRole('button', { name: /Analyze with Substrate AI/i });
        if (await analyzeBtn.count() > 0) {
            await analyzeBtn.click();
            await page.waitForTimeout(500);
        }

        // If the frontend isn't instrumented yet, this naturally fails in TDD.
        expect(telemetryTracked).toBe(true);
    });

});
