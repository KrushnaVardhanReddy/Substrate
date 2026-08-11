// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	let { data } = $props();

	let impactData: any = $state(null);

	onMount(async () => {
		const { org, repo } = $page.params;
		try {
			const res = await fetch(`/api/v1/impact/${org}/${repo}`);
			if (res.ok) {
				impactData = await res.json();
			}
		} catch (e) {
			console.error(e);
		}
	});
</script>

<div class="page-header flex justify-between items-center">
	<div>
		<h1 class="page-title">{data.repoName}</h1>
		<p class="page-subtitle">Viewing schema definition for this repository.</p>
	</div>
	<div class="flex gap-2">
		{#if data.canDeploy !== undefined}
			<span class="badge" style="background: {data.canDeploy ? '#10B981' : '#EF4444'}; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold;">
				Can-Deploy
			</span>
			<span class="badge" style="background: {data.canRollback ? '#10B981' : '#EF4444'}; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold;">
				Can-Rollback
			</span>
		{/if}
	</div>
</div>

{#if impactData}
<div class="card mb-4">
	<h2>Deployment Status</h2>
	<div class="flex gap-4 mt-2">
		<div class="badge {impactData.can_deploy ? 'safe' : 'breaking'}">
			Can-Deploy: {impactData.can_deploy ? 'Yes' : 'No'}
		</div>
		<div class="badge {impactData.can_rollback ? 'safe' : 'breaking'}">
			Can-Rollback: {impactData.can_rollback ? 'Yes' : 'No'}
		</div>
	</div>
</div>
{/if}

<div class="card p-0 flex flex-col h-[600px] overflow-hidden">
	<div class="p-3 border-b border-[var(--border)] bg-[var(--bg-card)] flex items-center gap-2 font-mono text-sm text-[var(--text-muted)]">
		<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
			<polyline points="14 2 14 8 20 8"></polyline>
			<line x1="16" y1="13" x2="8" y2="13"></line>
			<line x1="16" y1="17" x2="8" y2="17"></line>
			<polyline points="10 9 9 9 8 9"></polyline>
		</svg>
		schema.yaml
	</div>

	<div class="flex-grow overflow-auto bg-[#090A0F] p-4 text-[var(--text-main)] font-mono text-sm whitespace-pre">
		{data.schemaContent}
	</div>
</div>

<style>
	/* Make sure JetBrains Mono is applied specifically if available */
	.font-mono {
		font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
	}
	.justify-between { justify-content: space-between; }
	.flex-col { flex-direction: column; }
	.h-\[600px\] { height: 600px; }
	.overflow-hidden { overflow: hidden; }
	.overflow-auto { overflow: auto; }
	.p-3 { padding: 0.75rem; }
	.flex-grow { flex-grow: 1; }
	.bg-\[\#090A0F\] { background-color: #090A0F; }
	.whitespace-pre { white-space: pre; }
	.mb-4 { margin-bottom: 1rem; }
	.mt-2 { margin-top: 0.5rem; }
	.gap-4 { gap: 1rem; }
	.flex { display: flex; }

	.badge {
		padding: 4px 12px;
		border-radius: 12px;
		font-size: 0.85rem;
		font-weight: 600;
	}
	.badge.safe {
		background-color: rgba(16, 185, 129, 0.2);
		color: #10B981;
	}
	.badge.breaking {
		background-color: rgba(239, 68, 68, 0.2);
		color: #EF4444;
	}
</style>
