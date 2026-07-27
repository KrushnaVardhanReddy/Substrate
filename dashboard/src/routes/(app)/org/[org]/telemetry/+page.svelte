<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	interface ROIMetrics {
		total_prevented_outages: number;
		total_undocumented_endpoints: number;
		hours_saved: number;
		estimated_dollar_value_saved: number;
	}

	let roiData: ROIMetrics | null = $state(null);
	let loadError = $state(false);

	onMount(async () => {
		try {
			const res = await fetch(`/api/v1/telemetry/roi/${page.params.org}`);
			if (res.ok) {
				roiData = await res.json();
			} else {
				loadError = true;
			}
		} catch (e) {
			console.error('Failed to load telemetry ROI data', e);
			loadError = true;
		}
	});
</script>

<div class="page-content">
	<h1 class="page-title">Telemetry &amp; ROI</h1>

	{#if roiData !== null}
		<div class="card mt-4" data-testid="roi-card">
			<h2>Return on Investment</h2>
			<p>Hours Saved: {roiData.hours_saved}</p>
			<p>Prevented Outages: {roiData.total_prevented_outages}</p>
			<p>Undocumented Endpoints Detected: {roiData.total_undocumented_endpoints}</p>
			<p>Estimated Value Saved: ${roiData.estimated_dollar_value_saved}</p>
		</div>
	{:else if loadError}
		<p class="error">Failed to load ROI data. Please try again.</p>
	{:else}
		<p>Loading ROI data...</p>
	{/if}
</div>
