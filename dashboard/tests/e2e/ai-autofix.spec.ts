import { test, expect } from '@playwright/test';

/**
 * AI Autofix API E2E (Suite 10)
 * Strategy: NO page.route(), NO mocking. All requests hit the live Go API.
 * The server returns a deterministic SSE response including fix code.
 */

test.describe('AI Autofix API E2E (Suite 10)', () => {

    test('should return finding via POST /api/v1/ai/analyze in playground UI', async ({ page }) => {
        await page.goto('/org/mcp-org/playground');

        // Fill an editor with valid content
        const editor = page.locator('textarea.code-editor').first();
        await editor.fill('type User { id: ID! name: String }');

        // Click Analyze — real API call, no intercept
        const analyzeBtn = page.getByRole('button', { name: /Analyze with Substrate AI/i });
        await analyzeBtn.click();

        // Wait for the real SSE stream to return findings
        const findings = page.locator('.analysis-panel');
        await expect(findings).toBeVisible({ timeout: 15000 });
        await expect(findings).toContainText(/BREAKING/i);
    });

    test('should apply fix patch via UI after POST /api/v1/ai/analyze completes', async ({ page }) => {
        await page.goto('/org/mcp-org/playground');

        const editor = page.locator('textarea.code-editor').first();
        await editor.fill('message User { string id = 1; }');

        // Click Analyze — real API, SSE response includes fix code
        const analyzeBtn = page.getByRole('button', { name: /Analyze with Substrate AI/i });
        await analyzeBtn.click();

        // Wait for fix section to appear (server returns fix in SSE stream)
        const autoFixSection = page.locator('.auto-fix-section');
        await expect(autoFixSection).toBeVisible({ timeout: 15000 });

        // Apply the fix
        const applyBtn = page.getByRole('button', { name: /Apply Fix/i });
        await expect(applyBtn).toBeVisible();
        await applyBtn.click();

        // Proposed schema editor must be updated with fix content
        const proposedEditor = page.locator('textarea.code-editor').nth(1);
        await expect(proposedEditor).toHaveValue(/deprecated: true/i);
    });

});
