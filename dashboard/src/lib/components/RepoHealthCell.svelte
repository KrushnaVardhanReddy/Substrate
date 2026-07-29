<script lang="ts">
	import { onMount } from 'svelte';
	import GradeBadge from './GradeBadge.svelte';
	import { env } from '$env/dynamic/public';
	import { page } from '$app/stores';

	let { org, repo } = $props<{ org: string, repo: string }>();

	let grade = $state<string | null>(null);
	let loading = $state(true);

	onMount(async () => {
		const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
		const token = env.PUBLIC_API_TOKEN || '';

		try {
			const res = await fetch(`${apiUrl}/api/v1/scorecard/${org}/${repo}`, {
				headers: {
					Authorization: `Bearer ${token}`
				}
			});
			if (res.ok) {
				const data = await res.json();
				grade = data.grade;
			}
		} catch (e) {
			console.error('Failed to fetch scorecard', e);
		} finally {
			loading = false;
		}
	});
</script>

{#if loading}
	<span class="text-sm text-gray-400">Loading...</span>
{:else if grade}
	<GradeBadge {grade} {org} {repo} />
{:else}
	<span class="text-sm text-gray-400">-</span>
{/if}
