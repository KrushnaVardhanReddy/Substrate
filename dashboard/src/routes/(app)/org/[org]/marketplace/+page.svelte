// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { onMount } from 'svelte';
	import { env } from '$env/dynamic/public';
	import SubstrateCertifiedBadge from '$lib/components/SubstrateCertifiedBadge.svelte';
	let plugins: any[] = $state([]);
	let loading = $state(true);
	let error: string | null = $state(null);
	let copiedPlugin: string | null = $state(null);

	onMount(async () => {
		try {
			const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
			const res = await fetch(`${apiUrl}/api/marketplace/plugins`);
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
				<div class="border rounded-lg p-5 shadow-sm bg-white hover:shadow-md transition-shadow flex flex-col">
					<div class="flex items-start justify-between mb-2">
						<h3 class="text-lg font-semibold text-gray-900">{plugin.name}</h3>
						<!-- For now, we simulate that vendors named 'Kong', 'AWS API Gateway', 'Apigee', 'Cloudflare' are certified -->
						{#if ['Kong', 'AWS API Gateway', 'Apigee', 'Cloudflare'].some(v => plugin.name.includes(v) || plugin.description.includes(v))}
							<SubstrateCertifiedBadge />
						{/if}
					</div>
					<p class="text-sm text-gray-600 mb-4 flex-grow overflow-hidden text-ellipsis">{plugin.description}</p>

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
