import { test, expect } from '@playwright/test';

test.describe('Phase 4 AI Diff Engine E2E', () => {
	test('AI Patch generation flow', async ({ page }) => {
		// Mock API response for diff
		await page.route('**/api/v1/diff/*', async route => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					base_schema: 'version: 1\nfields:\n  - id',
					head_schema: 'version: 2\nfields:\n  - id\n  - new_field'
				})
			});
		});

		// Mock API response for autofix to ensure it returns something we can stream
		await page.route('**/api/v1/ai/autofix', async route => {
			await route.fulfill({
				status: 200,
				contentType: 'text/event-stream',
				body: 'data: {"type": "thinking", "content": "Analyzing..."}\n\ndata: {"type": "finding", "severity": "BREAKING"}\n\ndata: {"type": "done"}\n\n'
			});
		});

		// Navigate to diff viewer
		await page.goto('/org/mcp-org/diff/diff-123');

		// Wait for viewer to load
		await expect(page.locator('.diff-viewer')).toBeVisible();

		// Click "Generate AI Patch"
		const generateBtn = page.getByRole('button', { name: /Generate AI Patch/i });
		await generateBtn.click();

		// Assert the UI receives the SSE stream and the "Apply Patch" button enables
		await expect(page.locator('.sse-stream')).toBeVisible();

		// The button text changes when we click it
		const applyBtn = page.getByRole('button', { name: /Apply Patch/i });
		await expect(applyBtn).toBeEnabled();

		await applyBtn.click();
		// Then it becomes the "Applied" button
		const appliedBtn = page.getByRole('button', { name: /Applied/i });
		await expect(appliedBtn).toBeDisabled();
	});
});
