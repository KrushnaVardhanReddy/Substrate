<script lang="ts">
	import { onMount } from 'svelte';

	let roiData: any = $state(null);

	onMount(async () => {
		try {
			const res = await fetch('/api/v1/telemetry/roi');
			if (res.ok) {
				roiData = await res.json();
			}
		} catch (e) {
			console.error('Failed to load telemetry ROI data', e);
		}
	});
</script>

<div class="page-content">
	<h1 class="page-title">Telemetry & ROI</h1>

	{#if roiData}
		<div class="card mt-4">
			<h2>Return on Investment</h2>
			<p>Hours Saved: {roiData.hours_saved}</p>
			<p>Events Tracked: {roiData.events_tracked}</p>
		</div>
	{:else}
		<p>Loading ROI data...</p>
	{/if}
</div>
