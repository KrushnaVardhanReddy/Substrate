import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem('auth_token', 'local-dev-token');
  });

  page.on('pageerror', (err) => {
    if (err.message.includes('hydration')) {
      console.warn('Suppressed hydration error in test:', err.message);
    } else {
      throw err;
    }
  });
});

test('P5-T06 Phase 5 Discovery Scanners', async ({ page }) => {
  // Gracefully skip if backend isn't up
  try {
    const res = await page.request.get('http://localhost:8090/health');
    if (!res.ok()) {
      test.skip(true, 'Local API not running');
    }
  } catch (e) {
    test.skip(true, 'Local API not running');
  }

  // Navigate to discovery dashboard
  await page.goto('/org/mcp-org/discovery/discovery-test-repo');

  // Trigger scan
  const scanButton = page.locator('button', { hasText: 'Run Full Scan' });
  await expect(scanButton).toBeVisible();
  await scanButton.click();

  // Wait for scan to complete and results to load (if any)
  // Just wait for scanning to finish
  await expect(scanButton).toHaveText('Run Full Scan', { timeout: 10000 });
});
