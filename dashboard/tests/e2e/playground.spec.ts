import { test, expect } from '@playwright/test';

/**
 * P12-T03: AI Playground E2E
 * Strategy: NO page.route(), NO mocking. All requests hit the live Go API.
 * The server returns a deterministic SSE fallback when SUBSTRATE_AI_BASE_URL is unset.
 */

const VALID_OPENAPI_SCHEMA = `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Success
`.trim();

test.describe('AI Playground (P12-T03)', () => {

    test.beforeEach(async () => {
        // Tests A & B require real LLM SSE stream. Skip when LLM_API_KEY is dummy/missing.
        const llmKey = process.env.LLM_API_KEY || '';
        const openrouterKey = process.env.OPENROUTER_API_KEY || '';
        if ((!llmKey || llmKey === 'dummy') && !openrouterKey) {
            const testInfo = test.info();
            if (testInfo.title.startsWith('Test A') || testInfo.title.startsWith('Test B')) {
                test.skip(true, 'Skipping: real LLM key required for AI SSE stream');
            }
        }
    });

    test('Test A — AI analysis stream returns findings panel', async ({ page }) => {
        await page.goto('/org/mcp-org/playground');

        // Assert page is the playground
        await expect(page.locator('h1.page-title')).toContainText(/Playground|AI|Studio/i);

        // Fill the current schema editor
        const editor = page.locator('textarea.code-editor').first();
        await editor.fill(VALID_OPENAPI_SCHEMA);

        // Click Analyze — hits real /api/v1/ai/analyze endpoint
        await page.getByRole('button', { name: /Analyze with Substrate AI/i }).click();

        // Wait for the SSE stream to complete and findings to render
        const analysisPanel = page.locator('.analysis-panel');
        await expect(analysisPanel).toBeVisible({ timeout: 15000 });

        // Server's real response (or deterministic fallback) must include BREAKING
        await expect(analysisPanel).toContainText(/BREAKING/i);
    });

    test('Test B — Apply Fix updates the proposed schema editor', async ({ page }) => {
        const consoleErrors: string[] = [];
        page.on('console', msg => {
            if (msg.type() === 'error') consoleErrors.push(msg.text());
        });

        await page.goto('/org/mcp-org/playground');

        const editor = page.locator('textarea.code-editor').first();
        await editor.fill(VALID_OPENAPI_SCHEMA);

        // Trigger analysis against real API
        await page.getByRole('button', { name: /Analyze with Substrate AI/i }).click();

        // Wait for auto-fix section to appear
        const autoFixSection = page.locator('.auto-fix-section');
        await expect(autoFixSection).toBeVisible({ timeout: 15000 });

        // Click Apply Fix
        await page.getByRole('button', { name: /Apply Fix/i }).click();

        // Short wait for state propagation
        await page.waitForTimeout(300);

        // The proposed schema editor (2nd textarea) should now contain the fix
        const proposedEditor = page.locator('textarea.code-editor').nth(1);
        await expect(proposedEditor).toHaveValue(/deprecated: true/i);

        // No critical Svelte errors
        const criticalErrors = consoleErrors.filter(e =>
            e.includes('$state') || e.includes('Cannot read properties of undefined')
        );
        expect(criticalErrors).toHaveLength(0);
    });

    test('Test C — Malformed input does not crash the page', async ({ page }) => {
        const consoleErrors: string[] = [];
        page.on('console', msg => {
            if (msg.type() === 'error') consoleErrors.push(msg.text());
        });

        await page.goto('/org/mcp-org/playground');

        // Fill with deliberately broken content
        const editor = page.locator('textarea.code-editor').first();
        await editor.fill('{{ INVALID_YAML_\x00 }}');

        // Click Analyze — real server will handle/reject this gracefully
        await page.getByRole('button', { name: /Analyze with Substrate AI/i }).click();

        // Give it time to resolve
        await page.waitForTimeout(1000);

        // Page must NOT crash — main container still visible
        await expect(page.locator('main, .playground-container, .page-content')).toBeVisible();

        // No uncaught Svelte errors
        const svelteErrors = consoleErrors.filter(e =>
            e.includes('$state') || e.includes('Cannot read properties of undefined')
        );
        expect(svelteErrors).toHaveLength(0);
    });
});
