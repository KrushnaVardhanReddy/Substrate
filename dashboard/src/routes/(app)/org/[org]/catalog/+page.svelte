<script lang="ts">
	import type { PageData } from './$types';
	import { goto } from '$app/navigation';
	import { Search } from 'lucide-svelte';

	let { data }: { data: PageData } = $props();

	let searchQuery = $state('');

	let filteredNodes = $derived(
		data.nodes.filter((node: any) =>
			node.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			node.owner.toLowerCase().includes(searchQuery.toLowerCase()) ||
			node.protocol.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	function navigateToDetail(repoName: string) {
		goto(`/org/${data.org}/catalog/${repoName}`);
	}
</script>

<div class="catalog-container">
	<header class="catalog-header">
		<div>
			<h1 class="page-title">Service Catalog</h1>
			<p class="page-subtitle">Discover and explore APIs, schemas, and dependencies automatically inferred from your repositories.</p>
		</div>
		<div class="search-box">
			<Search class="search-icon" size={18} />
			<input type="text" placeholder="Search services..." bind:value={searchQuery} class="search-input" />
		</div>
	</header>

	<div class="table-wrapper">
		<table class="catalog-table">
			<thead>
				<tr>
					<th>Service Name</th>
					<th>Owner</th>
					<th>Protocol</th>
					<th>Status</th>
					<th>Upstream</th>
					<th>Downstream</th>
				</tr>
			</thead>
			<tbody>
				{#if filteredNodes.length === 0}
					<tr>
						<td colspan="6" class="empty-state">No services found matching your search.</td>
					</tr>
				{/if}
				{#each filteredNodes as node}
					<tr onclick={() => navigateToDetail(node.name)} class="clickable-row">
						<td class="font-mono font-medium text-main">{node.name}</td>
						<td><span class="badge badge-outline">{node.owner}</span></td>
						<td><span class="badge badge-gray">{node.protocol}</span></td>
						<td>
							{#if node.status === 'SAFE'}
								<span class="status-indicator safe">SAFE</span>
							{:else if node.status === 'BREAKING'}
								<span class="status-indicator breaking">BREAKING</span>
							{:else}
								<span class="status-indicator unknown">{node.status}</span>
							{/if}
						</td>
						<td>{node.upstreamCount}</td>
						<td>{node.downstreamCount}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

<style>
	.catalog-container {
		padding: 32px;
		max-width: 1200px;
		margin: 0 auto;
		width: 100%;
	}

	.catalog-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		margin-bottom: 32px;
		gap: 24px;
	}

	.page-title {
		font-size: 2rem;
		margin-bottom: 8px;
		color: var(--text-main);
	}

	.page-subtitle {
		color: var(--text-muted);
		font-size: 1.05rem;
		max-width: 600px;
	}

	.search-box {
		position: relative;
		width: 300px;
	}

	.search-icon {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		color: var(--text-muted);
	}

	.search-input {
		width: 100%;
		padding: 10px 12px 10px 40px;
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 6px;
		color: var(--text-main);
		font-size: 14px;
	}

	.search-input:focus {
		outline: none;
		border-color: var(--accent);
	}

	.table-wrapper {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 8px;
		overflow: hidden;
	}

	.catalog-table {
		width: 100%;
		border-collapse: collapse;
		text-align: left;
	}

	.catalog-table th {
		padding: 16px;
		background-color: var(--bg-dark);
		color: var(--text-muted);
		font-size: 12px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		border-bottom: 1px solid var(--border);
	}

	.catalog-table td {
		padding: 16px;
		border-bottom: 1px solid var(--border);
		color: var(--text-muted);
		font-size: 14px;
	}

	.catalog-table tr:last-child td {
		border-bottom: none;
	}

	.clickable-row {
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.clickable-row:hover {
		background-color: var(--bg-hover);
	}

	.font-mono {
		font-family: 'JetBrains Mono', monospace;
	}

	.font-medium {
		font-weight: 500;
	}

	.text-main {
		color: var(--text-main);
	}

	.badge {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 12px;
		font-weight: 500;
	}

	.badge-outline {
		border: 1px solid var(--border);
		color: var(--text-main);
	}

	.badge-gray {
		background-color: var(--bg-dark);
		color: var(--text-main);
	}

	.status-indicator {
		display: inline-block;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 12px;
		font-weight: 600;
	}

	.safe {
		background-color: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}

	.breaking {
		background-color: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}

	.unknown {
		background-color: var(--bg-dark);
		color: var(--text-muted);
	}

	.empty-state {
		text-align: center;
		padding: 48px !important;
		color: var(--text-muted);
	}
</style>
