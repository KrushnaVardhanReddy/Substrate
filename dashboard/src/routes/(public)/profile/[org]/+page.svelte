// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	let orgName = $page.params.org;
	let profileData: any = null;
	let error: string | null = null;
	let loading = true;
	let copied = false;

	onMount(async () => {
		try {
			const res = await fetch(`http://localhost:8080/api/v1/public/profile/${orgName}`);
			if (!res.ok) {
				if (res.status === 404) {
					error = 'Organization not found or not eligible for public profile';
				} else {
					error = 'Failed to load profile data';
				}
				loading = false;
				return;
			}
			profileData = await res.json();
		} catch (err) {
			error = 'Failed to load profile data';
		} finally {
			loading = false;
		}
	});

	function copyUrl() {
		navigator.clipboard.writeText(window.location.href);
		copied = true;
		setTimeout(() => {
			copied = false;
		}, 2000);
	}
</script>

<svelte:head>
	<title>{orgName} API Profile - Substrate</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 flex flex-col items-center py-12 px-4 sm:px-6 lg:px-8">
	<div class="w-full max-w-4xl">
		<div class="flex justify-between items-center mb-8">
			<div>
				<h1 class="text-3xl font-bold text-gray-900">{orgName}</h1>
				<p class="text-sm text-gray-500 mt-1">Public API Reliability Profile</p>
			</div>
			<button
				onclick={copyUrl}
				class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
			>
				{#if copied}
					Copied!
				{:else}
					Copy URL
				{/if}
			</button>
		</div>

		{#if loading}
			<div class="flex justify-center p-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			</div>
		{:else if error}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
				<span class="block sm:inline">{error}</span>
			</div>
		{:else if profileData}
			<div class="bg-[var(--bg-card)] shadow overflow-hidden sm:rounded-lg border border-[var(--border)] mb-8">
				<div class="px-4 py-5 sm:px-6 border-b border-[var(--border)]">
					<h3 class="text-lg leading-6 font-medium text-[var(--text-main)]">90-Day API Health</h3>
					<p class="mt-1 max-w-2xl text-sm text-[var(--text-muted)]">Rolling reliability metrics for {orgName}'s APIs.</p>
				</div>
				<div class="px-4 py-5 sm:p-6 grid grid-cols-1 gap-6 sm:grid-cols-3">

					<!-- Breaking Changes -->
					<div class="bg-gray-50 p-4 rounded-lg">
						<h4 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">Breaking Changes</h4>
						<div class="flex items-end gap-2">
							<span class="text-3xl font-bold text-gray-900">0</span>
							<span class="text-sm text-gray-500 mb-1">avg/week</span>
						</div>
						<div class="mt-4 h-16 flex items-end gap-1">
							{#each profileData.metrics.breaking_changes as point, i}
								{#if i % 3 === 0} <!-- Downsample for simple viz -->
									<div
										class="w-full bg-blue-200 rounded-t-sm transition-all hover:bg-blue-300"
										style="height: {Math.max(4, point.value * 10)}%"
										title="{point.date}: {point.value}"
									></div>
								{/if}
							{/each}
						</div>
					</div>

					<!-- Blast Radius -->
					<div class="bg-gray-50 p-4 rounded-lg">
						<h4 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">Avg Blast Radius</h4>
						<div class="flex items-end gap-2">
							<span class="text-3xl font-bold text-gray-900">10%</span>
							<span class="text-sm text-gray-500 mb-1">consumers affected</span>
						</div>
						<div class="mt-4 h-16 flex items-end gap-1">
							{#each profileData.metrics.blast_radius as point, i}
								{#if i % 3 === 0}
									<div
										class="w-full bg-purple-200 rounded-t-sm transition-all hover:bg-purple-300"
										style="height: {point.value}%"
										title="{point.date}: {point.value}%"
									></div>
								{/if}
							{/each}
						</div>
					</div>

					<!-- Uptime -->
					<div class="bg-gray-50 p-4 rounded-lg">
						<h4 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">Uptime (estimated)</h4>
						<div class="flex items-end gap-2">
							<span class="text-3xl font-bold text-green-600">100%</span>
						</div>
						<div class="mt-4 h-16 flex items-end gap-1">
							{#each profileData.metrics.uptime as point, i}
								{#if i % 3 === 0}
									<div
										class="w-full bg-green-400 rounded-t-sm transition-all hover:bg-green-500"
										style="height: {point.value}%"
										title="{point.date}: {point.value}%"
									></div>
								{/if}
							{/each}
						</div>
					</div>

				</div>
			</div>

			<div class="text-center text-sm text-gray-500">
				Powered by <a href="https://substrate.io" class="text-blue-600 hover:underline">Substrate</a>
			</div>
		{/if}
	</div>
</div>
