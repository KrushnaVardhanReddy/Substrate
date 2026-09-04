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
			<h1 class="page-title">{data.repo} Docs</h1>
			<p class="page-subtitle">Interactive API Documentation</p>
		</div>
		<a href={`/org/${data.org}/catalog/${data.repo}`} class="btn-back">
			Back to Catalog
		</a>
	</header>

	<div class="elements-wrapper">
		{#if data.specString}
			<elements-api apiDescriptionDocument={data.specString} router="hash" hideTryIt="true"></elements-api>
		{:else}
			<div class="no-spec">
				<p>No API specification found for this repository.</p>
			</div>
		{/if}
	</div>
</div>

<style>
	.detail-container {
		padding: 32px;
		max-width: 1400px;
		margin: 0 auto;
		width: 100%;
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

	.btn-back {
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

	.btn-back:hover {
		background-color: var(--bg-hover);
	}

	.elements-wrapper {
		flex-grow: 1;
		background-color: #fff;
		border: 1px solid var(--border);
		border-radius: 8px;
		overflow: hidden;
		position: relative;
	}

	.elements-wrapper :global(elements-api) {
		height: 100%;
		display: block;
	}

	.no-spec {
		padding: 32px;
		text-align: center;
		color: var(--text-muted);
		font-size: 1.1rem;
	}
</style>
