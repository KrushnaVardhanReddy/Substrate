<script lang="ts">
	import { Handle, Position, type NodeProps } from '@xyflow/svelte';

	let { data }: NodeProps = $props();
</script>

<div class="service-node-card" class:origin={data.isOrigin} class:affected={data.isAffected} class:faded={data.isFaded}>
	<Handle type="target" position={Position.Top} style="background: #555; width: 8px; height: 8px;" />

	<div class="header">
		<div class="status-indicator" class:breaking={data.status === 'BREAKING'} class:safe={data.status !== 'BREAKING'}></div>
		<span class="service-name">{data.label}</span>
	</div>

	<div class="badge-container">
		<span class="badge">{data.type || 'Service'}</span>
	</div>

	<Handle type="source" position={Position.Bottom} style="background: #555; width: 8px; height: 8px;" />
</div>

<style>
	.service-node-card {
		background-color: #1E222C;
		border: 1px solid #2D3240;
		border-radius: 8px;
		padding: 12px;
		min-width: 150px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		color: white;
		font-family: 'Inter', sans-serif;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
	}

	.header {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.status-indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.status-indicator.safe {
		background-color: #10B981; /* Green */
		box-shadow: 0 0 4px #10B981;
	}

	.status-indicator.breaking {
		background-color: #EF4444; /* Red */
		box-shadow: 0 0 4px #EF4444;
	}

	.service-name {
		font-size: 14px;
		font-weight: 600;
		color: #E2E8F0;
	}

	.badge-container {
		display: flex;
	}

	.badge {
		background-color: #2D3240;
		color: #94A3B8;
		font-size: 10px;
		padding: 2px 6px;
		border-radius: 4px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		font-weight: 500;
	}

	.service-node-card.origin {
		border-color: var(--danger, #EF4444);
		box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
	}

	.service-node-card.affected {
		border-color: var(--safe, #F59E0B);
		box-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
	}

	.service-node-card.faded {
		opacity: 0.2;
	}
</style>
