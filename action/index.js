const fs = require('node:fs');
const { WASI } = require('node:wasi');
const path = require('node:path');

const wasi = new WASI({
  version: 'preview1',
  args: [
    'substrate',
    'diff',
    process.env.INPUT_BASE_SCHEMA || 'openapi.yaml',
    process.env.INPUT_HEAD_SCHEMA || 'openapi.yaml',
    '--config',
    process.env.INPUT_CONFIG || 'substrate.yaml'
  ],
  env: process.env,
  preopens: {
    '/': '/',
    '.': process.cwd()
  },
  returnOnExit: true
});

const importObject = { wasi_snapshot_preview1: wasi.wasiImport };

(async () => {
  try {
    const wasmFile = path.join(__dirname, 'substrate.wasm');
    const wasm = await WebAssembly.compile(fs.readFileSync(wasmFile));
    const instance = await WebAssembly.instantiate(wasm, importObject);
    const exitCode = wasi.start(instance);
    if (exitCode !== 0) {
      process.exitCode = exitCode;
    }
  } catch (err) {
    console.error(err);
    process.exitCode = 1;
  }
})();
