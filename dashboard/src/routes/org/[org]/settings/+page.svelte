<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	let enforcing = $state(false);
	let loading = $state(false);
	let resultMessage = $state('');
	let isAdmin = $state(true); // Default true for mock/testing, but could be set to false if we strictly enforce check

	onMount(() => {
		// Mock logic: check for admin role in localStorage or derived from token scopes.
		const role = localStorage.getItem('github_role');
		// In a real application, you would verify OAuth Admin Scopes securely.
		if (role && role !== 'admin') {
			isAdmin = false;
		}
	});

	async function toggleEnforce() {
		loading = true;
		resultMessage = '';
		try {
			// Get token (simplified for this mockup - typically injected by load function)
			const token = localStorage.getItem('github_token') || 'dummy-token';

			const res = await fetch(`/api/v1/org/${$page.params.org}/enforce`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${token}`
				},
				body: JSON.stringify({ enforce: enforcing })
			});

			if (res.ok) {
				resultMessage = enforcing
					? 'Successfully enforced Substrate globally for the organization.'
					: 'Successfully removed global enforcement.';
			} else {
				resultMessage = `Failed to update global enforcement: ${res.statusText}`;
				enforcing = !enforcing; // revert state
			}
		} catch (e: any) {
			resultMessage = `Error: ${e.message}`;
			enforcing = !enforcing; // revert state
		} finally {
			loading = false;
		}
	}
</script>

<div class="page-header">
	<h1 class="page-title">Organization Settings</h1>
	<p class="page-subtitle">Manage organization-level Substrate configuration.</p>
</div>

{#if !isAdmin}
<div class="card p-6 max-w-2xl bg-red-900/10 border border-red-900/30">
	<h2 class="text-lg font-medium text-red-400 mb-2">Access Denied</h2>
	<p class="text-sm text-[var(--text-muted)]">
		You do not have the required GitHub Organization Admin permissions to modify these settings.
	</p>
</div>
{:else}
<div class="card p-6 max-w-2xl">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h2 class="text-lg font-medium text-[var(--text-main)] mb-1">Enforce Substrate Globally</h2>
			<p class="text-sm text-[var(--text-muted)]">
				When enabled, Substrate will be injected into Branch Protection Rules across all repositories, making it a required status check.
			</p>
		</div>
		<div class="ml-4 flex-shrink-0">
			<!-- A simple toggle switch UI -->
			<label class="relative inline-flex items-center cursor-pointer">
				<input type="checkbox" class="sr-only peer" bind:checked={enforcing} onchange={toggleEnforce} disabled={loading}>
				<div class="w-11 h-6 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
			</label>
		</div>
	</div>

	{#if resultMessage}
		<div class="mt-4 p-3 rounded text-sm {resultMessage.includes('Error') || resultMessage.includes('Failed') ? 'bg-red-900/30 text-red-400' : 'bg-green-900/30 text-green-400'}">
			{resultMessage}
		</div>
	{/if}
</div>
{/if}
