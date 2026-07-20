<script lang="ts">
	import { BaseEdge, EdgeLabel, getBezierPath, type EdgeProps } from '@xyflow/svelte';

	let {
		id,
		sourceX,
		sourceY,
		targetX,
		targetY,
		sourcePosition,
		targetPosition,
		style,
		markerEnd
	}: EdgeProps = $props();

	let isHovered = $state(false);

	let pathParams = $derived(
		getBezierPath({
			sourceX,
			sourceY,
			sourcePosition,
			targetX,
			targetY,
			targetPosition
		})
	);

	let edgePath = $derived(pathParams[0]);
	let labelX = $derived(pathParams[1]);
	let labelY = $derived(pathParams[2]);

	let currentStyle = $derived(
		isHovered
			? `${style || ''} stroke-width: 4px; stroke: #3b82f6; filter: drop-shadow(0 0 4px rgba(59, 130, 246, 0.5)); transition: all 0.2s ease;`
			: `${style || ''} transition: all 0.2s ease;`
	);

	function onMouseEnter() {
		isHovered = true;
	}

	function onMouseLeave() {
		isHovered = false;
	}
</script>

<g
	role="presentation"
	onmouseenter={onMouseEnter}
	onmouseleave={onMouseLeave}
>
	<!-- Standard invisible interaction path to make hovering easier -->
	<path
		d={edgePath}
		fill="none"
		stroke-opacity={0}
		stroke-width={20}
		class="svelte-flow__edge-interaction"
	/>

	<!-- The visible edge path -->
	<BaseEdge {id} path={edgePath} style={currentStyle} {markerEnd} />

	<!-- Tooltip Label -->
	{#if isHovered}
		<EdgeLabel
			x={labelX}
			y={labelY}
			class="custom-tooltip-label"
		>
			<div class="tooltip-card">
				<div class="tooltip-row">
					<span class="tooltip-key">Protocol:</span>
					<span class="tooltip-value">gRPC</span>
				</div>
				<div class="tooltip-row">
					<span class="tooltip-key">Method:</span>
					<span class="tooltip-value">VerifyToken</span>
				</div>
			</div>
		</EdgeLabel>
	{/if}
</g>

<style>
	:global(.custom-tooltip-label) {
		pointer-events: none !important;
	}

	.tooltip-card {
		background-color: var(--bg-card, #1E222C);
		border: 1px solid var(--border, #2D3240);
		border-radius: 6px;
		padding: 8px 12px;
		color: #E2E8F0;
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
		display: flex;
		flex-direction: column;
		gap: 4px;
		white-space: nowrap;
		/* Ensure text is above everything */
		z-index: 50;
	}

	.tooltip-row {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.tooltip-key {
		color: var(--text-muted, #94A3B8);
		font-weight: 500;
	}

	.tooltip-value {
		color: var(--text-main, #F8FAFC);
		font-weight: 600;
	}
</style>
