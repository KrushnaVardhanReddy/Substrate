// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import type { PageData } from './$types';
	import { env } from '$env/dynamic/public';

	let { data }: { data: PageData } = $props();

	let items = $state(data.items || []);

	$effect(() => {
		items = data.items || [];
	});

	async function resolveItem(id: number, status: 'approved' | 'rejected') {
		try {
			const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/mcp/hitl-queue/${id}/resolve`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ status })
			});

			if (res.ok) {
				items = items.map((item: any) => item.id === id ? { ...item, status } : item);
			} else {
				console.error('Failed to resolve HITL item');
			}
		} catch (error) {
			console.error('Error resolving HITL item:', error);
		}
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-semibold text-white">HITL Queue</h1>
		<p class="mt-2 text-sm text-gray-400">
			Review and approve or reject pending agent tool calls that require human intervention.
		</p>
	</div>

	{#if items.length === 0}
		<div class="rounded-lg border border-gray-700 bg-gray-800 p-8 text-center text-gray-400">
			No pending items in the HITL queue.
		</div>
	{:else}
		<div class="overflow-hidden rounded-lg border border-gray-700 bg-gray-800 shadow-sm">
			<table class="min-w-full divide-y divide-gray-700">
				<thead class="bg-gray-900">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-400 w-1/4">Tool Name</th>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-400 w-1/2">Arguments</th>
						<th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-400">Status</th>
						<th class="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-400">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-700 bg-gray-800">
					{#each items as item}
						<tr>
							<td class="px-6 py-4 text-sm font-medium text-white break-words">
								{item.tool_name}
							</td>
							<td class="px-6 py-4 text-sm text-gray-400">
								<pre class="max-h-32 max-w-xs overflow-auto rounded bg-gray-900 p-2 text-xs text-gray-300">{JSON.stringify(item.arguments, null, 2)}</pre>
							</td>
							<td class="whitespace-nowrap px-6 py-4 text-sm">
								{#if item.status === 'pending'}
									<span class="inline-flex items-center rounded-full bg-yellow-900 px-2.5 py-0.5 text-xs font-medium text-yellow-300">
										Pending
									</span>
								{:else if item.status === 'approved'}
									<span class="inline-flex items-center rounded-full bg-green-900 px-2.5 py-0.5 text-xs font-medium text-green-300">
										Approved
									</span>
								{:else if item.status === 'rejected'}
									<span class="inline-flex items-center rounded-full bg-red-900 px-2.5 py-0.5 text-xs font-medium text-red-300">
										Rejected
									</span>
								{/if}
							</td>
							<td class="whitespace-nowrap px-6 py-4 text-right text-sm font-medium">
								{#if item.status === 'pending'}
									<button
										onclick={() => resolveItem(item.id, 'approved')}
										class="mr-2 inline-flex items-center rounded border border-transparent bg-green-600 px-3 py-1.5 text-xs font-medium text-white shadow-sm hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
									>
										Approve
									</button>
									<button
										onclick={() => resolveItem(item.id, 'rejected')}
										class="inline-flex items-center rounded border border-gray-600 bg-gray-700 px-3 py-1.5 text-xs font-medium text-white shadow-sm hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
									>
										Reject
									</button>
								{:else}
									<span class="text-gray-500">Resolved</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
