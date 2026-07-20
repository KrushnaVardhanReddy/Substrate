<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
	<link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
</svelte:head>

<div class="detail-container">
	<header class="detail-header">
		<div>
			<h1 class="page-title">{data.repo}</h1>
			<p class="page-subtitle">API Documentation and Service Details</p>

			<div class="badge-snippet">
				<span class="badge-label">Embed Badge:</span>
				<code class="badge-code" data-testid="badge-snippet">[![Contract Score]({data.apiBaseUrl}/api/badges/{data.org}/{data.repo})]({data.dashboardUrl}/org/{data.org}/catalog/{data.repo})</code>
				<button class="btn-copy" onclick={() => navigator.clipboard.writeText(`[![Contract Score](${data.apiBaseUrl}/api/badges/${data.org}/${data.repo})](${data.dashboardUrl}/org/${data.org}/catalog/${data.repo})`)}>Copy</button>
			</div>
		</div>
		<a href={`https://github.com/${data.org}/${data.repo}`} target="_blank" rel="noopener noreferrer" class="btn-github">
			View on GitHub
		</a>
	</header>

	<div class="elements-wrapper">
		<elements-api apiDescriptionDocument={data.yamlString} router="hash" hideTryIt="true"></elements-api>
	</div>
</div>

<style>
	.detail-container {
		padding: 32px;
		max-width: 1400px;
		margin: 0 auto;
		width: 100%;
		height: calc(100vh - 60px); /* Adjust for topnav */
		display: flex;
		flex-direction: column;
	}

	.detail-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		margin-bottom: 24px;
		flex-shrink: 0;
	}

	.page-title {
		font-size: 2rem;
		margin-bottom: 8px;
		color: var(--text-main);
	}

	.page-subtitle {
		color: var(--text-muted);
		font-size: 1.05rem;
	}

	.btn-github {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		color: var(--text-main);
		padding: 8px 16px;
		border-radius: 6px;
		text-decoration: none;
		font-size: 14px;
		font-weight: 500;
		transition: background-color 0.2s;
	}

	.btn-github:hover {
		background-color: var(--bg-hover);
	}

	.badge-snippet {
		margin-top: 16px;
		display: flex;
		align-items: center;
		gap: 8px;
		background: var(--bg-card);
		border: 1px solid var(--border);
		padding: 8px 12px;
		border-radius: 6px;
	}

	.badge-label {
		font-weight: 500;
		color: var(--text-main);
		font-size: 14px;
	}

	.badge-code {
		font-family: monospace;
		color: var(--text-muted);
		font-size: 13px;
		background: rgba(0,0,0,0.05);
		padding: 4px 8px;
		border-radius: 4px;
		user-select: all;
	}

	.btn-copy {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		color: var(--text-main);
		padding: 4px 10px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 13px;
		transition: background-color 0.2s;
	}

	.btn-copy:hover {
		background-color: var(--bg-hover);
	}

	.elements-wrapper {
		flex-grow: 1;
		background-color: #fff; /* Elements needs a light background generally, or we override its variables */
		border: 1px solid var(--border);
		border-radius: 8px;
		overflow: hidden;
		position: relative;
	}

	/* Elements component styles are injected by stoplight but we may need to make sure the wrapper fills space */
	.elements-wrapper :global(elements-api) {
		height: 100%;
		display: block;
	}
</style>
