const assert = require('assert');
const child_process = require('child_process');
const path = require('path');
const fs = require('fs');

const indexJsPath = path.join(__dirname, 'index.js');
const wasmPath = path.join(__dirname, 'substrate.wasm');

// Check if WASM exists
if (!fs.existsSync(wasmPath)) {
  console.error("WASM file not found, test fails");
  process.exit(1);
}

const env = {
  ...process.env,
  INPUT_BASE_SCHEMA: 'base.yaml',
  INPUT_HEAD_SCHEMA: 'head.yaml',
  INPUT_CONFIG: 'sub.yaml'
};

fs.writeFileSync('base.yaml', 'openapi: "3.0.0"\ninfo:\n  title: API\n  version: 1.0.0\npaths: {}');
fs.writeFileSync('head.yaml', 'openapi: "3.0.0"\ninfo:\n  title: API\n  version: 1.0.0\npaths: {}');
fs.writeFileSync('sub.yaml', 'service: "test"');

try {
  const result = child_process.execSync(`node ${indexJsPath}`, { env });
  const stdout = result.toString();
  console.log("Output:");
  console.log(stdout);
  assert(stdout.includes('"schema_type": "openapi"'), "Output should include openapi schema type");
  console.log("Tests passed!");
} catch (e) {
  console.error("Test failed:");
  console.error(e.stdout ? e.stdout.toString() : "");
  console.error(e.stderr ? e.stderr.toString() : "");
  console.error(e);
  process.exit(1);
} finally {
  fs.unlinkSync('base.yaml');
  fs.unlinkSync('head.yaml');
  fs.unlinkSync('sub.yaml');
}
