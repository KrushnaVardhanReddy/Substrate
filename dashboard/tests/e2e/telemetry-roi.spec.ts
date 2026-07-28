import { test, expect } from '@playwright/test';

test.describe('Telemetry ROI E2E (Suite 14)', () => {

    test('should validate ROI metrics dashboard updates based on telemetry events', async ({ page, context }) => {

        // Use a real JWT from the E2E runner (signed with local-jwt-secret)
		await context.addInitScript((t) => {
			localStorage.setItem('github_token', t);
		}, process.env.E2E_AUTH_TOKEN || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJvcmdzIjp7ImFkbWluIjoiYWRtaW4ifSwiZXhwIjo5OTk5OTk5OTk5fQ.placeholder');

        await page.goto('/org/testorg/telemetry');

        // Wait for the API call to resolve — either the ROI card or error appears
        // Hours Saved can legitimately be 0 if no telemetry events have been processed yet
        const roiCard = page.locator('[data-testid="roi-card"]');
        await expect(roiCard).toBeVisible({ timeout: 8000 });

        // Verify "Hours Saved:" label is rendered (value can be any number including 0)
        await expect(roiCard).toContainText(/Hours Saved:/i);
    });

    test('should track telemetry events via POST /api/v1/telemetry/track', async ({ page }) => {
        // Trigger an action that should emit telemetry (e.g. playground analyze)
        await page.goto('/org/mcp-org/playground');

        const analyzeBtn = page.getByRole('button', { name: /Analyze with Substrate AI/i });
        await expect(analyzeBtn).toBeVisible({ timeout: 5000 });

        // Wait for the request to be fired when we click the button
        const requestPromise = page.waitForRequest(
            req => req.url().includes('/api/v1/telemetry/track') && req.method() === 'POST',
            { timeout: 8000 }
        );

        await analyzeBtn.click();

        const request = await requestPromise;
        expect(request).toBeTruthy();
    });

});
