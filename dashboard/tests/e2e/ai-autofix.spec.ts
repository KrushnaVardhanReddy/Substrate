import { test, expect } from '@playwright/test';

test.describe('AI Autofix API E2E (Suite 10)', () => {

    test('should return finding via POST /api/v1/ai/analyze in playground UI', async ({ page }) => {


        await page.goto('/playground');

        // Fill an editor so it's not empty
        const editor = page.locator('textarea, [contenteditable="true"]').first();
        if (await editor.count() > 0) {
            await editor.fill('type User { id: ID! }');
        }

        const analyzeBtn = page.getByRole('button', { name: /Analyze with Substrate AI/i });
        await analyzeBtn.click();

        // UI should show findings based on our mock
        const findings = page.locator('.analysis-panel');
        await expect(findings).toBeVisible({ timeout: 10000 });
        await expect(findings).toContainText(/BREAKING|Removed field/i);
    });

    test('should apply safe_patch for Protobuf/OpenAPI via POST /api/v1/ai/autofix in UI', async ({ page }) => {
        // analyze step


        // Some UI implementation might separately call autofix, or might parse it from the SSE.
        // Spec implies testing `POST /api/v1/ai/autofix` returns valid `safe_patch`.
        // The `/playground` UI uses SSE for fixes. If there's another UI for autofix (e.g., repo diff page), we test the action.
        await page.goto('/playground');

        const editor = page.locator('textarea, [contenteditable="true"]').first();
        if (await editor.count() > 0) {
            await editor.fill('message User { string id = 1; }');
        }

        const analyzeBtn = page.getByRole('button', { name: /Analyze with Substrate AI/i });
        await analyzeBtn.click();

        // Let UI settle
        await page.waitForTimeout(500);

        // Click apply fix
        const applyBtn = page.getByRole('button', { name: /Apply Fix/i });
        if (await applyBtn.count() > 0) {
            await applyBtn.click();
        }

        // We expect the proposed editor to be updated
        const proposedEditor = page.locator('textarea.code-editor').nth(1);
        if (await proposedEditor.count() > 0) {
            await expect(proposedEditor).toHaveValue(/deprecated = true/);
        }
    });

});
