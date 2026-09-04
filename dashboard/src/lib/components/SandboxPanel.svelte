<script lang="ts">
	import { onMount } from 'svelte';

	let { org, repo, apiBaseUrl } = $props<{ org: string; repo: string; apiBaseUrl: string }>();

	let token = $state<string | null>(null);
	let method = $state<string>('GET');
	let path = $state<string>('/');
	let body = $state<string>('');
	let headersText = $state<string>('{}');

	let loading = $state<boolean>(false);
	let response = $state<{ status_code: number; headers: Record<string, string>; body: string } | null>(null);
	let error = $state<string | null>(null);

	let isExpanded = $state<boolean>(false);

	async function fetchToken() {
		try {
			const res = await fetch(`${apiBaseUrl}/api/v1/sandbox/token`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('token')}` // assuming user is logged in
				},
				body: JSON.stringify({ org, repo })
			});
			if (!res.ok) {
				throw new Error('Failed to fetch sandbox token');
			}
			const data = await res.json();
			token = data.token;
		} catch (e: any) {
			error = e.message;
		}
	}

	async function sendRequest() {
		if (!token) {
			await fetchToken();
			if (!token) return;
		}

		loading = true;
		error = null;
		response = null;

		try {
			let headersMap: Record<string, string> = {};
			try {
				headersMap = JSON.parse(headersText);
			} catch (e) {
				throw new Error('Invalid JSON in headers field');
			}

			const res = await fetch(`${apiBaseUrl}/api/v1/sandbox/request`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${token}`
				},
				body: JSON.stringify({
					org,
					repo,
					method,
					path,
					headers: headersMap,
					body
				})
			});

			if (!res.ok) {
				const errBody = await res.text();
				throw new Error(`Proxy error: ${res.status} - ${errBody}`);
			}

			response = await res.json();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}
</script>

<div class="sandbox-panel">
	<button class="sandbox-toggle" onclick={() => { isExpanded = !isExpanded; if(isExpanded && !token) fetchToken(); }}>
		{isExpanded ? '▼ Close API Sandbox' : '▶ Try It Out (Interactive API Sandbox)'}
	</button>

	{#if isExpanded}
		<div class="sandbox-content">
			<div class="request-builder">
				<h4>Request Builder</h4>
				<div class="input-group row">
					<select bind:value={method} class="method-select" aria-label="HTTP Method">
						<option value="GET">GET</option>
						<option value="POST">POST</option>
						<option value="PUT">PUT</option>
						<option value="DELETE">DELETE</option>
						<option value="PATCH">PATCH</option>
					</select>
					<input type="text" bind:value={path} placeholder="/api/users" class="path-input" aria-label="Path" />
				</div>

				<div class="input-group">
					<label for="headers-input">Headers (JSON format)</label>
					<textarea id="headers-input" bind:value={headersText} rows="2"></textarea>
				</div>

				<div class="input-group">
					<label for="body-input">Body</label>
					<textarea id="body-input" bind:value={body} rows="4" placeholder=""></textarea>
				</div>

				<button class="btn-send" onclick={sendRequest} disabled={loading}>
					{loading ? 'Sending...' : 'Send Request'}
				</button>

				{#if error}
					<div class="error-msg">{error}</div>
				{/if}
			</div>

			<div class="response-viewer">
				<h4>Response</h4>
				<div id="sandbox-response">
					{#if loading}
						<div class="loader">Loading...</div>
					{:else if response}
						<div class="status-code" class:success={response.status_code < 400} class:error={response.status_code >= 400}>
							Status: {response.status_code}
						</div>
						<pre><code>{response.body}</code></pre>
					{:else}
						<div class="empty-state">Submit a request to see the response.</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.sandbox-panel {
		margin: 20px 0;
		border: 1px solid var(--border);
		border-radius: 8px;
		background-color: var(--bg-card);
		overflow: hidden;
	}

	.sandbox-toggle {
		width: 100%;
		text-align: left;
		padding: 12px 16px;
		background-color: var(--bg-hover);
		border: none;
		font-weight: 600;
		color: var(--text-main);
		cursor: pointer;
		font-size: 14px;
	}

	.sandbox-toggle:hover {
		background-color: #e5e7eb;
	}

	.sandbox-content {
		display: flex;
		flex-direction: column;
		gap: 20px;
		padding: 16px;
		border-top: 1px solid var(--border);
	}

	@media (min-width: 768px) {
		.sandbox-content {
			flex-direction: row;
		}
		.request-builder, .response-viewer {
			flex: 1;
		}
	}

	.request-builder {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.response-viewer {
		background: #1e1e1e;
		color: #d4d4d4;
		border-radius: 6px;
		padding: 12px;
		overflow: auto;
		max-height: 400px;
	}

	h4 {
		margin: 0 0 8px 0;
		font-size: 14px;
		color: var(--text-main);
	}

	.response-viewer h4 {
		color: #fff;
	}

	.input-group {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.input-group.row {
		flex-direction: row;
	}

	label {
		font-size: 12px;
		color: var(--text-muted);
	}

	input, select, textarea {
		padding: 8px;
		border: 1px solid var(--border);
		border-radius: 4px;
		font-family: monospace;
		background: #fff;
	}

	.method-select {
		width: 100px;
	}

	.path-input {
		flex-grow: 1;
	}

	.btn-send {
		background-color: #2563eb;
		color: white;
		border: none;
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
		font-weight: 600;
	}

	.btn-send:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.error-msg {
		color: #dc2626;
		font-size: 13px;
		margin-top: 8px;
	}

	.status-code {
		font-weight: 600;
		margin-bottom: 8px;
	}

	.status-code.success {
		color: #4ade80;
	}

	.status-code.error {
		color: #f87171;
	}

	pre {
		margin: 0;
		white-space: pre-wrap;
		word-break: break-all;
	}

	.empty-state, .loader {
		font-size: 13px;
		color: #888;
		text-align: center;
		padding: 20px;
	}
</style>
