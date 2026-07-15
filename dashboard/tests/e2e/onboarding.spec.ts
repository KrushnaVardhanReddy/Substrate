import { test, expect } from '@playwright/test';

test.describe('Onboarding Wizard E2E', () => {
    test('User completes the onboarding flow successfully', async ({ page }) => {
        // Step 1: Navigate to onboarding
        await page.goto('/onboarding');
        
        // Assert Step 1 content
        await expect(page.getByText('Connect GitHub', { exact: false })).toBeVisible();
        
        // Click connect button
        const connectBtn = page.getByRole('button', { name: /connect github/i });
        await expect(connectBtn).toBeVisible();
        await connectBtn.click();
        
        // Step 2: Assert scanning state
        await expect(page.locator('text=Scanning your Repositories').first()).toBeVisible();
        
        // Wait for step 3 to appear (progress bar + delay takes ~3.5 seconds)
        // We'll give it a generous timeout of 10s to avoid flakes
        await expect(page.locator('text=Ready to Launch')).toBeVisible({ timeout: 10000 });
        
        // Step 3: Enter Dashboard
        const enterBtn = page.getByRole('button', { name: /enter dashboard/i });
        await expect(enterBtn).toBeVisible();
        await enterBtn.click();
        
        // Ensure redirected back to root path (which might redirect to /playground or /org if unauthenticated, etc. But we check for '/' base redirect action)
        await page.waitForURL('**/', { timeout: 5000 }).catch(() => page.waitForURL('**/playground', { timeout: 5000 }));
        
        // Let's just assume we hit a valid page and the URL changed off of /onboarding
        expect(page.url()).not.toContain('/onboarding');
    });
});
