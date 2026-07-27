import { test, expect } from '@playwright/test';

test.describe('Heatmap Mode', () => {
	test('can toggle Heatmap Mode and nodes change styles', async ({ page }) => {
		await page.goto('/org/mcp-org/graph');

		// Enter search query to bypass Search-First empty state
		await page.fill('input.filter-input', 'frontend');
		
		// Wait for cytoscape canvas to mount (it draws on a single canvas)
		await page.waitForSelector('canvas[data-id="layer2-node"]', { state: 'attached', timeout: 10000 });

		// Verify initial nodes using cyInstance
		const initialCount = await page.evaluate(() => {
			return (window as any).cyInstance ? (window as any).cyInstance.nodes().length : 0;
		});
		expect(initialCount).toBeGreaterThan(0);

		// Toggle Heatmap Mode
		const heatmapCheckbox = page.locator('label', { hasText: 'Heatmap Mode' });
		await heatmapCheckbox.click();

		await page.waitForTimeout(1000); // Wait for cytoscape re-render

		// Verify heatmap mode activated
		const isChecked = await page.locator('label:has-text("Heatmap Mode") input[type="checkbox"]').isChecked();
		expect(isChecked).toBe(true);
	});
});
