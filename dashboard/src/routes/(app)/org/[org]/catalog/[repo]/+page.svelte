<script lang="ts">
	import type { PageData } from './$types';
	import { onMount } from 'svelte';
	import { marked } from 'marked';
	import DOMPurify from 'dompurify';

	let { data }: { data: PageData } = $props();

	let activeTab = $state('schema');
	let guides = $state<any[]>([]);
	let selectedGuideContent = $state<string | null>(null);

	// In Svelte 5, we use an effect to react to data changes if needed,
	// but for initial fetch we can use an effect.
	$effect(() => {
		if (activeTab === 'guides' && guides.length === 0) {
			fetch(`/api/v1/docs/${data.org}/${data.repo}`)
				.then(res => res.json())
				.then(data => {
					guides = data || [];
				})
				.catch(err => console.error("Failed to load guides", err));
		}
	});

	function selectGuide(slug: string) {
		fetch(`/api/v1/docs/${data.org}/${data.repo}/${slug}`)
			.then(res => res.json())
			.then(guide => {
				const html = marked.parse(guide.content);
				// DOMPurify needs to be used safely, especially in Node environment for SSR
				// But this runs on client side when clicking a guide, so DOMPurify works.
				selectedGuideContent = DOMPurify.sanitize(html as string);
			})
			.catch(err => console.error("Failed to load guide", err));
	}
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
		<div class="header-actions">
			<a href={`/org/${data.org}/catalog/${data.repo}/docs`} class="btn-docs">
				View API Docs
			</a>
			<a href={`https://github.com/${data.org}/${data.repo}`} target="_blank" rel="noopener noreferrer" class="btn-github">
				View on GitHub
			</a>
		</div>
	</header>

	<div class="tabs">
		<button class="tab-btn" class:active={activeTab === 'schema'} onclick={() => activeTab = 'schema'}>Schema</button>
		<button id="guides-tab" class="tab-btn" class:active={activeTab === 'guides'} onclick={() => activeTab = 'guides'}>Guides</button>
	</div>

	{#if activeTab === 'schema'}
		<div class="elements-wrapper">
			<elements-api apiDescriptionDocument={data.yamlString} router="hash" hideTryIt="true"></elements-api>
		</div>
	{:else if activeTab === 'guides'}
		<div class="guides-wrapper">
			<div class="guides-sidebar">
				<h3>Guides</h3>
				<ul>
					{#each guides as guide}
						<li>
							<button class="guide-link" onclick={() => selectGuide(guide.file_path)}>{guide.title}</button>
						</li>
					{/each}
					{#if guides.length === 0}
						<li>No guides found.</li>
					{/if}
				</ul>
			</div>
			<div class="guide-content">
				{#if selectedGuideContent !== null}
					{@html selectedGuideContent}
				{:else}
					<p>Select a guide to view.</p>
				{/if}
			</div>
		</div>
	{/if}
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

	.header-actions {
		display: flex;
		gap: 12px;
		align-items: center;
	}

	.btn-docs,
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

	.btn-docs:hover,
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

	.tabs {
		display: flex;
		gap: 16px;
		margin-bottom: 16px;
		border-bottom: 1px solid var(--border);
	}

	.tab-btn {
		background: none;
		border: none;
		padding: 8px 16px;
		font-size: 16px;
		cursor: pointer;
		color: var(--text-muted);
		border-bottom: 2px solid transparent;
	}

	.tab-btn.active {
		color: var(--text-main);
		border-bottom: 2px solid var(--primary-main, #007bff);
	}

	.guides-wrapper {
		display: flex;
		flex-grow: 1;
		gap: 24px;
		background-color: #fff;
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 24px;
		overflow: hidden;
	}

	.guides-sidebar {
		width: 250px;
		flex-shrink: 0;
		border-right: 1px solid var(--border);
		padding-right: 16px;
	}

	.guides-sidebar h3 {
		margin-top: 0;
	}

	.guides-sidebar ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.guides-sidebar li {
		margin-bottom: 8px;
	}

	.guide-link {
		background: none;
		border: none;
		color: var(--text-main);
		cursor: pointer;
		text-align: left;
		padding: 4px 8px;
		width: 100%;
		border-radius: 4px;
	}

	.guide-link:hover {
		background-color: var(--bg-hover);
	}

	.guide-content {
		flex-grow: 1;
		overflow-y: auto;
	}
</style>
