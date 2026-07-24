import { test, expect } from '@playwright/test';

test.describe('Schema Insurance Settings', () => {
	test.beforeEach(async ({ page }) => {
		// Mock layout requests that might cause unhandled promise rejections / timeouts
		await page.route('**/api/v1/repos/*', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{ id: '1', name: 'core-auth', full_name: 'core/auth' }
				])
			});
		});
		await page.route('**/api/v1/user', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ name: 'Test User' })
			});
		});

		// Setup local storage to bypass token parsing errors in layout
		await page.addInitScript(() => {
			const token = btoa(JSON.stringify({ orgs: { testorg: 'admin' } }));
			window.localStorage.setItem('github_token', `header.${token}.signature`);
		});

		// Ignore hydration errors in dev mode
		page.on('pageerror', (err) => {
			if (
				err.message.includes('hydration') ||
				err.message.includes('No matching export') ||
				err.message.includes('Svelte') ||
				err.message.includes('lifecycle_outside_component')
			) {
				return;
			}
			console.error(err);
		});

		// In SvelteKit, route requests also attempt to fetch graph data for sidebar layout or so
		await page.route('**/api/v1/graph/*', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' });
		});
	});

	test('should render without an active policy', async ({ page }) => {
		await page.route('**/api/v1/org/*/insurance/policy', async (route) => {
			await route.fulfill({ status: 404 });
		});
		await page.route('**/api/v1/org/*/insurance/claims', async (route) => {
			await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' });
		});

		await page.goto('/org/testorg/settings/insurance');

		await expect(page.getByRole('heading', { name: 'Schema Insurance' })).toBeVisible();
		await expect(page.getByText('No active insurance policy found for this organization.')).toBeVisible();
		await expect(page.getByText('You need an active policy to file a claim.')).toBeVisible();
		await expect(page.getByText('No claims filed yet.')).toBeVisible();
	});

	test('should render an active policy and claims history', async ({ page }) => {
		await page.route('**/api/v1/org/*/insurance/policy', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					id: 'POL-123',
					policy_limit_cents: 1000000
				})
			});
		});

		await page.route('**/api/v1/org/*/insurance/claims', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{
						id: 'CLM-1',
						incident_date: new Date('2023-10-01').toISOString(),
						github_pr_url: 'https://github.com/foo/bar/pull/1',
						amount_cents: 50000,
						status: 'PENDING'
					}
				])
			});
		});

		await page.goto('/org/testorg/settings/insurance');

		await expect(page.getByText('Policy ID:')).toBeVisible();
		await expect(page.getByText('POL-123')).toBeVisible();
		await expect(page.getByText('Limit:')).toBeVisible();
		await expect(page.getByText('$10000.00')).toBeVisible();

		await expect(page.getByRole('heading', { name: 'Claim History' })).toBeVisible();
		await expect(page.getByText('$500.00')).toBeVisible();
		await expect(page.getByText('PENDING')).toBeVisible();
	});

	test('should successfully file a new claim', async ({ page }) => {
		await page.route('**/api/v1/org/*/insurance/policy', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					id: 'POL-123',
					policy_limit_cents: 1000000
				})
			});
		});

		await page.route('**/api/v1/org/*/insurance/claims', async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: '[]'
				});
			} else if (route.request().method() === 'POST') {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						id: 'CLM-2',
						incident_date: new Date('2023-10-15').toISOString(),
						github_pr_url: 'https://github.com/foo/bar/pull/2',
						amount_cents: 15050,
						status: 'APPROVED'
					})
				});
			} else {
				await route.continue();
			}
		});

		await page.goto('/org/testorg/settings/insurance');

		await expect(page.getByText('Policy ID:')).toBeVisible();

		await page.fill('#prUrl', 'https://github.com/foo/bar/pull/2');
		await page.fill('#incidentDate', '2023-10-15');
		await page.fill('#amount', '150.50');

		await page.click('button:has-text("File Claim")');

		await expect(page.getByText('$150.50')).toBeVisible();
		await expect(page.getByText('APPROVED')).toBeVisible();
	});

	test('should handle API errors gracefully', async ({ page }) => {
		await page.route('**/api/v1/org/*/insurance/policy', async (route) => {
			await route.fulfill({ status: 500 });
		});

		await page.goto('/org/testorg/settings/insurance');

		await expect(page.getByText('Failed to load policy')).toBeVisible();
	});
});
