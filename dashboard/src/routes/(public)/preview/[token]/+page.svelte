<script lang="ts">
	import DiffViewer from '$lib/components/DiffViewer.svelte';
	let { data } = $props();
</script>

<div class="diff-page">
	<div class="page-header">
		<h1 class="page-title">Schema Preview</h1>
		<p class="page-subtitle">Interactive diff for the proposed changes.</p>
	</div>

	<div class="diff-container">
		{#if data.error}
			<div class="error-box">
				<h2>Preview Not Available</h2>
				<p>{data.error}</p>
			</div>
		{:else}
			<DiffViewer before={data.oldText} after={data.newText} />
		{/if}
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
	.error-box {
		padding: 40px;
		text-align: center;
		color: var(--text-muted);
	}
</style>
