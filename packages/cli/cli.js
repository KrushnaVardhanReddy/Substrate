#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Load wasm_exec.js from Go environment if available, otherwise use a minimal polyfill
const wasmExecPath = process.env.GOROOT ? path.join(process.env.GOROOT, 'lib', 'wasm', 'wasm_exec.js') : '/usr/local/go/lib/wasm/wasm_exec.js';
if (fs.existsSync(wasmExecPath)) {
    require(wasmExecPath);
} else {
    console.error("Warning: wasm_exec.js not found. WASM execution may fail.");
    global.Go = class {
        async run() {}
    };
}

const wasmPath = path.resolve(__dirname, '../../engine/substrate-cli.wasm');

async function runDiff(baseStr, headStr, failOnBreaking) {
    if (!fs.existsSync(wasmPath)) {
        console.error(`Error: WASM binary not found at ${wasmPath}`);
        console.error("Please build the engine using: GOOS=js GOARCH=wasm go build -o substrate-cli.wasm ./cmd/wasm/main.go");
        process.exit(1);
    }

    const go = new Go();
    try {
        const wasmBuffer = fs.readFileSync(wasmPath);
        const { instance } = await WebAssembly.instantiate(wasmBuffer, go.importObject);
        go.run(instance);

        const resultStr = global.substrateDiff(baseStr, headStr);
        let result;
        try {
            result = JSON.parse(resultStr);
        } catch (e) {
            console.error("Failed to parse diff result:", resultStr);
            process.exit(1);
        }

        if (result.error) {
            console.error("Error from diff engine:", result.error);
            process.exit(1);
        }

        console.log(JSON.stringify(result, null, 2));

        if (failOnBreaking && result.breaking_changes && result.breaking_changes.length > 0) {
            console.error("\nError: Breaking changes detected in API schema!");
            process.exit(1);
        }
    } catch (e) {
        console.error("Failed to execute WASM:", e);
        process.exit(1);
    }
}

function initHusky() {
    const huskyDir = path.join(process.cwd(), '.husky');
    const preCommitPath = path.join(huskyDir, 'pre-commit');

    if (!fs.existsSync(huskyDir)) {
        fs.mkdirSync(huskyDir, { recursive: true });
    }

    const hookContent = `#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

# Note: pre-commit hook compares origin/main to the locally staged/unstaged schema.yaml
npx @substrate/cli diff --base origin/main --head schema.yaml --fail-on-breaking
`;

    fs.writeFileSync(preCommitPath, hookContent);
    fs.chmodSync(preCommitPath, '755');
    console.log("Successfully added Substrate diff engine to .husky/pre-commit");
}

async function main() {
    const args = process.argv.slice(2);

    if (args.length === 0) {
        console.log("Usage: substrate <command> [options]");
        console.log("Commands:");
        console.log("  diff --base <file_or_ref> --head <file_or_ref> [--fail-on-breaking]");
        console.log("  init --husky");
        process.exit(0);
    }

    const command = args[0];

    if (command === 'init' && args[1] === '--husky') {
        initHusky();
        return;
    }

    if (command === 'diff') {
        let baseRef = '';
        let headRef = '';
        let failOnBreaking = false;

        for (let i = 1; i < args.length; i++) {
            if (args[i] === '--base') {
                baseRef = args[++i];
            } else if (args[i] === '--head') {
                headRef = args[++i];
            } else if (args[i] === '--fail-on-breaking') {
                failOnBreaking = true;
            }
        }

        if (!baseRef || !headRef) {
            console.error("Error: --base and --head arguments are required for 'diff'");
            process.exit(1);
        }

        let baseStr = '';
        let headStr = '';

        const resolveRef = (ref) => {
            if (fs.existsSync(ref)) {
                return fs.readFileSync(ref, 'utf8');
            }
            const { execFileSync } = require('child_process');
            try {
                // Try fetching a standard file if the ref implies a commit/branch
                // Using execFileSync to prevent shell injection vulnerabilities
                return execFileSync('git', ['show', `${ref}:schema.yaml`], { stdio: ['pipe', 'pipe', 'ignore'] }).toString();
            } catch (e) {
                // If it fails, maybe the ref itself was meant to be exactly something like "origin/main:path/to/spec.yaml"
                return execFileSync('git', ['show', ref], { stdio: ['pipe', 'pipe', 'ignore'] }).toString();
            }
        };

        try {
            baseStr = resolveRef(baseRef);
        } catch (e) {
            console.error(`Failed to read base ref ${baseRef}:`, e.message);
            process.exit(1);
        }

        try {
            headStr = resolveRef(headRef);
        } catch (e) {
            console.error(`Failed to read head ref ${headRef}:`, e.message);
            process.exit(1);
        }

        await runDiff(baseStr, headStr, failOnBreaking);
        return;
    }

    console.error(`Unknown command: ${command}`);
    process.exit(1);
}

main();
