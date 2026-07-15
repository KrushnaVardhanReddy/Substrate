<script lang="ts">
	const dates = [
		"Jan 15, 2026", "Feb 28, 2026", "Mar 10, 2026", "Apr 05, 2026",
		"May 22, 2026", "Jun 18, 2026", "Jul 30, 2026", "Aug 12, 2026",
		"Sep 05, 2026", "Oct 24, 2026", "Nov 11, 2026", "Dec 01, 2026"
	];

	let isPlaying = $state(false);
	let sliderValue = $state(75);

	let percent = $derived(sliderValue);
	let dateIndex = $derived(Math.min(Math.floor((percent / 100) * dates.length), dates.length - 1));
	let currentDate = $derived(dates[dateIndex]);

	let thumbOffset = $derived((percent / 100) * 16 - 8);
	let tooltipLeft = $derived(`calc(${percent}% - ${thumbOffset}px)`);

	function togglePlay() {
		isPlaying = !isPlaying;
	}
</script>

<div class="glass-scrubber">
	<!-- Play/Pause Control -->
	<button class="play-pause-btn" aria-label={isPlaying ? "Pause" : "Play"} onclick={togglePlay}>
		{#if isPlaying}
			<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
				<path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" fill="currentColor"/>
			</svg>
		{:else}
			<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
				<path d="M8 5v14l11-7z" fill="currentColor"/>
			</svg>
		{/if}
	</button>

	<!-- Timeline Scrubber -->
	<div class="scrubber-container">
		<!-- Tooltip (Contextual Info) -->
		<div class="scrubber-tooltip" style="left: {tooltipLeft};">
			<span>{currentDate}</span>
		</div>

		<!-- Slider Container -->
		<div class="slider-container">
			<div class="tick-marks">
				<div class="tick-major"></div>
				<div class="tick-minor"></div>
				<div class="tick-minor"></div>
				<div class="tick-major"></div>
				<div class="tick-minor"></div>
				<div class="tick-minor"></div>
				<div class="tick-major"></div>
				<div class="tick-minor"></div>
				<div class="tick-minor"></div>
				<div class="tick-major"></div>
			</div>

			<!-- The actual slider -->
			<div class="slider-wrapper">
				<div class="progress-fill" style="width: {percent}%;"></div>
				<input type="range" id="timeline-slider" min="1" max="100" bind:value={sliderValue} />
			</div>
		</div>
	</div>

	<!-- Status Text -->
	<div class="status-text">
		<span class="status-dot"></span>
		<span class="status-label">Historical View</span>
	</div>
</div>

<style>
	.glass-scrubber {
		background-color: rgba(20, 22, 31, 0.8);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		border: 1px solid var(--border);
		border-radius: 9999px;
		padding: 12px 24px;
		display: flex;
		align-items: center;
		gap: 24px;
		box-shadow: 0 0 30px rgba(0,0,0,0.5);
		pointer-events: auto;
		width: 100%;
		max-width: 800px;
		margin: 0 auto;
	}

	.play-pause-btn {
		flex-shrink: 0;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background-color: var(--bg-card);
		display: flex;
		align-items: center;
		justify-content: center;
		border: 1px solid var(--border);
		cursor: pointer;
		transition: border-color 0.15s ease;
		color: var(--accent);
	}

	.play-pause-btn:hover {
		border-color: var(--accent);
	}

	.play-pause-btn svg {
		transition: transform 0.15s ease;
	}

	.play-pause-btn:hover svg {
		transform: scale(1.1);
	}

	.scrubber-container {
		flex-grow: 1;
		position: relative;
		display: flex;
		flex-direction: column;
		justify-content: center;
		height: 48px;
	}

	.scrubber-tooltip {
		position: absolute;
		top: -32px;
		transform: translateX(-50%);
		background-color: var(--bg-hover);
		border: 1px solid var(--border);
		padding: 4px 12px;
		border-radius: 4px;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
		pointer-events: none;
		transition: left 0.1s ease-out;
		z-index: 20;
		font-size: 12px;
		font-weight: 500;
		color: var(--accent);
		letter-spacing: 0.05em;
		white-space: nowrap;
	}

	.scrubber-tooltip::after {
		content: '';
		position: absolute;
		bottom: -5px;
		left: 50%;
		transform: translateX(-50%) rotate(45deg);
		width: 8px;
		height: 8px;
		background-color: var(--bg-hover);
		border-bottom: 1px solid var(--border);
		border-right: 1px solid var(--border);
	}

	.slider-container {
		position: relative;
		width: 100%;
		display: flex;
		align-items: center;
	}

	.tick-marks {
		position: absolute;
		width: 100%;
		display: flex;
		justify-content: space-between;
		padding: 0 4px;
		pointer-events: none;
		z-index: 0;
	}

	.tick-major {
		width: 2px;
		height: 8px;
		background-color: var(--border);
	}

	.tick-minor {
		width: 1px;
		height: 6px;
		background-color: var(--text-muted);
		opacity: 0.5;
	}

	.slider-wrapper {
		width: 100%;
		position: relative;
		z-index: 10;
	}

	.progress-fill {
		position: absolute;
		height: 4px;
		background: var(--accent);
		border-radius: 2px;
		top: 50%;
		transform: translateY(-50%);
		left: 0;
		pointer-events: none;
	}

	input[type=range] {
		-webkit-appearance: none;
		width: 100%;
		background: transparent;
		margin: 0;
	}

	input[type=range]::-webkit-slider-thumb {
		-webkit-appearance: none;
		height: 16px;
		width: 16px;
		border-radius: 50%;
		background: var(--accent);
		cursor: pointer;
		margin-top: -6px;
		box-shadow: 0 0 10px rgba(0, 191, 165, 0.5);
		transition: transform 0.1s ease;
	}

	input[type=range]::-webkit-slider-thumb:hover {
		transform: scale(1.2);
	}

	input[type=range]::-webkit-slider-runnable-track {
		width: 100%;
		height: 4px;
		cursor: pointer;
		background: var(--border);
		border-radius: 2px;
	}

	input[type=range]:focus {
		outline: none;
	}

	.status-text {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		gap: 8px;
		border-left: 1px solid var(--border);
		padding-left: 24px;
		padding-top: 4px;
		padding-bottom: 4px;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background-color: var(--text-muted);
	}

	.status-label {
		font-size: 12px;
		color: var(--text-muted);
		letter-spacing: 0.05em;
		text-transform: uppercase;
		font-weight: 500;
	}
</style>
