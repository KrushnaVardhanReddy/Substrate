import { test, expect } from '@playwright/test';

test.describe('Public Preview Page', () => {
	const MOCK_TOKEN = 'mock-token-123';

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
		await page.route(`**/api/v1/preview/${MOCK_TOKEN}`, async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					base_schema: 'old schema content\nline 2',
					head_schema: 'new schema content\nline 2'
				})
			});
		});

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
		const mockReport = {
			some_data: 'test value',
			nested: {
				key: 'value'
			}
		};

		await page.route(`**/api/v1/preview/${MOCK_TOKEN}`, async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockReport)
			});
		});

		await page.goto(`/preview/${MOCK_TOKEN}`);

		await expect(page.locator('.diff-viewer')).toBeVisible();

		// Stringified JSON should be shown in diff view
		await expect(page.locator('.diff-body')).toContainText('"some_data": "test value"');
	});

	test('should display expiration error on 410 status', async ({ page }) => {
		await page.route(`**/api/v1/preview/${MOCK_TOKEN}`, async (route) => {
			await route.fulfill({
				status: 410,
				contentType: 'application/json',
				body: JSON.stringify({ message: 'gone' })
			});
		});

		await page.goto(`/preview/${MOCK_TOKEN}`);

		await expect(page.locator('.error-box')).toBeVisible();
		await expect(page.locator('.error-box p')).toContainText('This preview link has expired.');
	});

	test('should display not found error on 404 status', async ({ page }) => {
		await page.route(`**/api/v1/preview/${MOCK_TOKEN}`, async (route) => {
			await route.fulfill({
				status: 404,
				contentType: 'application/json',
				body: JSON.stringify({ message: 'not found' })
			});
		});

		await page.goto(`/preview/${MOCK_TOKEN}`);

		await expect(page.locator('.error-box')).toBeVisible();
		await expect(page.locator('.error-box p')).toContainText('Preview not found or invalid token.');
	});

	test('should display general error on 500 status', async ({ page }) => {
		await page.route(`**/api/v1/preview/${MOCK_TOKEN}`, async (route) => {
			await route.fulfill({
				status: 500,
				contentType: 'application/json',
				body: JSON.stringify({ message: 'internal error' })
			});
		});

		await page.goto(`/preview/${MOCK_TOKEN}`);

		await expect(page.locator('.error-box')).toBeVisible();
		await expect(page.locator('.error-box p')).toContainText('An error occurred while loading the preview.');
	});
});
