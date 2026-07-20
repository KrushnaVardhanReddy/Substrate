<script lang="ts">
	import { onMount } from 'svelte';
	// @ts-ignore
	import { PUBLIC_API_URL } from '$env/static/public';

	let plugins: any[] = $state([]);
	let loading = $state(true);
	let error: string | null = $state(null);
	let copiedPlugin: string | null = $state(null);

	onMount(async () => {
		try {
			const res = await fetch(`${PUBLIC_API_URL}/api/marketplace/plugins`);
			if (!res.ok) throw new Error('Failed to fetch plugins');
			plugins = await res.json();
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function copyInstallCommand(name: string) {
		const cmd = `substrate plugin install ${name}`;
		navigator.clipboard.writeText(cmd);
		copiedPlugin = name;
		setTimeout(() => { copiedPlugin = null; }, 2000);
	}
</script>

<div class="p-6">
	<h1 class="text-2xl font-bold mb-4">Marketplace Plugins</h1>
	<p class="text-gray-600 mb-6">Discover and install community rule packs for Substrate.</p>

	{#if loading}
		<div class="text-gray-500">Loading plugins...</div>
	{:else if error}
		<div class="text-red-500 bg-red-50 p-4 rounded-md">{error}</div>
	{:else if plugins.length === 0}
		<div class="text-gray-500">No plugins available yet. Be the first to publish one!</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each plugins as plugin}
				<div class="border rounded-lg p-5 shadow-sm bg-white hover:shadow-md transition-shadow">
					<h3 class="text-lg font-semibold text-gray-900 mb-2">{plugin.name}</h3>
					<p class="text-sm text-gray-600 mb-4 h-10 overflow-hidden text-ellipsis">{plugin.description}</p>

					<div class="flex items-center justify-between border-t pt-4">
						<code class="text-xs bg-gray-100 px-2 py-1 rounded text-gray-800 font-mono">
							substrate plugin install {plugin.name}
						</code>
						<button
							class="text-sm bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700 transition-colors ml-2 flex-shrink-0"
							onclick={() => copyInstallCommand(plugin.name)}
						>
							{copiedPlugin === plugin.name ? 'Copied!' : 'Copy'}
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
