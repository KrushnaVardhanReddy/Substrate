<script lang="ts">
	let telemetryEnabled = $state(true);

	$effect(() => {
		if (typeof window !== 'undefined') {
			const saved = localStorage.getItem('telemetryEnabled');
			if (saved !== null) {
				telemetryEnabled = saved === 'true';
			}
		}
	});

	function toggleTelemetry() {
		telemetryEnabled = !telemetryEnabled;
		if (typeof window !== 'undefined') {
			localStorage.setItem('telemetryEnabled', String(telemetryEnabled));
		}
	}
</script>

<div class="page-content" style="padding: 24px; max-width: 800px; margin: 0 auto;">
	<h1 class="page-title" style="font-size: 1.8rem; margin-bottom: 32px; color: var(--text-main);">User Settings</h1>

	<section class="settings-section" style="background-color: var(--bg-card); border: 1px solid var(--border); border-radius: 8px; padding: 24px;">
		<h3 style="font-size: 1.2rem; margin-bottom: 20px; color: var(--text-main); border-bottom: 1px solid var(--border); padding-bottom: 12px;">Privacy</h3>

		<div class="form-group" style="display: flex; flex-direction: row; align-items: center; gap: 12px; margin-bottom: 20px;">
			<input type="checkbox" id="telemetry-toggle" checked={telemetryEnabled} onchange={toggleTelemetry} />
			<label for="telemetry-toggle" style="color: var(--text-main); font-size: 0.95rem; font-weight: 500;">Enable Telemetry</label>
		</div>
		<p style="color: var(--text-muted); font-size: 0.9rem; margin-top: -12px; margin-left: 28px;">Help us improve Substrate by sending anonymous usage data.</p>
	</section>
</div>
