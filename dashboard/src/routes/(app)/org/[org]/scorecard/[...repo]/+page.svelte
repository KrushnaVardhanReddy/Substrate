// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { env } from '$env/dynamic/public';

	let loading = $state(true);
	let scorecard = $state<any>(null);

	onMount(async () => {
		const org = $page.params.org;
		const repo = $page.params.repo;

		const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
		const token = env.PUBLIC_API_TOKEN || '';

		try {
			const res = await fetch(`${apiUrl}/api/v1/scorecard/${org}/${repo}`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});
			if (res.ok) {
				scorecard = await res.json();
			}
		} catch (e) {
			console.error('Failed to fetch scorecard details', e);
		} finally {
			loading = false;
		}
	});
</script>

<div class="p-6">
	<h1 class="text-2xl font-bold mb-4" data-testid="scorecard-title">Scorecard for {$page.params.repo}</h1>

	{#if loading}
		<p>Loading...</p>
	{:else if scorecard}
		<div class="border rounded p-4 mb-6">
			<div class="flex gap-4 items-center mb-4">
				<div class="text-4xl font-bold">{scorecard.grade}</div>
				<div class="text-xl">Score: {scorecard.total}/100</div>
			</div>

			<h2 class="text-xl font-bold mb-2">Breakdown</h2>
			<div class="bg-gray-100 p-4 rounded" data-testid="breakdown-chart">
				<pre>{JSON.stringify(scorecard.breakdown, null, 2)}</pre>
			</div>
		</div>
	{:else}
		<p>No scorecard available.</p>
	{/if}
</div>
