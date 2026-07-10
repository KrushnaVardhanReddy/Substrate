<script lang="ts">
	import { page } from '$app/stores';
	
	let { data } = $props();
	
	let selectedCell = $state<any>(null);

	function selectCell(provider: string, version: string, consumer: string, result: string) {
		selectedCell = { provider, version, consumer, result };
	}
</script>

<div class="matrix-page">
	<div class="page-header">
		<h1 class="page-title">Compatibility Matrix</h1>
		<p class="page-subtitle">Provider versions vs Consumer environments</p>
	</div>

	<div class="matrix-container card p-0">
		<table class="matrix-table">
			<thead>
				<tr>
					<th class="matrix-th empty-th"></th>
					{#each data.matrixData.consumers as consumer}
						<th class="matrix-th consumer-th">
							<div class="consumer-name">{consumer}</div>
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each data.matrixData.grid as providerGroup}
					<tr class="provider-row">
						<td colspan={data.matrixData.consumers.length + 1} class="provider-name-cell">
							{providerGroup.provider}
						</td>
					</tr>
					{#each providerGroup.versions as ver}
						<tr>
							<td class="matrix-td version-td">
								{ver.version}
							</td>
							{#each ver.results as result, idx}
								<!-- svelte-ignore a11y_click_events_have_key_events -->
								<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
								<td class="matrix-td cell-td" onclick={() => selectCell(providerGroup.provider, ver.version, data.matrixData.consumers[idx], result)}>
									{#if result === 'COMPATIBLE'}
										<div class="status-icon safe-icon" title="Compatible">
											<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"></polyline></svg>
										</div>
									{:else if result === 'INCOMPATIBLE'}
										<div class="status-icon danger-icon" title="Incompatible">
											<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
										</div>
									{:else}
										<div class="status-icon unknown-icon" title="Unknown">
											<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="5" y1="12" x2="19" y2="12"></line></svg>
										</div>
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				{/each}
			</tbody>
		</table>
	</div>

	{#if selectedCell}
	<aside class="cell-detail-panel">
		<header class="detail-header">
			<h3 class="detail-title">Check Details</h3>
			<button class="close-btn" aria-label="Close" onclick={() => selectedCell = null}>
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
			</button>
		</header>
		<div class="detail-content">
			<div class="meta-row">
				<span class="meta-label">Provider</span>
				<span class="meta-value">{selectedCell.provider} @ {selectedCell.version}</span>
			</div>
			<div class="meta-row">
				<span class="meta-label">Consumer</span>
				<span class="meta-value">{selectedCell.consumer}</span>
			</div>
			<div class="meta-row mt-4">
				<span class="meta-label">Status</span>
				<span class="meta-value {selectedCell.result === 'INCOMPATIBLE' ? 'error-text' : (selectedCell.result === 'COMPATIBLE' ? 'safe-text' : '')}">
					{selectedCell.result}
				</span>
			</div>
			
			{#if selectedCell.result === 'INCOMPATIBLE'}
				<div class="error-box mt-4">
					<strong>Breaking Change Detected:</strong>
					<p>Field `email` was removed from `User` object, but consumer expects it.</p>
				</div>
			{/if}
		</div>
	</aside>
	{/if}
</div>

<style>
	.matrix-page {
		padding: 24px;
		display: flex;
		flex-direction: column;
		height: 100%;
		position: relative;
	}

	.matrix-container {
		overflow-x: auto;
		flex-grow: 1;
	}

	.matrix-table {
		width: 100%;
		border-collapse: separate;
		border-spacing: 0;
		min-width: 800px;
	}

	.matrix-th {
		background-color: var(--bg-dark);
		border-bottom: 1px solid var(--border);
		border-right: 1px solid var(--border);
		padding: 16px;
		text-align: center;
		position: sticky;
		top: 0;
		z-index: 10;
	}

	.empty-th {
		border-right: none;
		position: sticky;
		left: 0;
		z-index: 20;
		background-color: var(--bg-dark);
		min-width: 200px;
	}

	.consumer-name {
		font-family: 'JetBrains Mono', monospace;
		font-size: 13px;
		color: var(--text-main);
	}

	.provider-row {
		background-color: rgba(59, 130, 246, 0.05);
	}

	.provider-name-cell {
		padding: 12px 16px;
		font-weight: bold;
		color: var(--accent);
		border-bottom: 1px solid var(--border);
		font-family: 'JetBrains Mono', monospace;
	}

	.matrix-td {
		border-bottom: 1px solid var(--border);
		border-right: 1px solid var(--border);
		padding: 12px;
	}

	.version-td {
		font-family: 'JetBrains Mono', monospace;
		font-size: 13px;
		color: var(--text-main);
		position: sticky;
		left: 0;
		background-color: var(--bg-card);
		z-index: 5;
		border-right: 1px solid var(--border);
	}

	.cell-td {
		text-align: center;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.cell-td:hover {
		background-color: var(--bg-hover);
	}

	.status-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 50%;
	}

	.safe-icon {
		background-color: rgba(34, 197, 94, 0.1);
		color: var(--safe);
	}

	.danger-icon {
		background-color: rgba(239, 68, 68, 0.1);
		color: var(--danger);
	}

	.unknown-icon {
		background-color: rgba(148, 163, 184, 0.1);
		color: var(--text-muted);
	}

	.cell-detail-panel {
		position: absolute;
		top: 24px;
		right: 24px;
		bottom: 24px;
		width: 350px;
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		border-radius: 8px;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
		display: flex;
		flex-direction: column;
		z-index: 30;
	}

	.detail-header {
		padding: 20px;
		border-bottom: 1px solid var(--border);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.detail-title {
		font-size: 1.2rem;
		margin: 0;
	}

	.close-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
	}

	.close-btn:hover {
		color: var(--text-main);
	}

	.detail-content {
		padding: 20px;
		flex-grow: 1;
		overflow-y: auto;
	}

	.meta-row {
		display: flex;
		flex-direction: column;
		margin-bottom: 12px;
	}

	.mt-4 { margin-top: 16px; }

	.meta-label {
		font-size: 11px;
		text-transform: uppercase;
		color: var(--text-muted);
		margin-bottom: 4px;
	}

	.meta-value {
		font-family: 'JetBrains Mono', monospace;
		font-size: 13px;
	}

	.error-text { color: var(--danger); font-weight: bold; }
	.safe-text { color: var(--safe); font-weight: bold; }

	.error-box {
		background-color: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.3);
		border-radius: 6px;
		padding: 12px;
		color: var(--text-main);
		font-size: 13px;
	}

	.error-box p {
		margin-top: 8px;
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		color: var(--danger);
	}
</style>
