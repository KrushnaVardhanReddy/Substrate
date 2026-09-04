<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import DollarSign from 'lucide-svelte/icons/dollar-sign';

	let zombies: any[] = $state([]);
	let totalMonthlySavings = $state(0);
	let savingsPerEndpoint = $state(0);
	let loading = $state(true);
	let error = $state('');

	onMount(() => {
		const org = $page.params.org;
		if (!org) return;
		loading = true;
		error = '';
		fetch(`/api/v1/org/${org}/zombies`)
			.then(res => {
				if (!res.ok) throw new Error('Failed to fetch zombies');
				return res.json();
			})
			.then(data => {
				zombies = data.zombies || [];
				totalMonthlySavings = data.total_monthly_savings || 0;
				savingsPerEndpoint = data.savings_per_endpoint || 0;
				loading = false;
			})
			.catch(err => {
				error = err.message;
				loading = false;
			});
	});

	async function createPR(zombie: any) {
		try {
			const org = $page.params.org;
			const res = await fetch(`/api/v1/org/${org}/zombies/pr`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ path: zombie.path, method: zombie.method, repo_name: "unknown" })
			});
			if (res.ok) {
				alert('PR created');
			} else {
				alert('Failed to create PR');
			}
		} catch (err: any) {
			alert('Failed to create PR');
		}
	}
</script>

<div class="zombies-dashboard p-6">
	<h1 class="page-title text-2xl font-bold mb-4">Zombie APIs</h1>

	{#if loading}
		<p class="loading-state">Loading zombies...</p>
	{:else if error}
		<p class="error-state text-red-500">{error}</p>
	{:else}
		<div class="finops-banner bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded flex items-center gap-2 mb-6 shadow-sm">
			<DollarSign size={24} />
			<div>
				<p class="font-bold">Total Potential Monthly Savings</p>
				<p class="text-xl total-savings">${totalMonthlySavings}</p>
			</div>
		</div>

		{#if zombies.length === 0}
			<p class="empty-state">No zombie APIs found.</p>
		{:else}
			<div class="zombies-list flex flex-col gap-4">
				{#each zombies as zombie}
					<div class="zombie-card border p-4 rounded bg-white shadow flex justify-between items-center">
						<div>
							<h2 class="endpoint-name font-semibold">{zombie.method} {zombie.path}</h2>
							<p class="text-sm text-gray-500">Traffic: {zombie.request_count}</p>
							<p class="text-sm text-green-600 font-medium potential-savings">Potential Savings: ${savingsPerEndpoint}/mo</p>
						</div>
						<button
							class="pr-button bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
							onclick={() => createPR(zombie)}
						>
							Prune (Create PR)
						</button>
					</div>
				{/each}
			</div>
		{/if}
	{/if}
</div>
