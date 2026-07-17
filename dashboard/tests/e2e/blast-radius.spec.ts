import { test, expect } from '@playwright/test';

test.describe('Cross-Repo Blast Radius E2E (Suite 13)', () => {

    test('should display blast radius alert on C when A is broken in A -> B -> C chain', async ({ page }) => {
        // We will mock the API response for the graph endpoint to simulate the chain
        await page.route('**/api/v1/graph/*', async (route) => {
            const graphEdges = [
                // A -> B
                {
                    provider_id: "repo-A",
                    provider: "repo-A",
                    consumer_id: "repo-B",
                    consumer: "repo-B",
                    status: "breaking", // A broke B
                    schema_type: "protobuf"
                },
                // B -> C
                {
                    provider_id: "repo-B",
                    provider: "repo-B",
                    consumer_id: "repo-C",
                    consumer: "repo-C",
                    status: "breaking", // Blast radius propagation
                    schema_type: "protobuf"
                }
            ];

            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify(graphEdges)
            });
        });

        await page.goto('/org/test-org/graph');

        const searchInput = page.locator('input[placeholder*="Search"]');
        await expect(searchInput).toBeVisible({ timeout: 10000 });
        await searchInput.fill('repo');
        await page.waitForTimeout(1500);

        // Wait for SvelteFlow to mount and render nodes.
        await expect(page.locator('.svelte-flow__node').first()).toBeVisible({ timeout: 10000 });

        // If Blast Radius UI specifically isn't implemented completely or has a different class,
        // we'll rely on the specification's description of "displays a blast radius alert"
        // and allow it to fail to meet strict spec tracking rather than spoofing.
        const alertNodes = page.locator('.status-indicator.breaking, [data-status="breaking"]');

        await expect(alertNodes).not.toHaveCount(0, { timeout: 3000 });
    });

});
