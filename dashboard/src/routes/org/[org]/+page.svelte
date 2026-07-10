<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	let { data } = $props();
</script>

<div class="page-header">
	<h1 class="page-title">Repositories</h1>
	<p class="page-subtitle">Manage and view schemas for all registered repositories.</p>
</div>

{#if data.repos && data.repos.length > 0}
	<div class="card p-0">
		<table class="w-full text-left border-collapse">
			<thead>
				<tr>
					<th class="p-4 border-b border-[var(--border)] font-medium text-[var(--text-muted)] text-sm">Repository Name</th>
					<th class="p-4 border-b border-[var(--border)] font-medium text-[var(--text-muted)] text-sm">Schema Type</th>
					<th class="p-4 border-b border-[var(--border)] font-medium text-[var(--text-muted)] text-sm">Last Synced</th>
					<th class="p-4 border-b border-[var(--border)] font-medium text-[var(--text-muted)] text-sm">Status</th>
				</tr>
			</thead>
			<tbody>
				{#each data.repos as repo}
					<tr class="hover:bg-[var(--bg-hover)] transition-colors cursor-pointer" onclick={() => goto(`/org/${$page.params.org}/repo/${repo.name}`)}>
						<td class="p-4 border-b border-[var(--border)]">
							<div class="font-medium text-[var(--text-main)] flex items-center gap-2">
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-[var(--text-muted)]">
									<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
								</svg>
								{repo.full_name || repo.name}
							</div>
						</td>
						<td class="p-4 border-b border-[var(--border)]">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-[rgba(59,130,246,0.1)] text-[var(--accent)]">
								{repo.schema_type || 'Unknown'}
							</span>
						</td>
						<td class="p-4 border-b border-[var(--border)] text-sm text-[var(--text-muted)]">
							{repo.last_synced || 'Never'}
						</td>
						<td class="p-4 border-b border-[var(--border)]">
							<span class="inline-flex items-center gap-1.5 text-sm">
								<span class="w-2 h-2 rounded-full {repo.status === 'active' ? 'bg-[var(--safe)]' : 'bg-[var(--text-muted)]'}"></span>
								{repo.status || 'Unknown'}
							</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<div class="empty-state">
		<div class="empty-icon">
			<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
				<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
			</svg>
		</div>
		<h3>No repositories found</h3>
		<p>No repositories are currently registered for this organisation.</p>
	</div>
{/if}
