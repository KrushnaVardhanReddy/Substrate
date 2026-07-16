declare const Go: any;

let wasmEngineInitialized = false;
let initPromise: Promise<void> | null = null;

/**
 * Initializes the Go WebAssembly runtime and fetches the `engine.wasm` binary.
 */
export async function initWasmEngine() {
	if (wasmEngineInitialized) return;
	if (initPromise) return initPromise;

	initPromise = new Promise<void>(async (resolve, reject) => {
		try {
			const go = new Go();
			if (typeof WebAssembly.instantiateStreaming === 'function') {
				const result = await WebAssembly.instantiateStreaming(
					fetch('/engine.wasm'),
					go.importObject
				);
				go.run(result.instance);
			} else {
				// Fallback for older browsers
				const response = await fetch('/engine.wasm');
				const buffer = await response.arrayBuffer();
				const result = await WebAssembly.instantiate(buffer, go.importObject);
				go.run(result.instance);
			}
			wasmEngineInitialized = true;
			console.log('Substrate WASM Engine initialized.');
			resolve();
		} catch (err) {
			console.error('Failed to initialize WASM engine:', err);
			reject(err);
		}
	});

	return initPromise;
}

/**
 * Wrapper for the global diff_schemas function injected by the WASM engine.
 */
export async function diffSchemas(baseSpec: string, revisionSpec: string): Promise<any> {
	await initWasmEngine();

	// @ts-ignore
	if (typeof window.diff_schemas !== 'function') {
		throw new Error('diff_schemas function not exported by WASM engine');
	}

	// @ts-ignore
	const result = window.diff_schemas(baseSpec, revisionSpec);
	
	if (typeof result === 'string') {
		try {
			const parsed = JSON.parse(result);
			if (parsed.error) throw new Error(parsed.error);
			return parsed;
		} catch (e) {
			// If it's already a valid JSON string but parsing fails, or it's a generic string
			throw new Error('Failed to parse WASM output: ' + result);
		}
	} else if (result && result.error) {
		throw new Error(result.error);
	}

	return result;
}
