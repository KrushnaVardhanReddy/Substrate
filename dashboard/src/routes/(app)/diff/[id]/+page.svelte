<script lang="ts">
	import DiffViewer from '$lib/components/DiffViewer.svelte';
	let { data } = $props();

	// Mock YAML data as fallback
	const mockOldYaml = `name: User Service
version: 1.0.0
description: Manages user accounts
dependencies:
  - name: Auth Service
    version: ^1.2.0`;

	const mockNewYaml = `name: User Service
version: 1.1.0
description: Manages user accounts and profiles
dependencies:
  - name: Auth Service
    version: ^2.0.0
  - name: Email Service
    version: ^1.0.0`;

	let before = $derived(data.oldText || mockOldYaml);
	let after = $derived(data.newText || mockNewYaml);
</script>

<div class="diff-page">
	<div class="diff-header">
		<h1 class="page-title">Schema Diff Viewer</h1>
		<p class="page-subtitle">Changes between previous and current schema.</p>
	</div>

	<div class="card diff-card">
		<DiffViewer {before} {after} />
	</div>
</div>

<style>
	.diff-page {
		padding: 20px;
	}
	.diff-header {
		margin-bottom: 20px;
	}
	.page-title {
		font-size: 1.5rem;
		margin-bottom: 8px;
		color: var(--text-main);
	}
	.page-subtitle {
		color: var(--text-muted);
		font-size: 0.95rem;
	}
	.diff-card {
		padding: 0;
		overflow: hidden;
	}
</style>
