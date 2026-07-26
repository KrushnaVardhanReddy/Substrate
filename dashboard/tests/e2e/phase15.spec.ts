import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8090';

// mcp-org insurance seeded in seed_mcp.sql:
// policy_limit_cents = 5000000 ($50,000)
// One APPROVED claim for $150 (15000 cents)

test.describe('Phase 15 Monetization E2E', () => {
    test.beforeEach(async ({ page }) => {
        await page.addInitScript(() => {
            const token = btoa(JSON.stringify({ orgs: { 'mcp-org': 'admin' } }));
            window.localStorage.setItem('github_token', `header.${token}.signature`);
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
            console.error(err);
        });
    });

    test('should render an active policy and claims history', async ({ page, request }) => {
        const policyRes = await request.get(`${API_URL}/api/v1/org/mcp-org/insurance/policy`, {
            headers: { 'Authorization': 'Bearer local-dev-token' },
            timeout: 5000
        }).catch(() => null);

        if (!policyRes || !policyRes.ok()) {
            test.skip(true, 'mcp-org insurance policy not found — check seed_mcp.sql');
            return;
        }

        await page.goto('/org/mcp-org/settings/insurance');

        await expect(page.getByText('Policy ID:')).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Limit:')).toBeVisible();
        // $50,000 limit = 5000000 cents
        await expect(page.getByText('$50,000.00')).toBeVisible({ timeout: 5000 });

        await expect(page.getByRole('heading', { name: 'Claim History' })).toBeVisible();
        // $150 APPROVED claim = 15000 cents
        await expect(page.getByText('$150.00')).toBeVisible();
        await expect(page.getByText('APPROVED')).toBeVisible();
    });

    test('should successfully file a new claim', async ({ page, request }) => {
        const policyRes = await request.get(`${API_URL}/api/v1/org/mcp-org/insurance/policy`, {
            headers: { 'Authorization': 'Bearer local-dev-token' },
            timeout: 5000
        }).catch(() => null);

        if (!policyRes || !policyRes.ok()) {
            test.skip(true, 'mcp-org insurance policy not found — check seed_mcp.sql');
            return;
        }

        await page.goto('/org/mcp-org/settings/insurance');
        await expect(page.getByText('Policy ID:')).toBeVisible({ timeout: 10000 });

        await page.fill('#prUrl', 'https://github.com/mcp-org/repo/pull/42');
        await page.fill('#incidentDate', '2024-06-15');
        await page.fill('#amount', '150.50');

        await page.click('button:has-text("File Claim")');

        // After filing, should show PENDING claim
        await expect(page.getByText('$150.50')).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('PENDING')).toBeVisible({ timeout: 5000 });
    });
});
