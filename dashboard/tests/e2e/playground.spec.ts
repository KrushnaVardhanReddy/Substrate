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

test.describe('AI Playground Analysis Workflow', () => {
  const ANALYZE_API = '**/api/v1/ai/analyze';
  const VALID_SCHEMA = `
openapi: 3.0.0
info:
  title: Test API
paths:
  /users:
    get:
      responses:
        200:
          content: {}
  `.trim();

  test('should stream AI analysis and display findings', async ({ page }) => {
    // a. Mock the AI analyze endpoint for SSE
    await page.route(ANALYZE_API, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'text/event-stream',
        headers: {
          'Connection': 'keep-alive',
          'Cache-Control': 'no-cache',
        },
        body: 'data: {"type":"finding","severity":"BREAKING","content":"Field userId removed"}\n\ndata: {"type":"fix","code":"openapi: 3.0.0\\ninfo:\\n  title: Fixed API"}\n\ndata: {"type":"done"}\n\n'
      });
    });

    // b. Navigate to playground
    await page.goto('/playground');

    // c. Assert heading
    await expect(page.locator('h1.page-title')).toContainText(/Playground|AI|Studio/i);

    // d. Locate and fill editor
    const editor = page.locator('textarea, [contenteditable="true"]').first();
    await editor.fill(VALID_SCHEMA);

    // e. Click Analyze
    await page.getByRole('button', { name: /Analyze with Substrate AI/i }).click();

    // f. Wait for findings panel
    const findings = page.locator('.analysis-panel');
    await expect(findings).toBeVisible({ timeout: 10000 });

    // g. Assert finding text
    await expect(findings).toContainText(/BREAKING|userId/);
  });

  test('should apply auto-fix to editor content', async ({ page }) => {
    // Setup console error listener
    const consoleErrors: string[] = [];
    page.on('console', msg => { 
      if (msg.type() === 'error') consoleErrors.push(msg.text()); 
    });

    // Repeat Test 1 setup
    await page.route(ANALYZE_API, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'text/event-stream',
        body: 'data: {"type":"finding","severity":"BREAKING","content":"Field userId removed"}\n\ndata: {"type":"fix","code":"Fixed Content"}\n\ndata: {"type":"done"}\n\n'
      });
    });

    await page.goto('/playground');
    const editor = page.locator('textarea, [contenteditable="true"]').first();
    await editor.fill(VALID_SCHEMA);
    await page.getByRole('button', { name: /Analyze with Substrate AI/i }).click();

    // b. Wait for and click Apply Auto-Fix
    const autoFixBtn = page.getByRole('button', { name: /apply fix/i });
    await expect(autoFixBtn).toBeVisible();
    await autoFixBtn.click();

    // c. Small wait for state update
    await page.waitForTimeout(500);

    // d. Read content from the Proposed Schema editor (the 2nd editor)
    const proposedEditor = page.locator('textarea.code-editor').nth(1);
    await expect(proposedEditor).toHaveValue(/Fixed Content/);

    // f. Assert no Svelte/State errors
    const criticalErrors = consoleErrors.filter(e => e.includes('$state') || e.includes('properties of undefined'));
    expect(criticalErrors).toHaveLength(0);
  });

  test('should handle malformed input gracefully', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('console', msg => { 
      if (msg.type() === 'error') consoleErrors.push(msg.text()); 
    });

    // a. Mock 400 error
    await page.route(ANALYZE_API, async (route) => {
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Invalid schema format' })
      });
    });

    // b. Navigate
    await page.goto('/playground');

    // c. Fill with malformed data
    const editor = page.locator('textarea.code-editor').first();
    await editor.fill("{{ INVALID_YAML_\x00 }}");

    // d. Click Analyze
    await page.getByRole('button', { name: /Analyze with Substrate AI/i }).click();

    // e. Assert error was logged and gracefully caught
    await page.waitForTimeout(500); // Give time for the fetch to resolve
    expect(consoleErrors.some(e => e.includes('Invalid schema format'))).toBeTruthy();
    
    // f. Assert main container still visible (no crash)
    await expect(page.locator('main, .playground-container, .page-content')).toBeVisible();
  });
});
