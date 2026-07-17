<script lang="ts">
	let { data } = $props();
</script>

<svelte:head>
	<script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
	<link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
</svelte:head>

<div class="api-docs-container">
	<div class="api-docs-header">
		<div>
			<a href={`/org/${data.org}/catalog`} class="back-link">
				&larr; Back to Catalog
			</a>
			<h1 class="page-title">{data.repoName} API</h1>
		</div>
	</div>

	<div class="elements-wrapper">
		<elements-api apiDescriptionDocument={data.rawYaml} router="hash" layout="sidebar"></elements-api>
	</div>
</div>

<style>
	.api-docs-container {
		display: flex;
		flex-direction: column;
		height: calc(100vh - 60px); /* Adjust based on top nav height */
		background-color: var(--bg-dark); /* Ensure Stoplight blends as well as possible */
		width: 100%;
	}

	.api-docs-header {
		padding: 24px 32px;
		background-color: var(--bg-card);
		border-bottom: 1px solid var(--border);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.back-link {
		color: var(--text-muted);
		text-decoration: none;
		font-size: 0.9rem;
		margin-bottom: 8px;
		display: inline-block;
	}

	.back-link:hover {
		color: var(--accent);
	}

	.page-title {
		font-size: 1.5rem;
		color: var(--text-main);
		margin: 0;
	}

	.elements-wrapper {
		flex-grow: 1;
		overflow: hidden; /* Stoplight Elements handles its own scrolling */
		background-color: #fff; /* Stoplight elements is generally light-themed by default */
	}

	/* Target inner web components to try to force some CSS variables if possible,
	   though Elements encapsulates styles heavily. We might need a transparent background wrapper. */
	:global(elements-api) {
		height: 100%;
		display: block;
	}
</style>
