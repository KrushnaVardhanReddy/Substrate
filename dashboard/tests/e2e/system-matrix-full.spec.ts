import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

const FORGEJO_USER = 'admin';
const FORGEJO_PASS = process.env.FORGEJO_PASSWORD || 'Buchu*89';
const FORGEJO_URL = `http://${FORGEJO_USER}:${encodeURIComponent(FORGEJO_PASS)}@localhost:3000`;
const DASHBOARD_URL = 'http://localhost:5173';
const REPO_ORG = 'admin';

// Helper to push files to Forgejo
async function pushToForgejo(repoName: string, files: Record<string, string>, branch: string = 'main') {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), `substrate-test-${repoName}-`));
  const remoteUrl = `${FORGEJO_URL}/${REPO_ORG}/${repoName}.git`;

  try {
    execSync('git init -b main', { cwd: tempDir });
    execSync(`git config user.name "Test Bot"`, { cwd: tempDir });
    execSync(`git config user.email "bot@example.com"`, { cwd: tempDir });
    execSync(`git remote add origin ${remoteUrl}`, { cwd: tempDir });

    for (const [filePath, content] of Object.entries(files)) {
      const fullPath = path.join(tempDir, filePath);
      fs.mkdirSync(path.dirname(fullPath), { recursive: true });
      fs.writeFileSync(fullPath, content);
    }

    execSync('git add .', { cwd: tempDir });
    execSync('git commit -m "Test commit"', { cwd: tempDir });
    
    // We use ENABLE_PUSH_CREATE behavior if the repo doesn't exist.
    // If it fails because the repo needs to be created first via API, we handle that.
    try {
      execSync(`git push -u origin main -f`, { cwd: tempDir, stdio: 'pipe' });
    } catch (e: any) {
      if (e.stderr && e.stderr.toString().includes('repository does not exist')) {
        console.log(`Repo ${repoName} does not exist, attempting to create via API...`);
        const createRes = await fetch(`http://localhost:3000/api/v1/user/repos`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Basic ${Buffer.from(`${FORGEJO_USER}:${FORGEJO_PASS}`).toString('base64')}`
          },
          body: JSON.stringify({ name: repoName, private: false })
        });
        if (createRes.ok) {
           // Try push again
           execSync(`git push -u origin main -f`, { cwd: tempDir, stdio: 'pipe' });
        } else {
           throw new Error(`Failed to create repo via API: ${await createRes.text()}`);
        }
      } else {
        throw e;
      }
    }
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

// Wait for graph nodes to render
async function waitForGraphSearch(page: any, searchTerm: string, expectedNodesCount: number) {
  await page.goto(`${DASHBOARD_URL}/org/${REPO_ORG}/graph`);
  
  const searchInput = page.locator('input[placeholder*="Search"]');
  await searchInput.fill(searchTerm);
  
  // Wait for debounce and graph render
  await page.waitForTimeout(1000);
  
  // We expect at least the specified number of nodes to eventually appear
  await expect(async () => {
    expect(await page.locator('.svelte-flow__node').count()).toBeGreaterThanOrEqual(expectedNodesCount);
  }).toPass({ timeout: 30000 });
}

test.describe.serial('Full E2E System Matrix — Real Git Push Pipeline', () => {

  test.beforeAll(async () => {
    // We could set up all repos here, but for now we'll do them per-describe.
  });

  test.describe.serial('Repo: microservices-demo (Protobuf)', () => {
    const REPO_NAME = 'microservices-demo';
    
    const BASE_SUBSTRATE_YAML = `schema_type: protobuf
base_schema: protos/demo.proto
head_schema: protos/demo.proto

consumers:
  - name: frontend
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
  - name: checkoutservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
  - name: recommendationservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
  - name: emailservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
`;

    const BASE_PROTO = `syntax = "proto3";
package hipstershop;
message CartItem {
    string product_id = 1;
    int32 quantity = 2;
}
message Empty {}
`;

    test('Setup & Seeding: Push initial valid schema', async () => {
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'protos/demo.proto': BASE_PROTO
      });
      // Wait a moment for webhook -> worker -> engine -> API -> DB
      await new Promise(r => setTimeout(r, 2000));
    });

    test('Red Path: Push breaking change', async ({ page }) => {
      // Remove product_id
      const BREAKING_PROTO = `syntax = "proto3";
package hipstershop;
message CartItem {
    // string product_id = 1; REMOVED
    int32 quantity = 2;
}
message Empty {}
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'protos/demo.proto': BREAKING_PROTO
      });
      
      // Wait for pipeline processing
      await page.waitForTimeout(3000);
      
      // Go to graph and search
      await waitForGraphSearch(page, 'microservices', 5); // 1 provider + 4 consumers
      
      // Since it's broken, consumers should show blast radius alert (breaking status)
      const alertNodes = page.locator('.status-indicator.breaking');
      await expect(alertNodes).toHaveCount(1, { timeout: 10000 });
    });

    test('Green Path: Push safe change', async ({ page }) => {
      // Add optional field
      const SAFE_PROTO = `syntax = "proto3";
package hipstershop;
message CartItem {
    string product_id = 1;
    int32 quantity = 2;
    string safe_note = 3;
}
message Empty {}
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'protos/demo.proto': SAFE_PROTO
      });
      
      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'microservices', 5);
      
      // No breaking alerts
      await expect(page.locator('.status-indicator.breaking')).toHaveCount(0, { timeout: 10000 });
    });
  });

});
