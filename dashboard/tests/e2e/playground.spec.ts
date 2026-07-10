import { test, expect } from '@playwright/test';

test.describe('AI Playground', () => {
	test.beforeEach(async ({ page }) => {
		await page.route('**/api/v1/ai/analyze', async (route) => {
			// Mock the Server-Sent Events (SSE) stream returned by the AI bridge
			const sseBody = [
				`data: {"type":"thinking","content":"Analyzing GraphQL schema for breaking changes..."}`,
				``,
				`data: {"type":"finding","severity":"BREAKING","content":"Field 'user_id' was removed from User type."}`,
				``,
				`data: {"type":"fix","code":"type User {\\n  id: ID!\\n  user_id: String! @deprecated(reason: \\"Use id instead\\")\\n  name: String\\n  email: String\\n}"}`,
				``,
				`data: {"type":"done"}`,
				``,
				``
			].join('\n');

			await route.fulfill({
				status: 200,
				contentType: 'text/event-stream',
				body: sseBody
			});
		});
	});

	test('should stream AI analysis and apply auto-fix', async ({ page }) => {
		// Navigate to the playground
		await page.goto('/playground');

		// Check if we are on the playground
		await expect(page.locator('h1.page-title')).toContainText('AI Schema Validator Playground');
		
		// Let the UI settle
		await page.waitForTimeout(500);

		// Click the Analyze button
		await page.getByRole('button', { name: /Analyze with Substrate AI/i }).click();

		// Verify that the thinking text appears (it might appear instantly due to the mock)
		await expect(page.locator('.analysis-panel')).toBeVisible();
		await expect(page.locator('.analysis-panel')).toContainText('Analyzing GraphQL schema for breaking changes...');

		// Verify the finding was rendered
		await expect(page.locator('.analysis-panel')).toContainText('BREAKING:');
		await expect(page.locator('.analysis-panel')).toContainText("Field 'user_id' was removed from User type.");

		// Verify the fix code block was rendered
		await expect(page.locator('.auto-fix-section')).toBeVisible();
		await expect(page.locator('.auto-fix-section .code-block')).toContainText('user_id: String! @deprecated');

		// Apply the fix
		await page.getByRole('button', { name: 'Apply Fix' }).click();

		// Verify the proposed schema editor was updated with the fix
		const proposedEditor = page.locator('textarea.code-editor').nth(1);
		await expect(proposedEditor).toHaveValue(/user_id: String! @deprecated/);
	});
});
