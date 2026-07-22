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
		await page.route('/api/v1/ai/analyze', async (route) => {
			// Instead of a direct standard json response, we need to mock a stream response.
			// Playwright route.fulfill allows sending body strings. Since it's SSE, we just send a formatted string.
			const streamData = [
				'data: {"type":"thinking","content":"Let me check that for you."}\n\n',
				'data: {"type":"finding","severity":"SAFE","content":"Looking good."}\n\n',
				'data: {"type":"fix","language":"yaml","code":"consumer:\\n  name: Test"}\n\n',
				'data: [DONE]\n\n'
			].join('');

			await route.fulfill({
				status: 200,
				headers: {
					'Content-Type': 'text/event-stream',
					'Cache-Control': 'no-cache',
					'Connection': 'keep-alive'
				},
				body: streamData
			});
		});

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
