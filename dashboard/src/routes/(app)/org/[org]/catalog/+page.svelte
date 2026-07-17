<script lang="ts">
	let { data } = $props();
</script>

<div class="catalog-container">
	<header class="catalog-header">
		<h1 class="page-title">Service Catalog</h1>
		<p class="page-subtitle">Zero-config directory of all discovered services and their API documentation.</p>
	</header>

	<div class="table-container">
		<table class="catalog-table">
			<thead>
				<tr>
					<th>Service Name</th>
					<th>Protocol</th>
					<th>Owner</th>
					<th>Status</th>
				</tr>
			</thead>
			<tbody>
				{#if data.catalogData.length === 0}
					<tr>
						<td colspan="4" style="text-align: center; padding: 32px; color: var(--text-muted);">
							No services found in this organization.
						</td>
					</tr>
				{:else}
					{#each data.catalogData as service}
						<tr class="catalog-row">
							<td>
								<a href={`/org/${data.org}/catalog/${service.name}`} class="service-link">
									{service.name}
								</a>
							</td>
							<td>
								<span class="protocol-badge">{service.protocol}</span>
							</td>
							<td>{service.owner}</td>
							<td>
								<span class={`status-indicator ${service.status === 'BREAKING' ? 'breaking' : 'safe'}`}>
									{service.status}
								</span>
							</td>
						</tr>
					{/each}
				{/if}
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
		box-sizing: border-box;
	}

	.catalog-header {
		margin-bottom: 24px;
	}

	.page-title {
		font-size: 1.8rem;
		margin-bottom: 8px;
		color: var(--text-main);
	}

	.page-subtitle {
		color: var(--text-muted);
		font-size: 1rem;
	}

	.table-container {
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
		background-color: var(--bg-dark);
		color: var(--text-muted);
		padding: 12px 24px;
		font-weight: 500;
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		border-bottom: 1px solid var(--border);
	}

	.catalog-table td {
		padding: 16px 24px;
		border-bottom: 1px solid var(--border);
		color: var(--text-main);
		font-size: 0.95rem;
	}

	.catalog-row:hover {
		background-color: var(--bg-hover, #2D3240);
	}

	.catalog-row:last-child td {
		border-bottom: none;
	}

	.service-link {
		color: var(--accent);
		text-decoration: none;
		font-weight: 500;
	}

	.service-link:hover {
		text-decoration: underline;
	}

	.protocol-badge {
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 0.8rem;
		font-family: 'JetBrains Mono', monospace;
	}

	.status-indicator {
		display: inline-block;
		padding: 4px 8px;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.status-indicator.safe {
		background-color: rgba(34, 197, 94, 0.1);
		color: #4ade80;
		border: 1px solid rgba(34, 197, 94, 0.2);
	}

	.status-indicator.breaking {
		background-color: rgba(239, 68, 68, 0.1);
		color: #f87171;
		border: 1px solid rgba(239, 68, 68, 0.2);
	}
</style>
