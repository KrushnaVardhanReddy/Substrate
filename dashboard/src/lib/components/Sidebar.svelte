<script lang="ts">
	let { repos = [] }: { repos?: { id: string; name: string; full_name: string }[] } = $props();
</script>

<div class="sidebar">
	<div class="sidebar-header">
		Substrate
	</div>
	<div class="sidebar-nav">
		<div class="nav-item active">Dashboard</div>
		<div class="nav-item">API Keys</div>
		<div class="nav-item">Settings</div>

		<div class="repo-list">
			<div class="repo-list-title">Repositories</div>
			{#each repos as repo}
				<div class="nav-item">{repo.full_name}</div>
			{/each}
		</div>
	</div>
</div>
