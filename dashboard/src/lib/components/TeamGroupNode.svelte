<script lang="ts">
	import type { NodeProps } from '@xyflow/svelte';

	let { data, width, height }: NodeProps = $props();
</script>

<div class="team-group-node" style="width: {width}px; height: {height}px;">
	<div class="team-label">{data.label}</div>
</div>

<style>
	.team-group-node {
		background-color: rgba(30, 34, 44, 0.4);
		border: 2px dashed #4B5563;
		border-radius: 12px;
		position: relative;
		box-sizing: border-box;
		pointer-events: none;
	}

	.team-label {
		position: absolute;
		top: -24px;
		left: 0;
		color: #94A3B8;
		font-family: 'JetBrains Mono', monospace;
		font-size: 14px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		background: var(--bg-dark);
		padding: 2px 8px;
		border-radius: 4px;
	}
</style>
