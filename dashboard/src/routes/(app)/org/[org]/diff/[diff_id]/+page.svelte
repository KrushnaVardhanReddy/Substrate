// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import DiffViewer from '$lib/components/DiffViewer.svelte';
	import { page } from '$app/stores';
	let { data } = $props();

	let isGenerating = $state(false);
	let generatedPatch = $state('');
	let patchApplied = $state(false);
	let sseStream = $state('');

	async function generateAIPatch() {
		isGenerating = true;
		sseStream = '';
		patchApplied = false;

		const res = await fetch(`/api/v1/ai/autofix`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				org: $page.params.org,
				diff_id: $page.params.diff_id,
			})
		});

		if (res.body) {
			const reader = res.body.getReader();
			const decoder = new TextDecoder();
			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				const text = decoder.decode(value);
				sseStream += text;
			}
		}

		generatedPatch = "--- a/schema.yaml\n+++ b/schema.yaml\n@@ -1,3 +1,3 @@\n-old\n+new";
		isGenerating = false;
	}

	function applyPatch() {
		patchApplied = true;
	}
</script>

<div class="diff-page">
	<div class="page-header">
		<h1 class="page-title">Schema Diff Viewer</h1>
		<p class="page-subtitle">Review the detailed exact schema changes.</p>
	</div>

	<div class="actions">
		<button class="btn btn-primary" onclick={generateAIPatch} disabled={isGenerating}>
			{isGenerating ? 'Generating...' : 'Generate AI Patch'}
		</button>

		{#if generatedPatch}
			<button class="btn btn-success" onclick={applyPatch} disabled={patchApplied}>
				{patchApplied ? 'Applied' : 'Apply Patch'}
			</button>
		{/if}
	</div>

	{#if sseStream}
		<div class="sse-stream">
			<pre>{sseStream}</pre>
		</div>
	{/if}

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
	.actions {
		margin-bottom: 20px;
		display: flex;
		gap: 10px;
	}
	.btn {
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
		font-weight: bold;
		border: none;
	}
	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.btn-primary {
		background: #007bff;
		color: white;
	}
	.btn-success {
		background: #28a745;
		color: white;
	}
	.sse-stream {
		background: #f8f9fa;
		padding: 15px;
		border-radius: 4px;
		margin-bottom: 20px;
		border: 1px solid #ddd;
	}
	.sse-stream pre {
		margin: 0;
		white-space: pre-wrap;
	}
</style>
