<script lang="ts">
	import type { PageData } from './$types';
	import { onMount } from 'svelte';
	import { marked } from 'marked';
	import DOMPurify from 'dompurify';
	import * as yaml from 'js-yaml';
	import CodeSnippetViewer from '$lib/components/CodeSnippetViewer.svelte';

	let { data }: { data: PageData } = $props();

	let activeTab = $state('schema');
	let guides = $state<any[]>([]);
	let selectedGuideContent = $state<string | null>(null);

	let endpoints = $state<Array<{ path: string; method: string; summary: string }>>([]);

	// Parse endpoints from OpenAPI yaml
	$effect(() => {
		if (activeTab === 'snippets' && endpoints.length === 0 && data.yamlString) {
			try {
				const doc = yaml.load(data.yamlString) as any;
				if (doc && doc.paths) {
					const parsedEndpoints = [];
					for (const [path, methods] of Object.entries(doc.paths)) {
						for (const [method, details] of Object.entries(methods as any)) {
							// Filter out common OpenAPI keys that aren't HTTP methods
							if (!['parameters', 'servers', 'summary', 'description'].includes(method)) {
								parsedEndpoints.push({
									path,
									method: method.toUpperCase(),
									summary: (details as any).summary || ''
								});
							}
						}
					}
					endpoints = parsedEndpoints;
				}
			} catch (e) {
				console.error("Failed to parse OpenAPI YAML for snippets", e);
			}
		}
	});

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
		<button id="snippets-tab" class="tab-btn" class:active={activeTab === 'snippets'} onclick={() => activeTab = 'snippets'}>Snippets</button>
		<button id="business-context-tab" class="tab-btn" class:active={activeTab === 'business'} onclick={() => activeTab = 'business'}>Business Context</button>
	</div>

	{#if activeTab === 'schema'}
		<div class="elements-wrapper">
			<elements-api apiDescriptionDocument={data.yamlString} router="hash" hideTryIt="true"></elements-api>
		</div>
	{:else if activeTab === 'snippets'}
		<div class="snippets-wrapper">
			<div class="snippets-container">
				{#if endpoints.length > 0}
					{#each endpoints as endpoint}
						<div class="endpoint-card">
							<div class="endpoint-header">
								<span class={`method-badge method-${endpoint.method.toLowerCase()}`}>{endpoint.method}</span>
								<span class="endpoint-path">{endpoint.path}</span>
							</div>
							{#if endpoint.summary}
								<p class="endpoint-summary">{endpoint.summary}</p>
							{/if}
							<div class="snippet-viewer-container">
								<CodeSnippetViewer
									method={endpoint.method}
									path={`${data.apiBaseUrl || 'https://api.example.com'}${endpoint.path}`}
								/>
							</div>
						</div>
					{/each}
				{:else}
					<p class="no-endpoints">No endpoints found to generate snippets.</p>
				{/if}
			</div>
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
	{:else if activeTab === 'business'}
		<div class="business-context-wrapper">
			<div class="business-card">
				<h3>Business Context</h3>
				{#if data.repoData?.metadata}
					<div class="metadata-grid">
						{#if data.repoData.metadata.owner}
							<div class="metadata-item">
								<span class="metadata-label">Owner</span>
								<span class="metadata-value">{data.repoData.metadata.owner}</span>
							</div>
						{/if}
						{#if data.repoData.metadata.slack_channel}
							<div class="metadata-item">
								<span class="metadata-label">Slack Channel</span>
								<span class="metadata-value">{data.repoData.metadata.slack_channel}</span>
							</div>
						{/if}
						{#if data.repoData.metadata.pagerduty}
							<div class="metadata-item">
								<span class="metadata-label">PagerDuty</span>
								<span class="metadata-value">{data.repoData.metadata.pagerduty}</span>
							</div>
						{/if}
						{#if data.repoData.metadata.pm}
							<div class="metadata-item">
								<span class="metadata-label">Product Manager</span>
								<span class="metadata-value">{data.repoData.metadata.pm}</span>
							</div>
						{/if}
						{#if data.repoData.metadata.sla_tier}
							<div class="metadata-item">
								<span class="metadata-label">SLA Tier</span>
								<span class="metadata-value">{data.repoData.metadata.sla_tier}</span>
							</div>
						{/if}
					</div>
					{#if !data.repoData.metadata.owner && !data.repoData.metadata.slack_channel && !data.repoData.metadata.pagerduty && !data.repoData.metadata.pm && !data.repoData.metadata.sla_tier}
						<p class="no-metadata">No business metadata provided.</p>
					{/if}
				{:else}
					<p class="no-metadata">No business metadata found for this repository.</p>
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

	.business-context-wrapper {
		flex-grow: 1;
		background-color: #fff;
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.business-card {
		background: var(--bg-card);
		border: 1px solid var(--border);
		padding: 24px;
		border-radius: 8px;
	}

	.business-card h3 {
		margin-top: 0;
		margin-bottom: 24px;
		font-size: 1.25rem;
		color: var(--text-main);
	}

	.metadata-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 24px;
	}

	.metadata-item {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.metadata-label {
		font-size: 0.875rem;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.metadata-value {
		font-size: 1rem;
		color: var(--text-main);
		font-weight: 500;
	}

	.no-metadata {
		color: var(--text-muted);
		font-style: italic;
	}

	.snippets-wrapper {
		flex-grow: 1;
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 24px;
		overflow-y: auto;
	}

	.snippets-container {
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.endpoint-card {
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 20px;
		background: var(--bg-card);
	}

	.endpoint-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
	}

	.method-badge {
		font-size: 13px;
		font-weight: 700;
		padding: 4px 8px;
		border-radius: 4px;
		color: #fff;
	}

	.method-get { background-color: #10b981; }
	.method-post { background-color: #3b82f6; }
	.method-put { background-color: #f59e0b; }
	.method-delete { background-color: #ef4444; }
	.method-patch { background-color: #14b8a6; }
	.method-options { background-color: #8b5cf6; }

	.endpoint-path {
		font-family: monospace;
		font-size: 1.1rem;
		color: var(--text-main);
	}

	.endpoint-summary {
		margin-top: 0;
		margin-bottom: 16px;
		color: var(--text-muted);
		font-size: 0.95rem;
	}

	.snippet-viewer-container {
		margin-top: 16px;
	}

	.no-endpoints {
		color: var(--text-muted);
		text-align: center;
		padding: 48px;
		font-size: 1.1rem;
	}
</style>
