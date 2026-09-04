<script lang="ts">
	import { generateSnippets } from '$lib/utils/snippetGenerator';

	let {
		method = 'GET',
		path = '',
		headers = {},
		queryParams = {},
		body = null
	} = $props<{
		method: string;
		path: string;
		headers?: Record<string, string>;
		queryParams?: Record<string, string>;
		body?: any;
	}>();

	let activeLanguage: 'curl' | 'python' | 'node' | 'go' = $state('curl');
	let copySuccess = $state(false);

	let snippets = $derived(generateSnippets(method, path, headers, queryParams, body));

	async function handleCopy() {
		try {
			await navigator.clipboard.writeText(snippets[activeLanguage]);
			copySuccess = true;
			setTimeout(() => {
				copySuccess = false;
			}, 2000);
		} catch (err) {
			console.error('Failed to copy snippet: ', err);
		}
	}
</script>

<div class="snippet-viewer">
	<div class="tab-bar">
		<div class="tabs">
			<button
				class="tab-btn"
				class:active={activeLanguage === 'curl'}
				onclick={() => (activeLanguage = 'curl')}
			>
				cURL
			</button>
			<button
				class="tab-btn"
				class:active={activeLanguage === 'python'}
				onclick={() => (activeLanguage = 'python')}
			>
				Python
			</button>
			<button
				class="tab-btn"
				class:active={activeLanguage === 'node'}
				onclick={() => (activeLanguage = 'node')}
			>
				Node.js
			</button>
			<button
				class="tab-btn"
				class:active={activeLanguage === 'go'}
				onclick={() => (activeLanguage = 'go')}
			>
				Go
			</button>
		</div>
		<button class="btn-copy" onclick={handleCopy} aria-label="Copy code">
			{#if copySuccess}
				Copied!
			{:else}
				Copy
			{/if}
		</button>
	</div>
	<div class="code-container">
		<pre><code>{snippets[activeLanguage]}</code></pre>
	</div>
</div>

<style>
	.snippet-viewer {
		background-color: var(--bg-card, #ffffff);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: 8px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		font-family:
			system-ui,
			-apple-system,
			sans-serif;
	}

	.tab-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		background-color: var(--bg-subtle, #f9fafb);
		border-bottom: 1px solid var(--border, #e5e7eb);
		padding: 0 8px;
	}

	.tabs {
		display: flex;
		gap: 4px;
	}

	.tab-btn {
		background: none;
		border: none;
		padding: 8px 12px;
		font-size: 14px;
		font-weight: 500;
		color: var(--text-muted, #6b7280);
		cursor: pointer;
		border-bottom: 2px solid transparent;
		transition:
			color 0.2s,
			border-color 0.2s;
	}

	.tab-btn:hover {
		color: var(--text-main, #111827);
	}

	.tab-btn.active {
		color: var(--primary-main, #2563eb);
		border-bottom-color: var(--primary-main, #2563eb);
	}

	.btn-copy {
		background-color: var(--bg-card, transparent);
		border: 1px solid var(--border, #e5e7eb);
		color: var(--text-main, #111827);
		padding: 4px 10px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 13px;
		transition: background-color 0.2s;
	}

	.btn-copy:hover {
		background-color: var(--bg-hover, #f3f4f6);
	}

	.code-container {
		padding: 16px;
		background-color: #1e1e1e; /* Dark theme for code */
		overflow-x: auto;
	}

	pre {
		margin: 0;
	}

	code {
		font-family: 'JetBrains Mono', 'Fira Code', 'Courier New', Courier, monospace;
		font-size: 13px;
		line-height: 1.5;
		color: #d4d4d4; /* Light text for dark theme */
		white-space: pre-wrap;
	}
</style>
