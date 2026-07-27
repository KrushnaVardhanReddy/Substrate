import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

const FORGEJO_USER = process.env.FORGEJO_USER || 'adminuser';
const FORGEJO_PASS = process.env.FORGEJO_PASS || 'Admin123!';
const FORGEJO_PORT = process.env.FORGEJO_PORT || '3000';
const FORGEJO_URL = `http://${FORGEJO_USER}:${encodeURIComponent(FORGEJO_PASS)}@localhost:${FORGEJO_PORT}`;
const DASHBOARD_URL = 'http://localhost:5173';
const REPO_ORG = FORGEJO_USER;

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
      // 1. Try to create the repo (ignore if it fails because it already exists)
      await fetch(`http://localhost:${FORGEJO_PORT}/api/v1/user/repos`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Basic ${Buffer.from(`${FORGEJO_USER}:${FORGEJO_PASS}`).toString('base64')}`
        },
        body: JSON.stringify({ name: repoName, private: false })
      });

      // 2. Ensure webhook is registered BEFORE pushing
      const hooksListRes = await fetch(`http://localhost:${FORGEJO_PORT}/api/v1/repos/${FORGEJO_USER}/${repoName}/hooks`, {
        headers: { 'Authorization': `Basic ${Buffer.from(`${FORGEJO_USER}:${FORGEJO_PASS}`).toString('base64')}` }
      });
      const hooks = await hooksListRes.json();
      if (Array.isArray(hooks)) {
        for (const hook of hooks) {
          await fetch(`http://localhost:${FORGEJO_PORT}/api/v1/repos/${FORGEJO_USER}/${repoName}/hooks/${hook.id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Basic ${Buffer.from(`${FORGEJO_USER}:${FORGEJO_PASS}`).toString('base64')}` }
          });
        }
      }

      // Always create exactly one fresh webhook
      const hookRes = await fetch(`http://localhost:${FORGEJO_PORT}/api/v1/repos/${FORGEJO_USER}/${repoName}/hooks`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Basic ${Buffer.from(`${FORGEJO_USER}:${FORGEJO_PASS}`).toString('base64')}`
        },
        body: JSON.stringify({
          type: 'gitea',
          config: { url: 'http://localhost:8090/api/v1/webhook', content_type: 'json' },
          events: ['push', 'pull_request'],
          active: true
        })
      });
      console.log(`Webhook creation response: ${hookRes.status} ${await hookRes.text()}`);

      // 3. Now perform the push, so the webhook fires
      execSync(`git push -u origin main -f`, { cwd: tempDir, stdio: 'pipe' });
    } catch (e: any) {
      throw e;
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
    await page.waitForFunction(() => (window as any).cyInstance !== undefined && (window as any).cyInstance !== null, { timeout: 10000 });
    const count = await page.evaluate(() => {
        return (window as any).cyInstance.nodes().length;
    });
    expect(count).toBeGreaterThanOrEqual(expectedNodesCount);
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
      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "BREAKING"]').length > 0;
      }, { timeout: 10000 });
      
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance.edges('[status = "BREAKING"]').length;
      });
      // The provider affects 4 consumers, so 4 edges should be breaking
      expect(breakingCount).toBe(4);
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
      
      // Wait for pipeline processing
      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'microservices', 5);
      
      // Wait for it to become safe (0 breaking edges)
      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "BREAKING"]').length === 0;
      }, { timeout: 15000 });
      
      // No breaking alerts
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance ? (window as any).cyInstance.edges('[status = "BREAKING"]').length : 0;
      });
      expect(breakingCount).toBe(0);
    });

    test('Yellow Path: Push breaking change with override', async ({ page }) => {
      // Push breaking schema again, but this time with override
      const BREAKING_PROTO = `syntax = "proto3";
package hipstershop;
message CartItem {
    // string product_id = 1; REMOVED
    int32 quantity = 2;
}
message Empty {}
`;
      const OVERRIDE_SUBSTRATE_YAML = `schema_type: protobuf
base_schema: protos/demo.proto
head_schema: protos/demo.proto

consumers:
  - name: frontend
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
    overrides:
      - rule_id: "*"
  - name: checkoutservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
    overrides:
      - rule_id: "*"
  - name: recommendationservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
    overrides:
      - rule_id: "*"
  - name: emailservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
    overrides:
      - rule_id: "*"
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': OVERRIDE_SUBSTRATE_YAML,
        'protos/demo.proto': BREAKING_PROTO
      });

      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'microservices', 5);

      // No breaking alerts, but expecting warning instead
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance ? (window as any).cyInstance.edges('[status = "BREAKING"]').length : 0;
      });
      expect(breakingCount).toBe(0);
      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "WARNING"]').length > 0;
      }, { timeout: 10000 });
      
      const warningCount = await page.evaluate(() => {
          return (window as any).cyInstance.edges('[status = "WARNING"]').length;
      });
      expect(warningCount).toBe(4); // All 4 consumers
    });
  });

  test.describe.serial('Repo: stripe-api (OpenAPI)', () => {
    const REPO_NAME = 'stripe-api';
    
    const BASE_SUBSTRATE_YAML = `schema_type: openapi
base_schema: openapi.yaml
head_schema: openapi.yaml

consumers:
  - name: billing-service
    provider_repo: admin/stripe-api
    schema_type: openapi
    provider_spec_path: openapi.yaml
    provider_branch: main
`;

    const BASE_OPENAPI = `openapi: 3.0.0
info:
  title: Stripe API
  version: 1.0.0
paths:
  /charges:
    post:
      summary: Create a charge
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - amount
              properties:
                amount:
                  type: integer
      responses:
        '200':
          description: OK
`;

    test('Setup & Seeding: Push initial valid schema', async () => {
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'openapi.yaml': BASE_OPENAPI
      });
      await new Promise(r => setTimeout(r, 2000));
    });

    test('Red Path: Push breaking change', async ({ page }) => {
      const BREAKING_OPENAPI = `openapi: 3.0.0
info:
  title: Stripe API
  version: 1.0.0
paths:
  /charges:
    post:
      summary: Create a charge
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - amount
                - currency # BREAKING: added required field
              properties:
                amount:
                  type: integer
                currency:
                  type: string
      responses:
        '200':
          description: OK
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'openapi.yaml': BREAKING_OPENAPI
      });
      
      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'stripe', 2); // 1 provider + 1 consumer
      
      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "BREAKING"]').length > 0;
      }, { timeout: 10000 });
      
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance.edges('[status = "BREAKING"]').length;
      });
      expect(breakingCount).toBe(1);
    });

    test('Green Path: Push safe change', async ({ page }) => {
      const SAFE_OPENAPI = `openapi: 3.0.0
info:
  title: Stripe API
  version: 1.0.0
paths:
  /charges:
    post:
      summary: Create a charge
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - amount
                - currency # Keep required from BREAKING
              properties:
                amount:
                  type: integer
                currency:
                  type: string
                description: # SAFE: added optional field
                  type: string
      responses:
        '200':
          description: OK
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'openapi.yaml': SAFE_OPENAPI
      });
      
      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'stripe', 2);
      
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance ? (window as any).cyInstance.edges('[status = "BREAKING"]').length : 0;
      });
      expect(breakingCount).toBe(0);
    });

    test('Yellow Path: Push breaking change with override', async ({ page }) => {
      const BREAKING_OPENAPI = `openapi: 3.0.0
info:
  title: Stripe API
  version: 1.0.0
paths:
  /charges:
    post:
      summary: Create a charge
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - amount
                - currency # BREAKING: added required field
              properties:
                amount:
                  type: integer
                currency:
                  type: string
      responses:
        '200':
          description: OK
`;
      const OVERRIDE_SUBSTRATE_YAML = `schema_type: openapi
base_schema: openapi.yaml
head_schema: openapi.yaml

consumers:
  - name: billing-service
    provider_repo: admin/stripe-api
    schema_type: openapi
    provider_spec_path: openapi.yaml
    provider_branch: main
    overrides:
      - rule_id: "*"
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': OVERRIDE_SUBSTRATE_YAML,
        'openapi.yaml': BREAKING_OPENAPI
      });

      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'stripe', 2);

      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance ? (window as any).cyInstance.edges('[status = "BREAKING"]').length : 0;
      });
      expect(breakingCount).toBe(0);
      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "WARNING"]').length > 0;
      }, { timeout: 10000 });
      
      const warningCount = await page.evaluate(() => {
          return (window as any).cyInstance.edges('[status = "WARNING"]').length;
      });
      expect(warningCount).toBe(1);
    });
  });

  test.describe.serial('Repo: github-graphql (GraphQL)', () => {
    const REPO_NAME = 'github-graphql';

    const BASE_SUBSTRATE_YAML = `schema_type: graphql
base_schema: schema.graphql
head_schema: schema.graphql

consumers:
  - name: github-action-runner
    provider_repo: admin/github-graphql
    schema_type: graphql
    provider_spec_path: schema.graphql
    provider_branch: main
`;

    const BASE_GRAPHQL = `type Query { user(id: ID!): User }
type User { id: ID!, name: String! }
`;

    test('Setup & Seeding: Push initial valid schema', async () => {
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'schema.graphql': BASE_GRAPHQL
      });
      await new Promise(r => setTimeout(r, 2000));
    });

    test('Red Path: Push breaking change', async ({ page }) => {
      const BREAKING_GRAPHQL = `type Query { user(id: ID!): User }
type User { id: ID! }
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'schema.graphql': BREAKING_GRAPHQL
      });

      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'github', 2); // 1 provider + 1 consumer

      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "BREAKING"]').length > 0;
      }, { timeout: 10000 });
      
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance.edges('[status = "BREAKING"]').length;
      });
      expect(breakingCount).toBe(1);
    });

    test('Green Path: Push safe change', async ({ page }) => {
      const SAFE_GRAPHQL = `type Query { user(id: ID!): User }
type User { id: ID!, name: String!, email: String }
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'schema.graphql': SAFE_GRAPHQL
      });

      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'github', 2);

      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance ? (window as any).cyInstance.edges('[status = "BREAKING"]').length : 0;
      });
      expect(breakingCount).toBe(0);
    });
  });

  test.describe.serial('Repo: ml-models (AI/ML Phase 1f)', () => {
    const REPO_NAME = 'ml-models';
    const BASE_SUBSTRATE_YAML = `schema_type: aiml
base_schema: model.yaml
head_schema: model.yaml

consumers:
  - name: recommendation-engine
    provider_repo: admin/ml-models
    schema_type: aiml
    provider_spec_path: model.yaml
    provider_branch: main
`;
    const BASE_AIML = `service: fraud-detector
ml_model:
  name: "FraudDetector"
  version: "1.0"
  inputs:
    - name: "amount"
      type: "float"
      required: true
  outputs:
    - name: "fraud_probability"
      type: "float"
      required: true
`;
    test('Setup & Seeding: Push initial valid schema', async () => {
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'model.yaml': BASE_AIML
      });
      await new Promise(r => setTimeout(r, 2000));
    });

    test('Red Path: Push breaking change', async ({ page }) => {
      const BREAKING_AIML = `service: fraud-detector
ml_model:
  name: "FraudDetector"
  version: "1.0"
  inputs:
    - name: "amount"
      type: "float"
      required: true
    - name: "new_required_field"
      type: "string"
      required: true
  outputs:
    - name: "fraud_probability"
      type: "float"
      required: true
`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'model.yaml': BREAKING_AIML
      });
      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'ml', 2);
      
      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "BREAKING"]').length > 0;
      }, { timeout: 10000 });
      
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance.edges('[status = "BREAKING"]').length;
      });
      expect(breakingCount).toBe(1);
    });
  });

  test.describe.serial('Repo: salesforce-crm (Enterprise Phase 1g)', () => {
    const REPO_NAME = 'salesforce-crm';
    const BASE_SUBSTRATE_YAML = `schema_type: salesforce
base_schema: Account.object
head_schema: Account.object

consumers:
  - name: sync-worker
    provider_repo: admin/salesforce-crm
    schema_type: salesforce
    provider_spec_path: Account.object
    provider_branch: main
`;
    const BASE_SOAP = `<?xml version="1.0" encoding="UTF-8"?>
<CustomObject xmlns="http://soap.sforce.com/2006/04/metadata">
    <fields>
        <fullName>AnnualRevenue</fullName>
        <type>Currency</type>
        <required>false</required>
    </fields>
</CustomObject>`;
    test('Setup & Seeding: Push initial valid schema', async () => {
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'Account.object': BASE_SOAP
      });
      await new Promise(r => setTimeout(r, 2000));
    });

    test('Red Path: Push breaking change', async ({ page }) => {
      const BREAKING_SOAP = `<?xml version="1.0" encoding="UTF-8"?>
<CustomObject xmlns="http://soap.sforce.com/2006/04/metadata">
</CustomObject>`;
      await pushToForgejo(REPO_NAME, {
        'substrate.yaml': BASE_SUBSTRATE_YAML,
        'Account.object': BREAKING_SOAP
      });
      await page.waitForTimeout(3000);
      await waitForGraphSearch(page, 'salesforce', 2);
      
      await page.waitForFunction(() => {
          const cy = (window as any).cyInstance;
          if (!cy) return false;
          return cy.edges('[status = "BREAKING"]').length > 0;
      }, { timeout: 10000 });
      
      const breakingCount = await page.evaluate(() => {
          return (window as any).cyInstance.edges('[status = "BREAKING"]').length;
      });
      expect(breakingCount).toBe(1);
    });
  });

});
