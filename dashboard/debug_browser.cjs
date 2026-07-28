const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });
  const page = await context.newPage();

  const logs = [];
  page.on('console', msg => logs.push(`[${msg.type()}] ${msg.text()}`));
  page.on('pageerror', error => logs.push(`[PAGE ERROR] ${error.message}`));
  page.on('response', resp => {
    if (resp.url().includes('/api/v1/')) logs.push(`[NET] ${resp.status()} ${resp.url().replace('http://localhost:5173', '')}`);
  });

  await page.goto('http://localhost:5173/org/mcp-org/graph');

  // Wait for the graph container to appear (not just networkidle)
  try {
    await page.waitForSelector('.graph-container', { timeout: 8000 });
    console.log('✅ graph-container found in DOM');
  } catch {
    console.log('❌ graph-container NOT found after 8s');
  }

  // Wait for SvelteFlow canvas
  try {
    await page.waitForSelector('.svelte-flow', { timeout: 5000 });
    console.log('✅ svelte-flow canvas found');
  } catch {
    console.log('❌ svelte-flow canvas NOT found after 5s');
  }

  // Check empty-state prompt
  const emptyPrompt = await page.$('.empty-state, .graph-empty, [class*="empty"]');
  console.log('Empty state element:', emptyPrompt ? 'found' : 'not found');

  // Find and use search input
  const searchInput = await page.$('input[placeholder="Search repository..."]');
  if (searchInput) {
    console.log('✅ Search input found, typing "front"...');
    await searchInput.fill('front');
    await page.waitForTimeout(800);
    
    const nodes = await page.$$('.svelte-flow__node');
    console.log('Nodes after search:', nodes.length);
    
    for (const n of nodes) {
      const text = await n.innerText().catch(() => '?');
      console.log('  Node text:', text.replace(/\s+/g, ' ').substring(0, 80));
    }
  } else {
    console.log('❌ Search input NOT found');
  }

  console.log('\n--- CONSOLE LOGS ---');
  logs.forEach(l => console.log(l));

  await browser.close();
})();
