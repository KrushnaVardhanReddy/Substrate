<script lang="ts">
	import { onMount } from 'svelte';
	let activeTab = $state('webhooks');
	let celRule = $state('');
	let ruleSaved = $state(false);

	function saveRule() {
		ruleSaved = true;
		setTimeout(() => {
			ruleSaved = false;
		}, 3000);
	}
</script>

<div class="page-content p-6 max-w-4xl mx-auto">
	<h1 class="page-title text-2xl font-bold mb-6">Enterprise Dashboard</h1>

	<div class="tabs border-b border-gray-200 mb-6">
		<button
			class="px-4 py-2 mr-2 {activeTab === 'webhooks' ? 'border-b-2 border-blue-500 font-medium' : 'text-gray-500'}"
			onclick={() => activeTab = 'webhooks'}
		>
			Webhooks
		</button>
		<button
			class="px-4 py-2 mr-2 {activeTab === 'custom_rules' ? 'border-b-2 border-blue-500 font-medium' : 'text-gray-500'}"
			onclick={() => activeTab = 'custom_rules'}
		>
			Custom Rules
		</button>
		<button
			class="px-4 py-2 {activeTab === 'drift' ? 'border-b-2 border-blue-500 font-medium' : 'text-gray-500'}"
			onclick={() => activeTab = 'drift'}
		>
			Drift Detection
		</button>
	</div>

	{#if activeTab === 'webhooks'}
		<div class="tab-content">
			<h2 class="text-xl font-semibold mb-4">Webhooks</h2>
			<div class="p-4 bg-gray-50 border rounded">
				<p>Active Webhook: <strong>https://example.com/webhook</strong></p>
			</div>
		</div>
	{/if}

	{#if activeTab === 'custom_rules'}
		<div class="tab-content">
			<h2 class="text-xl font-semibold mb-4">Custom Rules (CEL)</h2>
			<textarea
				name="cel-rule"
				class="w-full h-32 p-3 border rounded mb-4"
				placeholder="Enter CEL rule..."
				bind:value={celRule}
			></textarea>
			<button
				class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
				onclick={saveRule}
			>
				Save Rule
			</button>
			{#if ruleSaved}
				<p class="text-green-600 mt-2">Rule saved successfully</p>
			{/if}
		</div>
	{/if}

	{#if activeTab === 'drift'}
		<div class="tab-content">
			<h2 class="text-xl font-semibold mb-4">Drift Detection</h2>
			<div class="p-4 border rounded border-red-200 bg-red-50">
				<h3 class="font-medium text-red-800">Drift Report</h3>
				<p class="text-red-600 mb-4">Drift detected for mcp-org/enterprise-repo</p>
				<button class="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700">
					Resolve
				</button>
			</div>
		</div>
	{/if}
</div>
