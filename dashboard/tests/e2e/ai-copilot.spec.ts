import { test, expect } from '@playwright/test';

test.describe('Support Copilot Widget', () => {
	test.beforeEach(async ({ page, context }) => {
		// Set mock token to bypass auth
		await context.addInitScript(() => {
			localStorage.setItem(
				'github_token',
				'mock.eyJvcmdzIjp7ImRlZmF1bHQiOiJhZG1pbiJ9fQ==.mock'
			);
			// clear messages
			localStorage.removeItem('copilot_messages');
		});

		// Mock the AI API endpoint with Server-Sent Events stream


		await page.goto('/org/default/catalog');
	});

	test('should open, send a message, and receive a streaming response', async ({ page }) => {
		// 1. Check widget is visible and closed initially
		const fab = page.getByRole('button', { name: 'Open Support Copilot' });
		await expect(fab).toBeVisible();
		await expect(page.getByText('Support Copilot', { exact: true })).not.toBeVisible();

		// 2. Open widget
		await fab.click();
		await expect(page.getByText('Support Copilot', { exact: true })).toBeVisible();

		const textarea = page.getByPlaceholder('Ask a question...');
		await expect(textarea).toBeVisible();

		// 3. Type a message and submit
		const testMessage = 'How do I add a consumer?';
		await textarea.fill(testMessage);

		const sendButton = page.getByRole('button', { name: 'Send message' });
		await sendButton.click();

		// 4. Verify user message appears
		await expect(page.getByText(testMessage)).toBeVisible();

		// 5. Verify the AI response streamed in
		await expect(page.getByText('Let me check that for you.')).toBeVisible();
		await expect(page.getByText('**SAFE**: Looking good.')).toBeVisible();
		await expect(page.getByText('consumer:\n  name: Test')).toBeVisible();

		// 6. Close the widget
		const closeButton = page.getByRole('button', { name: 'Close Copilot' });
		await closeButton.click();
		await expect(page.getByText('Support Copilot', { exact: true })).not.toBeVisible();
	});
});
