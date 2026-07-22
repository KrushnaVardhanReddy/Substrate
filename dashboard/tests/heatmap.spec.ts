import { test, expect } from '@playwright/test';

test.describe('Heatmap Mode', () => {
	test('can toggle Heatmap Mode and nodes change styles', async ({ page }) => {
		// Mock graph data
		await page.route('**/api/v1/graph/*', async route => {
			const json = [
				{ provider: 'service-A', consumer: 'service-B', status: 'SAFE' },
				{ provider: 'service-C', consumer: 'service-D', status: 'BREAKING' }
			];
			await route.fulfill({ json });
		});

		await page.goto('/org/myorg/graph');

		// Wait for canvas to load
		await expect(page.locator('.main-canvas')).toBeVisible();

		// Wait for the mock fetch to complete
		await page.waitForTimeout(1000);

		// Enter search query to bypass Search-First empty state
		await page.fill('input.filter-input', 'service');
		await page.waitForSelector('.svelte-flow', { state: 'attached' });

		// Get all service node cards
		const serviceNodes = page.locator('.service-node-card');
		await expect(serviceNodes.first()).toBeVisible();

		// Ensure initial state doesn't have the explicit style overriding the default (or has the default if inline style was added)
		// With our implementation, default background is `#1E222C` which is applied via inline style too now! Let's check for it.
		// Svelte reactive styles might apply 'background-color: rgb(30, 34, 44);'
		await expect(serviceNodes.first()).toHaveCSS('background-color', 'rgb(30, 34, 44)');

		// Check the checkbox for Heatmap Mode
		const heatmapCheckbox = page.locator('label:has-text("Heatmap Mode") input[type="checkbox"]');
		await expect(heatmapCheckbox).toBeVisible();
		await heatmapCheckbox.check();

		// Give reactivity a moment
		await page.waitForTimeout(500);

		// After checking, the background color of at least one node should change to one of our threshold colors.
		// The score is random, so it might be green, yellow, or red.
		// Green: rgb(0, 191, 165)
		// Safe/Yellow: rgb(255, 171, 64)
		// Danger/Red: rgb(255, 82, 82)
		// Let's assert that the color is NOT the default dark gray.
		const bgAfter = await serviceNodes.first().evaluate((el) => window.getComputedStyle(el).backgroundColor);
		expect(bgAfter).not.toBe('rgb(30, 34, 44)');

        // Assert it's one of the expected valid RGB colors. We can parse or just use string matches since we know them.
        const validColors = ['rgb(0, 191, 165)', 'rgb(255, 171, 64)', 'rgb(255, 82, 82)'];
        expect(validColors.includes(bgAfter)).toBeTruthy();
	});
});
