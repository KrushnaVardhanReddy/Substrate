<script lang="ts">
	import Book from '@lucide/svelte/icons/book';
	let { repos = [], org = '', pathname = '', isAdmin = true }: { repos?: { id: string; name: string; full_name: string }[], org?: string, pathname?: string, isAdmin?: boolean } = $props();
</script>

<div class="sidebar">
	<div class="sidebar-header">
		Substrate
	</div>
	<div class="sidebar-nav">
		{#if org}
			<a href="/org/{org}" class="nav-item {pathname === `/org/${org}` ? 'active' : ''}">Repositories</a>
			<a href="/org/{org}/graph" class="nav-item {pathname === `/org/${org}/graph` ? 'active' : ''}">Dependency Graph</a>
			<a href="/org/{org}/catalog" class="nav-item {pathname === `/org/${org}/catalog` ? 'active' : ''}">
				<Book size={16} class="mr-2 inline-block" />
				Catalog
			</a>
			<a href="/org/{org}/matrix" class="nav-item {pathname === `/org/${org}/matrix` ? 'active' : ''}">Compatibility Matrix</a>
		{:else}
			<div class="nav-item active">Dashboard</div>
		{/if}
		<a href="/playground" class="nav-item {pathname === '/playground' ? 'active' : ''}" style="text-decoration:none;">✨ AI Playground</a>
		<div class="nav-item">API Keys</div>
		{#if isAdmin}
			{#if org}
				<a href="/org/{org}/settings" class="nav-item {pathname === `/org/${org}/settings` ? 'active' : ''}">Settings</a>
			{:else}
				<div class="nav-item">Settings</div>
			{/if}
		{/if}

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
