import { test, expect } from '@playwright/test';

test.describe('Public Preview Page', () => {
	const MOCK_TOKEN = '11111111-1111-1111-1111-111111111111';
	const FALLBACK_TOKEN = '22222222-2222-2222-2222-222222222222';
	const EXPIRED_TOKEN = '33333333-3333-3333-3333-333333333333';
	const NOT_FOUND_TOKEN = '99999999-9999-9999-9999-999999999999';
	const BAD_TOKEN = 'not-a-uuid'; // causes 400 or 500 depending on backend error handling

	test.beforeEach(async ({ page }) => {
		// Ignore hydration errors in dev mode that might fail the test
		page.on('pageerror', (err) => {
			if (err.message.includes('hydration')) {
				console.warn('Ignored hydration error:', err.message);
			} else {
				throw err;
			}
		});
	});

	test('should display schema diff when base_schema and head_schema are present', async ({ page }) => {
		await page.goto(`/preview/${MOCK_TOKEN}`);

		// Check headers
		await expect(page.locator('h1.page-title')).toContainText('Schema Preview');
		await expect(page.locator('.page-subtitle')).toContainText('Interactive diff for the proposed changes.');

		// DiffViewer component should render Diff
		await expect(page.locator('.diff-viewer')).toBeVisible();

		// Check for specific diff content
		await expect(page.locator('.diff-body')).toContainText('old schema content');
		await expect(page.locator('.diff-body')).toContainText('new schema content');
	});

	test('should fallback to JSON stringify when base_schema and head_schema are missing', async ({ page }) => {
		await page.goto(`/preview/${FALLBACK_TOKEN}`);

		await expect(page.locator('.diff-viewer')).toBeVisible();

		// Stringified JSON should be shown in diff view
		await expect(page.locator('.diff-body')).toContainText('"some_data": "test value"');
	});

	test('should display expiration error on 410 status', async ({ page }) => {
		await page.goto(`/preview/${EXPIRED_TOKEN}`);

		await expect(page.locator('.error-box')).toBeVisible();
		await expect(page.locator('.error-box p')).toContainText('This preview link has expired.');
	});

	test('should display not found error on 404 status', async ({ page }) => {
		await page.goto(`/preview/${NOT_FOUND_TOKEN}`);

		await expect(page.locator('.error-box')).toBeVisible();
		await expect(page.locator('.error-box p')).toContainText('Preview not found or invalid token.');
	});

	test('should display general error on 500/400 status', async ({ page }) => {
		await page.goto(`/preview/${BAD_TOKEN}`);

		await expect(page.locator('.error-box')).toBeVisible();
		await expect(page.locator('.error-box p')).toContainText('An error occurred while loading the preview.');
	});
});
