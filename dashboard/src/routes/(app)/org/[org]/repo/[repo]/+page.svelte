<script lang="ts">
	let { data } = $props();
</script>

<div class="page-header flex justify-between items-center">
	<div>
		<h1 class="page-title">{data.repoName}</h1>
		<p class="page-subtitle">Viewing schema definition for this repository.</p>
	</div>
</div>

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
</style>
