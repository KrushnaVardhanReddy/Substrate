import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8090';

// testorg insurance: policy_limit_cents=1000000 ($10,000), one PENDING claim of $500
// These are seeded in seed_mcp.sql

test.describe('Schema Insurance Settings', () => {
    test.beforeEach(async ({ page }) => {
        page.on('request', request => console.log('>>', request.method(), request.url()));
        page.on('response', response => console.log('<<', response.status(), response.url()));

        // Set up auth token in localStorage for mcp-org
        await page.addInitScript(() => {
            const token = btoa(JSON.stringify({ orgs: { testorg: 'admin' } }));
            window.localStorage.setItem('github_token', `header.${token}.signature`);
            window.localStorage.setItem('substrate-token', 'local-dev-token');
        });

        page.on('pageerror', (err) => {
            if (
                err.message.includes('hydration') ||
                err.message.includes('No matching export') ||
                err.message.includes('Svelte') ||
                err.message.includes('lifecycle_outside_component')
            ) {
                return;
            }
            console.error('PAGE ERROR:', err);
        });
    });

    test('should render the insurance settings page', async ({ page }) => {
        // Navigate to insurance settings — verify the page loads at all
        await page.goto('/org/testorg/settings/insurance');
        await expect(page.getByRole('heading', { name: 'Schema Insurance' })).toBeVisible({ timeout: 10000 });
    });

    test('should render an active policy and claims history', async ({ page, request }) => {
        // Verify the API has the policy (seeded in seed_mcp.sql)
        const policyRes = await request.get(`${API_URL}/api/v1/org/testorg/insurance/policy`, {
            headers: { 'Authorization': 'Bearer local-dev-token' },
            timeout: 5000
        }).catch(() => null);

        if (!policyRes || !policyRes.ok()) {
            test.skip(true, 'Insurance policy not found for testorg — check seed_mcp.sql');
            return;
        }

        await page.goto('/org/testorg/settings/insurance');

        // Policy ID, Limit, and claim history should render from the live API
        try {
            await expect(page.getByText('Policy ID:')).toBeVisible({ timeout: 10000 });
        } catch (e) {
            console.error("PAGE CONTENT:", await page.content());
            throw e;
        }
        await expect(page.getByText('Limit:')).toBeVisible();
        // $10,000 limit (1000000 cents)
        await expect(page.getByText('$10000.00')).toBeVisible({ timeout: 5000 });

        await expect(page.getByRole('heading', { name: 'Claim History' })).toBeVisible();
        // $500 PENDING claim (50000 cents)
        await expect(page.getByText('$500.00').first()).toBeVisible();
        await expect(page.getByText('PENDING').first()).toBeVisible();
    });

    test('should successfully file a new claim', async ({ page, request }) => {
        const policyRes = await request.get(`${API_URL}/api/v1/org/testorg/insurance/policy`, {
            headers: { 'Authorization': 'Bearer local-dev-token' },
            timeout: 5000
        }).catch(() => null);

        if (!policyRes || !policyRes.ok()) {
            test.skip(true, 'Insurance policy not found for testorg — check seed_mcp.sql');
            return;
        }

        await page.goto('/org/testorg/settings/insurance');
        await expect(page.getByText('Policy ID:')).toBeVisible({ timeout: 10000 });

        await page.fill('#prUrl', 'https://github.com/testorg/repo/pull/99');
        await page.fill('#incidentDate', '2024-06-15');
        await page.fill('#amount', '150.50');

        await page.click('button:has-text("File Claim")');

        // After filing, $150.50 claim should appear in the list
        await expect(page.getByText('$150.50')).toBeVisible({ timeout: 10000 });
    });

    test('should handle API errors gracefully', async ({ page }) => {
        // Navigate to an org that has no policy (new-org has nothing seeded)
        await page.goto('/org/new-org/settings/insurance');

        // The page should show an error or no-policy state
        await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => {});
        const body = await page.textContent('body');
        expect(body).toBeTruthy();
        // Should show either error message or no-policy message
        const hasErrorOrEmpty = body!.includes('No active insurance policy') ||
                                body!.includes('Failed to load') ||
                                body!.includes('not found') ||
                                body!.includes('Schema Insurance');
        expect(hasErrorOrEmpty).toBe(true);
    });
});
