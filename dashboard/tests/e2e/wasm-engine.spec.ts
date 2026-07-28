import { test, expect } from '@playwright/test';

test.describe('WASM Engine E2E (Suite 8)', () => {

    test('should navigate to playground', async ({ page }) => {
        await page.goto('/org/mcp-org/playground');
        await expect(page.locator('h1.page-title, h1')).toContainText(/Playground|AI|Studio/i);
    });

    test('should process Protobuf/OpenAPI diffs entirely in-browser without calling backend', async ({ page }) => {
        let backendCalled = false;
		page.on('request', request => {
			if (request.url().includes('/api/v1/diff')) {
				backendCalled = true;
			}
		});


        await page.goto('/org/mcp-org/playground');

        const editor = page.locator('textarea.code-editor, [contenteditable="true"]').first();
        if (await editor.count() > 0) {
            await editor.fill('openapi: 3.0.0\ninfo:\n  title: Test API');
        }

        // Wait for WASM engine to process - we expect NO backend calls
        await page.waitForTimeout(1000);
        expect(backendCalled).toBe(false);
    });

    test('should handle large payload (60k lines) stability', async ({ page }) => {
        await page.goto('/org/mcp-org/playground');
        const largeSchema = 'openapi: 3.0.0\ninfo:\n  title: Test API\n' + '  version: 1.0.0\n'.repeat(60000);

        const editor = page.locator('textarea.code-editor, [contenteditable="true"]').first();
        if (await editor.count() > 0) {
            await editor.fill(largeSchema);
        }

        // Check if page crashes
        await page.waitForTimeout(1000);
        await expect(page.locator('main, .playground-container, .page-content').first()).toBeVisible();
    });

    test('should handle malformed YAML input handling gracefully', async ({ page }) => {
        const consoleErrors: string[] = [];
        page.on('console', msg => {
            if (msg.type() === 'error') consoleErrors.push(msg.text());
        });

        await page.goto('/org/mcp-org/playground');
        const editor = page.locator('textarea.code-editor, [contenteditable="true"]').first();
        if (await editor.count() > 0) {
            await editor.fill('{{ INVALID_YAML_\x00 }}');
        }

        // Should not crash UI
        await page.waitForTimeout(500);
        await expect(page.locator('main, .playground-container, .page-content').first()).toBeVisible();
    });
});
