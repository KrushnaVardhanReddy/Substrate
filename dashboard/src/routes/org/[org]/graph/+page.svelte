<script lang="ts">
	import { page } from '$app/stores';
	
	let { data } = $props();
	let selectedNode = $state<any>(null);

	function selectNode(node: any) {
		selectedNode = node;
	}
	
	// Separate upstream and downstream for simple rendering
	let providers = $derived(data.graphData.filter((e: any) => e.provider !== "api/gateway-service"));
	let consumers = $derived(data.graphData.filter((e: any) => e.consumer !== "api/gateway-service"));
	
	// We'll just hardcode a layout for the demo like the mockup did
</script>

<div class="graph-container">
	<!-- Main Canvas: Dependency Graph -->
	<main class="main-canvas">
		<div class="canvas-header">
			<h1 class="page-title">Dependency Graph</h1>
			<p class="page-subtitle">Visualizing dependencies for {$page.params.org}</p>
		</div>

		<!-- Graph Controls overlay -->
		<div class="graph-controls">
			<button class="icon-btn" aria-label="Refresh">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line><line x1="11" y1="8" x2="11" y2="14"></line><line x1="8" y1="11" x2="14" y2="11"></line></svg>
			</button>
			<button class="icon-btn" aria-label="Zoom Out">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line><line x1="8" y1="11" x2="14" y2="11"></line></svg>
			</button>
		</div>

		<!-- SVG for connections (absolute positioned behind nodes) -->
		<svg class="connections-layer" style="z-index: 0;">
			<!-- Paths connecting roughly where the nodes are placed below -->
			<path class="svg-connection" d="M 220 150 C 350 150, 350 300, 480 300"></path>
			<path class="svg-connection" d="M 220 300 C 350 300, 350 300, 480 300"></path>
			<path class="svg-connection svg-connection-error" d="M 220 450 C 350 450, 350 300, 480 300"></path>
			<path class="svg-connection" d="M 720 300 C 850 300, 850 200, 980 200"></path>
			<path class="svg-connection" d="M 720 300 C 850 300, 850 400, 980 400"></path>
		</svg>

		<!-- Graph Nodes Container -->
		<div class="nodes-layer" style="z-index: 10;">
			<!-- Upstream Column -->
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="node-card" style="left: 20px; top: 110px; width: 200px;" onclick={() => selectNode({ name: 'core/auth', version: 'v1.0.5', status: 'SAFE' })}>
				<div class="node-header">
					<span class="node-title">core/auth</span>
					<div class="status-badge safe">SAFE</div>
				</div>
				<div class="node-version">v1.0.5</div>
			</div>

			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="node-card" style="left: 20px; top: 260px; width: 200px;" onclick={() => selectNode({ name: 'db/postgres-driver', version: 'v3.2.1', status: 'SAFE' })}>
				<div class="node-header">
					<span class="node-title">db/postgres-driver</span>
					<div class="status-badge safe">SAFE</div>
				</div>
				<div class="node-version">v3.2.1</div>
			</div>

			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="node-card error-card" style="left: 20px; top: 410px; width: 200px;" onclick={() => selectNode({ name: 'utils/logger', version: 'v0.9.8', status: 'BREAKING' })}>
				<div class="node-header">
					<span class="node-title">utils/logger</span>
					<div class="status-badge error">BREAKING</div>
				</div>
				<div class="node-version error-text">v0.9.8 (Deprecated)</div>
			</div>

			<!-- Central Target Column -->
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="node-card selected-card" style="left: 480px; top: 260px; width: 240px; z-index: 20;" onclick={() => selectNode({ name: 'api/gateway-service', version: 'v4.1.0', status: 'SAFE', selected: true })}>
				<div class="node-header">
					<span class="node-title bold">api/gateway-service</span>
					<div class="status-badge safe">SAFE</div>
				</div>
				<div class="node-version mb-sm">v4.1.0</div>
				<div class="node-stats">
					<span class="stat-badge">3 In</span>
					<span class="stat-badge">2 Out</span>
				</div>
			</div>

			<!-- Downstream Column -->
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="node-card" style="left: 980px; top: 160px; width: 200px;" onclick={() => selectNode({ name: 'frontend/dashboard', version: 'v2.2.0', status: 'SAFE' })}>
				<div class="node-header">
					<span class="node-title">frontend/dashboard</span>
					<div class="status-badge safe">SAFE</div>
				</div>
				<div class="node-version">v2.2.0</div>
			</div>

			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="node-card" style="left: 980px; top: 360px; width: 200px;" onclick={() => selectNode({ name: 'workers/indexer', version: 'v1.0.1', status: 'SAFE' })}>
				<div class="node-header">
					<span class="node-title">workers/indexer</span>
					<div class="status-badge safe">SAFE</div>
				</div>
				<div class="node-version">v1.0.1</div>
			</div>
		</div>
	</main>

	<!-- Side Panel (Detail View) -->
	{#if selectedNode}
	<aside class="detail-panel">
		<header class="detail-header">
			<div>
				<h3 class="detail-title">{selectedNode.name}</h3>
				<div class="detail-meta">
					<span class="meta-badge">{selectedNode.version}</span>
					<span class="meta-time">Updated 2h ago</span>
				</div>
			</div>
			<button class="close-btn" aria-label="Close" onclick={() => selectedNode = null}>
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
			</button>
		</header>
		
		<div class="detail-content">
			<!-- Schema Definition -->
			<section class="detail-section">
				<h4 class="section-title">Schema Definition</h4>
				<div class="code-block">
<pre><code>type GatewayConfig &#123;
  port: Int!
  routes: [Route!]!
  rateLimit: RateLimitPolicy
  authProvider: AuthReference
&#125;</code></pre>
				</div>
			</section>

			<!-- Manifest Metadata -->
			<section class="detail-section">
				<h4 class="section-title">Manifest Metadata</h4>
				<div class="metadata-grid">
					<div class="meta-label">Maintainer</div>
					<div class="meta-value">Platform Team</div>
					<div class="meta-label">License</div>
					<div class="meta-value monospace">MIT</div>
					<div class="meta-label">Status</div>
					<div class="meta-value {selectedNode.status === 'BREAKING' ? 'error-text' : ''}">{selectedNode.status}</div>
				</div>
			</section>
		</div>

		<div class="detail-footer">
			<button class="btn-secondary">View Logs</button>
			<button class="btn-primary">Open in IDE</button>
		</div>
	</aside>
	{/if}
</div>

<style>
	.graph-container {
		display: flex;
		height: calc(100vh - 60px); /* 60px is topnav */
		width: 100%;
		overflow: hidden;
		position: relative;
		background-color: var(--bg-dark);
	}

	.main-canvas {
		flex-grow: 1;
		position: relative;
		padding: 24px;
		overflow: hidden;
	}

	.canvas-header {
		position: absolute;
		top: 24px;
		left: 24px;
		z-index: 20;
	}

	.page-title {
		font-size: 1.8rem;
		margin-bottom: 5px;
		color: var(--text-main);
	}

	.page-subtitle {
		color: var(--text-muted);
		font-size: 0.95rem;
	}

	.graph-controls {
		position: absolute;
		top: 24px;
		right: 24px;
		display: flex;
		gap: 8px;
		z-index: 20;
	}

	.icon-btn {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 4px;
		padding: 6px;
		color: var(--text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.icon-btn:hover {
		color: var(--text-main);
		background-color: var(--bg-hover);
	}

	.connections-layer {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
	}

	.svg-connection {
		fill: none;
		stroke: var(--border);
		stroke-width: 1.5;
		stroke-dasharray: 4;
		animation: dash 20s linear infinite;
	}

	.svg-connection-error {
		stroke: var(--danger);
		stroke-dasharray: none;
	}

	@keyframes dash {
		to { stroke-dashoffset: -100; }
	}

	.nodes-layer {
		position: relative;
		width: 100%;
		height: 100%;
		margin-top: 60px;
	}

	.node-card {
		position: absolute;
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 12px;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.node-card:hover {
		border-color: var(--accent);
		transform: translateY(-2px);
	}

	.node-card.selected-card {
		border-color: var(--accent);
		box-shadow: 0 0 15px rgba(59, 130, 246, 0.15);
	}

	.error-card {
		border-color: rgba(239, 68, 68, 0.5);
	}

	.node-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 8px;
	}

	.node-title {
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		color: var(--text-main);
	}

	.node-title.bold {
		font-weight: bold;
		font-size: 13px;
	}

	.status-badge {
		font-size: 10px;
		padding: 2px 6px;
		border-radius: 4px;
		font-weight: 600;
		letter-spacing: 0.05em;
	}

	.status-badge.safe {
		background-color: rgba(34, 197, 94, 0.1);
		color: var(--safe);
		border: 1px solid rgba(34, 197, 94, 0.2);
	}

	.status-badge.error {
		background-color: rgba(239, 68, 68, 0.1);
		color: var(--danger);
		border: 1px solid rgba(239, 68, 68, 0.2);
	}

	.node-version {
		color: var(--text-muted);
		font-family: 'JetBrains Mono', monospace;
		font-size: 11px;
	}

	.error-text {
		color: var(--danger);
	}

	.mb-sm {
		margin-bottom: 8px;
	}

	.node-stats {
		display: flex;
		gap: 8px;
	}

	.stat-badge {
		font-size: 10px;
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		padding: 2px 4px;
		border-radius: 4px;
		color: var(--text-muted);
	}

	/* Detail Panel */
	.detail-panel {
		width: 350px;
		border-left: 1px solid var(--border);
		background-color: var(--bg-card);
		display: flex;
		flex-direction: column;
		z-index: 30;
	}

	.detail-header {
		padding: 20px;
		border-bottom: 1px solid var(--border);
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
	}

	.detail-title {
		font-family: 'JetBrains Mono', monospace;
		font-size: 1.2rem;
		color: var(--text-main);
		margin-bottom: 6px;
	}

	.detail-meta {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.meta-badge {
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		color: var(--text-muted);
		padding: 2px 8px;
		border-radius: 4px;
		font-size: 11px;
		font-family: 'JetBrains Mono', monospace;
	}

	.meta-time {
		color: var(--text-muted);
		font-size: 12px;
	}

	.close-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
	}

	.close-btn:hover {
		color: var(--text-main);
	}

	.detail-content {
		flex-grow: 1;
		overflow-y: auto;
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.section-title {
		font-size: 11px;
		text-transform: uppercase;
		color: var(--text-muted);
		letter-spacing: 0.05em;
		margin-bottom: 12px;
	}

	.code-block {
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 12px;
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
		color: var(--text-main);
		overflow-x: auto;
	}

	.metadata-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		row-gap: 12px;
		font-size: 13px;
	}

	.meta-label {
		color: var(--text-muted);
	}

	.meta-value {
		color: var(--text-main);
	}

	.monospace {
		font-family: 'JetBrains Mono', monospace;
		font-size: 12px;
	}

	.detail-footer {
		padding: 20px;
		border-top: 1px solid var(--border);
		display: flex;
		gap: 12px;
	}

	.btn-secondary {
		flex: 1;
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		color: var(--text-main);
		padding: 8px;
		border-radius: 4px;
		cursor: pointer;
		font-weight: 500;
	}

	.btn-secondary:hover {
		background-color: var(--bg-hover);
	}

	.btn-primary {
		flex: 1;
		background-color: var(--accent);
		border: none;
		color: white;
		padding: 8px;
		border-radius: 4px;
		cursor: pointer;
		font-weight: 500;
	}

	.btn-primary:hover {
		opacity: 0.9;
	}
</style>
