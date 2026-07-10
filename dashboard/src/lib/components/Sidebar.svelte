<script lang="ts">
	let { repos = [], org = '', pathname = '' }: { repos?: { id: string; name: string; full_name: string }[], org?: string, pathname?: string } = $props();
</script>

<div class="sidebar">
	<div class="sidebar-header">
		Substrate
	</div>
	<div class="sidebar-nav">
		{#if org}
			<a href="/org/{org}" class="nav-item {pathname === `/org/${org}` ? 'active' : ''}">Repositories</a>
			<a href="/org/{org}/graph" class="nav-item {pathname === `/org/${org}/graph` ? 'active' : ''}">Dependency Graph</a>
		{:else}
			<div class="nav-item active">Dashboard</div>
		{/if}
		<div class="nav-item">API Keys</div>
		<div class="nav-item">Settings</div>

		{#if repos && repos.length > 0}
		<div class="repo-list">
			<div class="repo-list-title">Repositories</div>
			{#each repos as repo}
				<a href="/org/{org}/repo/{repo.name}" class="nav-item {pathname.includes(`/repo/${repo.name}`) ? 'active' : ''} block" style="text-decoration:none;">{repo.full_name}</a>
			{/each}
		</div>
		{/if}
	</div>
</div>
