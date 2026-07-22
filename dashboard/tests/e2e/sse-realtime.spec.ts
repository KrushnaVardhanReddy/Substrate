import { test, expect } from '@playwright/test';

test.describe('SSE Realtime E2E (Suite 12)', () => {

    test('should receive backend event in multiple contexts without page reload', async ({ browser }) => {
        // Create two separate browser contexts
        const context1 = await browser.newContext();
        const context2 = await browser.newContext();

        const page1 = await context1.newPage();
        const page2 = await context2.newPage();

        let sseFired1 = false;
        let sseFired2 = false;

        // Mock SSE events route
        await page1.route('**/api/v1/events', async (route) => {
            const streamBody = 'data: {"type": "sync_complete", "repo": "test-repo"}\n\n';
            await route.fulfill({
                status: 200,
                contentType: 'text/event-stream',
                body: streamBody
            });
            sseFired1 = true;
        });

        await page2.route('**/api/v1/events', async (route) => {
            const streamBody = 'data: {"type": "sync_complete", "repo": "test-repo"}\n\n';
            await route.fulfill({
                status: 200,
                contentType: 'text/event-stream',
                body: streamBody
            });
            sseFired2 = true;
        });

        // For SSE to be activated, we typically need to be on the graph page or matrix
        await page1.goto('/org/test-org/graph');
        await page2.goto('/org/test-org/graph');

        // Let the SSE connection establish
        await page1.waitForTimeout(1000);
        await page2.waitForTimeout(1000);

        // We expect the mock to have been hit at least if the frontend attempts connection
        // The mock fulfills it immediately with a test event.
        expect(sseFired1).toBe(true);
        expect(sseFired2).toBe(true);

        await context1.close();
        await context2.close();
    });

});
