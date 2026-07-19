const assert = require('assert');
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

describe('CLI E2E Tests', function() {
    this.timeout(10000);

    let baseFile;
    let headFile;
    let headBreakingFile;
    const cliPath = path.resolve(__dirname, 'cli.js');

    before(() => {
        baseFile = path.resolve(__dirname, 'test-base.yaml');
        headFile = path.resolve(__dirname, 'test-head.yaml');
        headBreakingFile = path.resolve(__dirname, 'test-head-breaking.yaml');

        fs.writeFileSync(baseFile, `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
`);

        fs.writeFileSync(headFile, `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
`);

        fs.writeFileSync(headBreakingFile, `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
`);

        // Build the WASM for the test
        execSync('cd ../../engine && GOOS=js GOARCH=wasm go build -o substrate-cli.wasm ./cmd/wasm/main.go', { stdio: 'inherit' });
    });

    after(() => {
        if (fs.existsSync(baseFile)) fs.unlinkSync(baseFile);
        if (fs.existsSync(headFile)) fs.unlinkSync(headFile);
        if (fs.existsSync(headBreakingFile)) fs.unlinkSync(headBreakingFile);
    });

    it('should detect no breaking changes', () => {
        const output = execSync(`node ${cliPath} diff --base ${baseFile} --head ${headFile} --fail-on-breaking`).toString();
        const jsonOutput = JSON.parse(output);
        assert.strictEqual(jsonOutput.summary.breaking_count, 0);
    });

    it('should exit with code 1 and error output on breaking changes', () => {
        try {
            execSync(`node ${cliPath} diff --base ${baseFile} --head ${headBreakingFile} --fail-on-breaking`, { stdio: 'pipe' });
            assert.fail('Expected command to throw due to exit code 1');
        } catch (error) {
            assert.strictEqual(error.status, 1);
            assert(error.stderr.toString().includes('Breaking changes detected'));
        }
    });

    it('should create husky pre-commit hook', () => {
        const huskyDir = path.join(process.cwd(), '.husky');
        const preCommitPath = path.join(huskyDir, 'pre-commit');
        if (fs.existsSync(huskyDir)) {
            fs.rmSync(huskyDir, { recursive: true, force: true });
        }

        execSync(`node ${cliPath} init --husky`);
        assert(fs.existsSync(preCommitPath));
        const content = fs.readFileSync(preCommitPath, 'utf8');
        assert(content.includes('npx @substrate/cli diff'));

        if (fs.existsSync(huskyDir)) {
            fs.rmSync(huskyDir, { recursive: true, force: true });
        }
    });
});
