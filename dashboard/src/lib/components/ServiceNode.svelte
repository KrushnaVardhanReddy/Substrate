<script lang="ts">
	import { Handle, Position } from '@xyflow/svelte';
	import { AppWindow, Database, Smartphone, Server } from '@lucide/svelte';

	let { data } = $props<{ data: any }>();

	let nodeType = $derived(data.metadata?.type || data.type || 'service');
</script>

<div class="service-node-card {nodeType}" class:origin={data.isOrigin} class:affected={data.isAffected} class:faded={data.isFaded}>
	<Handle type="target" position={Position.Top} style="background: #555; width: 8px; height: 8px;" />

	<div class="header">
		<div class="icon-wrapper {nodeType}">
			{#if nodeType === 'frontend'}
				<AppWindow size={14} color="#8B5CF6" />
			{:else if nodeType === 'database'}
				<Database size={14} color="#06B6D4" />
			{:else if nodeType === 'mobile'}
				<Smartphone size={14} color="#F43F5E" />
			{:else}
				<Server size={14} color="#6366F1" />
			{/if}
		</div>
		<div class="status-indicator" class:breaking={data.status === 'BREAKING'} class:safe={data.status !== 'BREAKING'}></div>
		<span class="service-name">{data.label}</span>
	</div>

	<div class="badge-container">
		<span class="badge">{nodeType}</span>
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

	.icon-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border-radius: 6px;
	}
	.icon-wrapper.database {
		background: rgba(6, 182, 212, 0.15);
	}
	.icon-wrapper.frontend {
		background: rgba(139, 92, 246, 0.15);
	}
	.icon-wrapper.mobile {
		background: rgba(244, 63, 94, 0.15);
	}
	.icon-wrapper.service, .icon-wrapper.backend, .icon-wrapper.consumer, .icon-wrapper.provider {
		background: rgba(99, 102, 241, 0.15);
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
