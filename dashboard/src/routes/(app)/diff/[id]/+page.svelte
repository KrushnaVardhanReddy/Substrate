<script lang="ts">
	import DiffViewer from '$lib/components/DiffViewer.svelte';
	let { data } = $props();
</script>

<div class="diff-page">
	<div class="page-header">
		<h1 class="page-title">Schema Diff Viewer</h1>
		<p class="page-subtitle">Review the detailed exact schema changes.</p>
	</div>

	<div class="diff-container">
		<DiffViewer before={data.oldText} after={data.newText} />
	</div>
</div>

<style>
	.diff-page {
		padding: 30px;
		max-width: 1200px;
		margin: 0 auto;
	}
	.page-header {
		margin-bottom: 20px;
	}
	.page-title {
		font-size: 1.8rem;
		margin-bottom: 5px;
	}
	.page-subtitle {
		color: var(--text-muted);
		font-size: 0.95rem;
	}
	.diff-container {
		margin-top: 20px;
		background: var(--bg-card);
		border-radius: 8px;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	}
</style>