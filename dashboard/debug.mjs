import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto('http://localhost:5175/org/stress-test/graph');
  
  // Wait for network idle and canvas to render
  await page.waitForTimeout(5000);
  
  const nodeCount = await page.locator('.svelte-flow__node').count();
  console.log('NODE COUNT:', nodeCount);
  
  // Save a screenshot to the workspace
  await page.screenshot({ path: '/home/krushna/Project/Substrate/screenshot.png', fullPage: true });
  console.log('Saved screenshot to /home/krushna/Project/Substrate/screenshot.png');
  
  await browser.close();
})();
