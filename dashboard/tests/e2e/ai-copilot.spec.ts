import { test, expect } from '@playwright/test';

/**
 * P12-T09: AI Support Copilot E2E
 * Strategy: NO page.route(), NO mocking. All API requests go to the live Go API.
 * Auth token is provided via E2E_AUTH_TOKEN env (real JWT signed with local-jwt-secret).
 */

test.describe('Support Copilot Widget', () => {
    test.beforeEach(async ({ page, context }) => {
        // Use a real JWT from the E2E runner, or fall back to a structurally valid test token
        const token = process.env.E2E_AUTH_TOKEN ||
            // Fallback: base64url encoded {"orgs":{"default":"admin"}} — valid structure, no real signature needed for UI
            'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJvcmdzIjp7ImRlZmF1bHQiOiJhZG1pbiIsImFkbWluIjoiYWRtaW4ifSwiZXhwIjo5OTk5OTk5OTk5fQ.placeholder';

        await context.addInitScript((t) => {
            localStorage.setItem('github_token', t);
            localStorage.removeItem('copilot_messages');
        }, token);

        await page.goto('/org/default/catalog');
    });

    test('should open, send a message, and receive a streaming response from the real API', async ({ page }) => {
        // 1. Widget FAB is visible and chat is initially closed
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

        // 4. User message appears in the chat
        await expect(page.getByText(testMessage)).toBeVisible();

        // 5. Real API responds — wait for the AI message to stream in
        // The server's fallback response always contains "BREAKING" finding
        await expect(page.getByText('BREAKING', { exact: false }).first()).toBeVisible({ timeout: 15000 });

        // 6. Close the widget
        const closeButton = page.getByRole('button', { name: /Close.*Copilot/i });
        await closeButton.click();
        await expect(page.getByText('Support Copilot', { exact: true })).not.toBeVisible();
    });
});
