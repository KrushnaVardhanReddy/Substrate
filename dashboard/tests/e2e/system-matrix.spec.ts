import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

test.setTimeout(120000);

function pushToGit(repoName: string, files: Record<string, string>) {
    const dir = `/tmp/${repoName}-test`;
    if (fs.existsSync(dir)) {
        fs.rmSync(dir, { recursive: true, force: true });
    }
    fs.mkdirSync(dir, { recursive: true });

    for (const [filename, content] of Object.entries(files)) {
        fs.writeFileSync(path.join(dir, filename), content);
    }

    execSync('git init', { cwd: dir, stdio: 'ignore' });
    try { execSync('git checkout -b main', { cwd: dir, stdio: 'ignore' }); } catch (e) {}
    execSync('git add .', { cwd: dir, stdio: 'ignore' });
    execSync('git config user.name "Test"', { cwd: dir, stdio: 'ignore' });
    execSync('git config user.email "test@example.com"', { cwd: dir, stdio: 'ignore' });
    execSync('git commit -m "update"', { cwd: dir, stdio: 'ignore' });
    execSync(`git remote add origin http://admin:admin@localhost:3000/admin/${repoName}.git`, { cwd: dir, stdio: 'ignore' });
    execSync('git push -u origin main -f', { cwd: dir, stdio: 'ignore' });
}

const testMatrix = [
    {
        repoName: 'microservices-demo',
        schemaFile: 'schema.proto',
        baseContent: 'syntax = "proto3";\npackage microservices;\nmessage CartItem {\n  string product_id = 1;\n}',
        redContent: 'syntax = "proto3";\npackage microservices;\nmessage CartItem {\n}',
        greenContent: 'syntax = "proto3";\npackage microservices;\nmessage CartItem {\n  string product_id = 1;\n  string notes = 3;\n}'
    },
    {
        repoName: 'stripe-api',
        schemaFile: 'openapi.yaml',
        baseContent: 'openapi: 3.0.0\ninfo:\n  title: Stripe API\n  version: 1.0.0\npaths:\n  /v1/charges:\n    get:\n      responses:\n        "200":\n          description: OK',
        redContent: 'openapi: 3.0.0\ninfo:\n  title: Stripe API\n  version: 1.0.0\npaths: {}',
        greenContent: 'openapi: 3.0.0\ninfo:\n  title: Stripe API\n  version: 1.0.0\npaths:\n  /v1/charges:\n    get:\n      responses:\n        "200":\n          description: OK\n  /v2/beta/charges:\n    get:\n      responses:\n        "200":\n          description: OK'
    },
    {
        repoName: 'realworld-api',
        schemaFile: 'openapi.yaml',
        baseContent: 'openapi: 3.0.0\ninfo:\n  title: RealWorld API\n  version: 1.0.0\npaths:\n  /api/articles:\n    get:\n      responses:\n        "200":\n          description: OK',
        redContent: 'openapi: 3.0.0\ninfo:\n  title: RealWorld API\n  version: 1.0.0\npaths: {}',
        greenContent: 'openapi: 3.0.0\ninfo:\n  title: RealWorld API\n  version: 1.0.0\npaths:\n  /api/articles:\n    get:\n      responses:\n        "200":\n          description: OK\n  /api/tags:\n    get:\n      responses:\n        "200":\n          description: OK'
    },
    {
        repoName: 'openai-api',
        schemaFile: 'openapi.yaml',
        baseContent: 'openapi: 3.0.0\ninfo:\n  title: OpenAI API\n  version: 1.0.0\ncomponents:\n  schemas:\n    ChatCompletionRequestMessage:\n      type: object\n      properties:\n        function_call:\n          type: object',
        redContent: 'openapi: 3.0.0\ninfo:\n  title: OpenAI API\n  version: 1.0.0\ncomponents:\n  schemas:\n    ChatCompletionRequestMessage:\n      type: object\n      properties: {}',
        greenContent: 'openapi: 3.0.0\ninfo:\n  title: OpenAI API\n  version: 1.0.0\ncomponents:\n  schemas:\n    ChatCompletionRequestMessage:\n      type: object\n      properties:\n        function_call:\n          type: object\n        metadata:\n          type: object'
    },
    {
        repoName: 'github-graphql',
        schemaFile: 'schema.graphql',
        baseContent: 'type User {\n  email: String\n}',
        redContent: 'type User {\n  id: ID\n}',
        greenContent: 'type User {\n  email: String\n  age: Int\n}'
    },
    {
        repoName: 'slack-webhooks',
        schemaFile: 'asyncapi.yaml',
        baseContent: 'asyncapi: 2.0.0\ninfo:\n  title: Slack API\n  version: 1.0.0\nchannels:\n  message.channels:\n    publish:\n      message:\n        payload:\n          type: object\n          properties:\n            channel_id:\n              type: string',
        redContent: 'asyncapi: 2.0.0\ninfo:\n  title: Slack API\n  version: 1.0.0\nchannels:\n  message.channels:\n    publish:\n      message:\n        payload:\n          type: object\n          properties: {}',
        greenContent: 'asyncapi: 2.0.0\ninfo:\n  title: Slack API\n  version: 1.0.0\nchannels:\n  message.channels:\n    publish:\n      message:\n        payload:\n          type: object\n          properties:\n            channel_id:\n              type: string\n            thread_ts:\n              type: string'
    },
    {
        repoName: 'jaffle-shop-db',
        schemaFile: 'schema.sql',
        baseContent: 'CREATE TABLE customers (\n  customer_lifetime_value INTEGER\n);',
        redContent: 'CREATE TABLE customers (\n);',
        greenContent: 'CREATE TABLE customers (\n  customer_lifetime_value INTEGER,\n  age INTEGER\n);'
    }
];

test.describe.serial('System Matrix E2E', () => {

    for (const repo of testMatrix) {
        test.describe.serial(`Repo: ${repo.repoName}`, () => {

            test.beforeAll(async () => {
                // Initial baseline push
                pushToGit(repo.repoName, {
                    [repo.schemaFile]: repo.baseContent,
                    'substrate.yaml': `consumers:\n  - name: test-consumer\n`
                });
                // Wait briefly to allow Forgejo/webhook pipeline
                await new Promise(r => setTimeout(r, 2000));
            });

            test('Test A (Red Path): Drop property and assert alert', async ({ page }) => {
                await page.goto('/org/admin/graph');

                pushToGit(repo.repoName, {
                    [repo.schemaFile]: repo.redContent,
                    'substrate.yaml': `consumers:\n  - name: test-consumer\n`
                });

                await expect(page.locator('.blast-radius-alert').or(page.locator('text=Broken'))).toBeVisible({ timeout: 15000 });
            });

            test('Test B (Green Path): Revert break, add safe property, assert green', async ({ page }) => {
                await page.goto('/org/admin/graph');

                pushToGit(repo.repoName, {
                    [repo.schemaFile]: repo.greenContent,
                    'substrate.yaml': `consumers:\n  - name: test-consumer\n`
                });

                await expect(page.locator('.blast-radius-alert').or(page.locator('text=Broken'))).not.toBeVisible({ timeout: 15000 });
            });

            test('Test C (Yellow Path): Break schema again with overrides, assert acknowledged', async ({ page }) => {
                await page.goto('/org/admin/graph');

                pushToGit(repo.repoName, {
                    [repo.schemaFile]: repo.redContent,
                    'substrate.yaml': `consumers:\n  - name: test-consumer\noverrides:\n  - approved_by: test@example.com\n    expires: '2099-01-01'\n    reason: Testing override\n`
                });

                await expect(
                    page.locator('text=Acknowledged').or(page.locator('.acknowledged-badge'))
                ).toBeVisible({ timeout: 15000 });
            });
        });
    }

    test.describe.serial('Configuration Modality', () => {
        test('Standalone Auto-Discovery vs Manual YAML', async ({ page }) => {
            const repoName = 'auto-discovery-test';
            await page.goto('/org/admin/graph');

            // 1. Auto-Discovery (no substrate.yaml)
            pushToGit(repoName, {
                'index.js': `const url = "https://api.stripe.com";\nconsole.log(url);`
            });

            // Wait for dynamic node appearance
            await expect(page.locator('.svelte-flow__node', { hasText: /stripe/i })).toBeVisible({ timeout: 15000 });

            // 2. Manual YAML
            pushToGit(repoName, {
                'index.js': `const url = "https://api.stripe.com";\nconsole.log(url);`,
                'substrate.yaml': `consumers:\n  - name: manual-consumer\n`
            });

            await expect(page.locator('.svelte-flow__node', { hasText: /manual-consumer/i })).toBeVisible({ timeout: 15000 });
        });
    });

});
