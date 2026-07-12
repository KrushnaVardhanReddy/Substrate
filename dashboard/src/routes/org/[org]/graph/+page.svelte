<script lang="ts">
	import { page } from '$app/stores';
	import cytoscape from 'cytoscape';
	import dagre from 'cytoscape-dagre';

	cytoscape.use(dagre);
	
	let { data } = $props();
	let selectedNode = $state<any>(null);
	let cyContainer: HTMLDivElement;

	function selectNode(node: any) {
		selectedNode = node;
	}
	
	$effect(() => {
		if (!cyContainer) return;

		const nodesMap = new Map();
		const elements: cytoscape.ElementDefinition[] = [];

		for (const edge of data.graphData) {
			const { provider, consumer, status } = edge;

			if (!nodesMap.has(provider)) {
				nodesMap.set(provider, true);
				elements.push({ data: { id: provider, label: provider, status: 'SAFE' } });
			}
			if (!nodesMap.has(consumer)) {
				nodesMap.set(consumer, true);
				elements.push({ data: { id: consumer, label: consumer, status: 'SAFE' } });
			}

			elements.push({
				data: {
					source: consumer,
					target: provider,
					status: status
				}
			});
		}

		for (const edge of data.graphData) {
			if (edge.status === 'BREAKING') {
				const providerNode = elements.find(e => e.data.id === edge.provider);
				if (providerNode) {
					providerNode.data.status = 'BREAKING';
				}
			}
		}

		const cy = cytoscape({
			container: cyContainer,
			elements: elements,
			style: [
				{
					selector: 'node',
					style: {
						'background-color': '#1e293b',
						'border-width': 2,
						'border-color': '#3b82f6',
						'color': '#f8fafc',
						'text-valign': 'center',
						'font-size': '12px',
						'font-family': 'monospace',
						'padding': '10px',
						'shape': 'round-rectangle',
						'label': 'data(label)'
					}
				},
				{
					selector: 'edge',
					style: {
						'width': 2,
						'line-color': '#334155',
						'target-arrow-color': '#334155',
						'target-arrow-shape': 'triangle',
						'curve-style': 'bezier'
					}
				},
				{
					selector: 'edge[status = "BREAKING"]',
					style: {
						'line-color': '#ef4444',
						'target-arrow-color': '#ef4444'
					}
				},
				{
					selector: 'node[status = "BREAKING"]',
					style: {
						'border-color': '#ef4444'
					}
				}
			],
			layout: {
				name: 'dagre',
				rankDir: 'LR',
				nodeSep: 50,
				rankSep: 150,
				fit: true,
				padding: 50
			} as cytoscape.LayoutOptions
		});

		cy.on('tap', 'node', (evt) => {
			selectNode({ name: evt.target.id(), version: "v1.0.0", status: evt.target.data('status') || "SAFE" });
		});

		return () => {
			cy.destroy();
		};
	});
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

		<div bind:this={cyContainer} style="width: 100%; height: 100%;"></div>
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
