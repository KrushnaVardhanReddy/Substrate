<script lang="ts">
	import { page } from '$app/state';
	let scanning = $state(false);
	let scanComplete = $state(false);
	let scanError = $state('');

	async function runScan() {
		scanning = true;
		scanComplete = false;
		scanError = '';
		try {
			const res = await fetch(`/api/v1/discovery/scan/${page.params.org}/${page.params.repo}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' }
			});
			if (res.ok) {
				scanComplete = true;
			} else {
				scanError = `Scan failed: ${res.status}`;
			}
		} catch (e) {
			scanError = 'Scan request failed. Please try again.';
		} finally {
			scanning = false;
		}
	}
</script>

<div class="page-content p-6 max-w-4xl mx-auto">
	<h1 class="page-title text-2xl font-bold mb-2">Discovery Scanners</h1>
	<p class="text-gray-500 mb-6">Scanning repository {page.params.repo} for automated dependency discovery.</p>

	<div class="card p-6 bg-white border rounded">
		<h2 class="text-lg font-semibold mb-4">Run Discovery Scan</h2>
		<p class="text-gray-600 mb-4">
			Manually trigger a full dependency scan (Env Vars, Packages, Terraform, Kafka, OTel) for this repository.
		</p>
		
		<button
			class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
			onclick={runScan}
			disabled={scanning}
			data-testid="run-full-scan-btn"
		>
			{scanning ? 'Scanning...' : 'Run Full Scan'}
		</button>

		{#if scanComplete}
			<div class="mt-4 p-4 bg-green-50 text-green-700 border border-green-200 rounded" data-testid="scan-success">
				Scan completed successfully. No new dependencies discovered.
			</div>
		{/if}

		{#if scanError}
			<div class="mt-4 p-4 bg-red-50 text-red-700 border border-red-200 rounded" data-testid="scan-error">
				{scanError}
			</div>
		{/if}
	</div>
</div>
