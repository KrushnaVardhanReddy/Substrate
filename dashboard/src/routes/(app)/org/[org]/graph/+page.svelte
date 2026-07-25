<script lang="ts">
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { onMount, untrack } from 'svelte';
	import cytoscape from 'cytoscape';
	import dagre from 'cytoscape-dagre';
	
	cytoscape.use(dagre);

	import { toPng } from 'html-to-image';
	import { Download, RotateCw } from 'lucide-svelte';
	import ServiceNode from '$lib/components/ServiceNode.svelte';
	import TeamGroupNode from '$lib/components/TeamGroupNode.svelte';
	import TimeTravelScrubber from '$lib/components/TimeTravelScrubber.svelte';
	import InteractiveEdge from '$lib/components/InteractiveEdge.svelte';

	let cyContainer: HTMLElement | undefined = $state();
	let cyInstance: cytoscape.Core | null = null;
	let { data }: { data: any } = $props();
	let rawNodes = $state<any[]>([]);
	let rawEdges = $state<any[]>([]);



	let showOnlyBreaking = $state(false);
	let hideOrphans = $state(false);
	let protocolFilter = $state('All');
	let searchQuery = $state('');
	let debouncedSearch = $state('');
	let includeNeighbors = $state(true);
	let layoutDirection = $state('BT'); // Bottom-to-Top puts databases at the bottom and frontends at the top
	let heatmapMode = $state(false);

	const toggleLayout = () => {
		const dirs = ['TB', 'BT', 'LR', 'RL'];
		const currentIdx = dirs.indexOf(layoutDirection);
		layoutDirection = dirs[(currentIdx + 1) % dirs.length];
	};

	$effect(() => {
		debouncedSearch = searchQuery;
	});

	let selectedNode: any = $state(null);

	let selectedDownstream = $derived(
		selectedNode ? rawEdges.filter(e => e.source === selectedNode.id).map(e => e.target) : []
	);
	let selectedUpstream = $derived(
		selectedNode ? rawEdges.filter(e => e.target === selectedNode.id).map(e => e.source) : []
	);

	let blastRadius = $derived.by(() => {
		if (!selectedNode) return { nodes: new Set<string>(), edges: new Set<string>() };

		const affectedNodes = new Set<string>();
		const affectedEdges = new Set<string>();
		
		// 1. Add immediate upstream providers so they are highlighted
		for (const edge of rawEdges) {
			if (edge.target === selectedNode.id) {
				affectedEdges.add(edge.id);
				affectedNodes.add(edge.source);
			}
		}

		// 2. Add downstream consumers recursively (Blast Radius)
		const queue = [selectedNode.id];

		while (queue.length > 0) {
			const current = queue.shift()!;
			// Find all edges where current node is the provider (source)
			// and consumers are the downstream dependencies (target)
			for (const edge of rawEdges) {
				if (edge.source === current) {
					affectedEdges.add(edge.id);
					if (!affectedNodes.has(edge.target)) {
						affectedNodes.add(edge.target);
						queue.push(edge.target);
					}
				}
			}
		}

		return { nodes: affectedNodes, edges: affectedEdges };
	});


	let isGraphEmpty = $derived(!debouncedSearch && !showOnlyBreaking && protocolFilter === 'All');

	let subsetData = $derived.by(() => {
		if (isGraphEmpty) return { nodes: [], edges: [] };

		let fNodes = rawNodes;
		
		if (showOnlyBreaking) {
			fNodes = fNodes.filter(n => n.data.status === 'BREAKING');
		}

		if (protocolFilter !== 'All') {
			fNodes = fNodes.filter(n => n.data.type === protocolFilter);
		}

		if (debouncedSearch) {
			const query = debouncedSearch.toLowerCase().trim();
			fNodes = fNodes.filter(n =>
				n.id.toLowerCase().includes(query) ||
				(typeof n.data?.label === 'string' && n.data.label.toLowerCase().includes(query))
			);
		}

		const matchedIds = new Set(fNodes.map(n => n.id));
		let fEdges: Edge[] = [];

		if (includeNeighbors) {
			const affectedIds = new Set(matchedIds);
			
			// 1. One layer Upstream (Dependencies)
			// Edge direction: source (provider) -> target (consumer)
			// If a matched node is a consumer (target), add its direct provider
			for (const edge of rawEdges) {
				if (matchedIds.has(edge.target)) {
					affectedIds.add(edge.source);
				}
			}

			// 2. Recursive Downstream (Blast Radius)
			// If a node is a provider (source), recursively add all its consumers (target)
			const queue = Array.from(matchedIds);
			while (queue.length > 0) {
				const current = queue.shift()!;
				for (const edge of rawEdges) {
					if (edge.source === current) {
						if (!affectedIds.has(edge.target)) {
							affectedIds.add(edge.target);
							queue.push(edge.target);
						}
					}
				}
			}

			fEdges = rawEdges.filter(e => affectedIds.has(e.source) && affectedIds.has(e.target));
			fNodes = rawNodes.filter(n => affectedIds.has(n.id));
		} else {
			fEdges = rawEdges.filter(e => matchedIds.has(e.source) && matchedIds.has(e.target));
		}

		if (hideOrphans) {
			const connectedIds = new Set<string>();
			fEdges.forEach(e => { connectedIds.add(e.source); connectedIds.add(e.target); });
			fNodes = fNodes.filter(n => connectedIds.has(n.id) || n.type === 'teamGroup'); // keep team nodes temporarily
		}

		// Ensure all parent nodes (teams) for the surviving children are included
		const requiredParentIds = new Set<string>();
		fNodes.forEach(n => {
			if (n.parentId) requiredParentIds.add(n.parentId);
		});

		// Add back missing parent nodes
		requiredParentIds.forEach(parentId => {
			if (!fNodes.some(n => n.id === parentId)) {
				const parentNode = rawNodes.find(n => n.id === parentId);
				if (parentNode) fNodes.push(parentNode);
			}
		});

		return { nodes: fNodes, edges: fEdges };
	});

	let displayData = $derived.by(() => {
		let dNodes = subsetData.nodes;
		let dEdges = subsetData.edges;

		// Inject heatmapMode into nodes before passing to Svelte Flow
		dNodes = dNodes.map(node => ({
			...node,
			data: { ...node.data, heatmapMode }
		}));

		// Apply blast radius highlighting and fading
		if (selectedNode) {
			dNodes = dNodes.map(n => ({
				...n,
				data: {
					...n.data,
					isOrigin: n.id === selectedNode.id,
					isAffected: blastRadius.nodes.has(n.id),
					isFaded: n.id !== selectedNode.id && !blastRadius.nodes.has(n.id)
				}
			}));

			dEdges = dEdges.map(e => ({
				...e,
				style: (blastRadius.edges.has(e.id) || e.source === selectedNode.id || e.target === selectedNode.id)
					? e.style
					: `${e.style || ""}; opacity: 0.2;`
			}));
		} else {
			dNodes = dNodes.map(n => ({
				...n,
				data: { ...n.data, isOrigin: false, isAffected: false, isFaded: false }
			}));
		}

		return { nodes: dNodes, edges: dEdges };
	});

	$effect(() => {
		if (!cyContainer) return;
		if (isGraphEmpty) {
			if (cyInstance) cyInstance.destroy();
			cyInstance = null;
			return;
		}

		const elements: cytoscape.ElementDefinition[] = [];

		for (const node of displayData.nodes) {
			elements.push({
				data: { id: node.id, label: node.data.label, status: node.data.status },
				classes: node.type
			});
		}

		for (const edge of displayData.edges) {
			elements.push({
				data: { source: edge.source, target: edge.target, status: edge.style?.includes('red') ? 'BREAKING' : 'SAFE' }
			});
		}

		untrack(() => {
			if (cyInstance) {
				cyInstance.destroy();
			}

			cyInstance = cytoscape({
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
							'text-halign': 'center',
							'font-size': '14px',
							'font-family': 'monospace',
							'padding': '15px',
							'shape': 'round-rectangle',
							'label': 'data(label)',
							'width': 'auto',
							'height': 'auto'
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
					},
					{
						selector: '.service',
						style: {
							'border-color': '#a855f7'
						}
					},
					{
						selector: '.teamGroup',
						style: {
							'background-opacity': 0.1,
							'border-style': 'dashed'
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
				}
			});

			if (browser) {
				(window as any).cyInstance = cyInstance;
			}
			cyInstance.on('tap', 'node', (evt) => {
				const nodeData = evt.target.data();
				const fullNode = rawNodes.find((n: any) => n.id === nodeData.id);
				if (fullNode) {
					// We construct an object matching what the detail panel expects
					selectedNode = {
						id: fullNode.id,
						label: fullNode.data.label,
						status: fullNode.data.status,
						metadata: fullNode.data.metadata,
						version: fullNode.data.version
					};
				}
			});
		});
	});

	const exportImage = () => {
		// Not implemented for cytoscape yet
	};

	const processGraphData = (edgesData: any) => {
		let newNodesMap = new Map<string, Node>();
		let newEdges: Edge[] = [];

		const addTeamNode = (teamName: string) => {
			if (!teamName) return;
			const teamId = `team-${teamName}`;
			if (!newNodesMap.has(teamId)) {
				newNodesMap.set(teamId, {
					id: teamId,
					type: 'teamGroup',
					position: { x: 0, y: 0 },
					data: { label: teamName }
				});
			}
		};

		const addNode = (id: string, status: string, type: string, metadata: any = {}) => {
			let teamName = metadata?.team || null;
			if (teamName) addTeamNode(teamName);

			if (!newNodesMap.has(id)) {
				const existingNode = rawNodes.find(n => n.id === id);
				const volatilityScore = existingNode?.data?.volatilityScore !== undefined
					? existingNode.data.volatilityScore
					: Math.floor(Math.random() * 101);

				const node: Node = {
					id,
					type: 'service',
					position: { x: 0, y: 0 },
					data: { label: id, status, type, metadata, volatilityScore }
				};
				if (teamName) {
					node.parentId = `team-${teamName}`;
					node.extent = 'parent';
				}
				newNodesMap.set(id, node);
			} else {
				const existing = newNodesMap.get(id);
				if (existing) {
					if (status === 'BREAKING') existing.data.status = 'BREAKING';
					if (metadata && Object.keys(metadata).length > 0) {
						existing.data.metadata = metadata;
					}
					if (teamName && !existing.parentId) {
						existing.parentId = `team-${teamName}`;
						existing.extent = 'parent';
					}
				}
			}
		};

		edgesData.forEach((edge: any) => {
			addNode(edge.provider, edge.status, 'provider', edge.provider_metadata);
			addNode(edge.consumer, 'SAFE', 'consumer', edge.consumer_metadata);

			newEdges.push({
				id: `e-${edge.provider}-${edge.consumer}`,
				source: edge.provider,
				target: edge.consumer,
				type: edgesData.length < 150 ? 'interactive' : 'straight',
				animated: true,
				style: `stroke: ${edge.status === 'BREAKING' ? '#EF4444' : '#64748b'}; stroke-width: 2px;`
			});
		});

		rawNodes = Array.from(newNodesMap.values()).sort((a, b) => {
			if (a.type === 'teamGroup' && b.type !== 'teamGroup') return -1;
			if (a.type !== 'teamGroup' && b.type === 'teamGroup') return 1;
			return 0;
		});
		rawEdges = newEdges;
		console.log('processGraphData finished. rawNodes length:', rawNodes.length);
	};

	let lastProcessedData: any = null;
	$effect(() => {
		const currentData = data?.graphData;
		if (currentData && currentData.length > 0 && currentData !== lastProcessedData) {
			lastProcessedData = currentData;
			untrack(() => {
				processGraphData(currentData);
			});
		}
	});

	// Run immediately with SSR/fallback data
	if (data && data.graphData) {
		processGraphData(data.graphData);
	}

	onMount(() => {
		console.log('Mounting component. data.graphData length:', data?.graphData?.length);

		const fetchInitialGraph = async () => {
			if ($page.params.org === 'stress-test') return;
			try {
				const token = localStorage.getItem('github_token');
				const headers: Record<string, string> = {};
				if (token) headers['Authorization'] = `Bearer ${token}`;
				const res = await fetch(`/api/v1/graph/${$page.params.org}`, { headers });
				if (res.ok) {
					const responseData = await res.json();
					processGraphData(Array.isArray(responseData) ? responseData : []);
				} else {
					console.error("Failed to fetch initial graph", res.status);
				}
			} catch (err) {
				console.error("Initial fetch error", err);
			}
		};

		let eventSource: EventSource | null = null;

		const initSSE = () => {
			if ($page.params.org === 'stress-test') return;

			eventSource = new EventSource('/api/v1/events');

			eventSource.onmessage = (event) => {
				if (event.data === 'heartbeat') return;
				try {
					const responseData = JSON.parse(event.data);
					processGraphData(Array.isArray(responseData) ? responseData : []);
				} catch (err) {
					console.error("Failed to parse SSE data", err);
				}
			};

			eventSource.onerror = (err) => {
				console.error("SSE connection error", err);
				eventSource?.close();
				setTimeout(initSSE, 5000);
			};
		};

		fetchInitialGraph().then(initSSE);

		return () => {
			if (eventSource) {
				eventSource.close();
			}
		};
	});
</script>

<div class="graph-container">
	<!-- Main Canvas: Dependency Graph -->
	<main class="main-canvas">
		<div class="canvas-header">
			<h1 class="page-title">Dependency Graph</h1>
			<p class="page-subtitle">Visualizing dependencies for {$page.params.org}</p>
			<div style="margin-top: 8px; padding: 4px 8px; background: rgba(59, 130, 246, 0.2); color: #60a5fa; border-radius: 4px; font-size: 12px; display: inline-block;">
				Debug: {rawNodes.length} nodes loaded
			</div>
		</div>

		<!-- Graph Controls overlay -->
		<div class="graph-controls" style="z-index: 20; display: flex; gap: 8px;">
			<div class="filter-panel">
				<div style="display: flex; gap: 8px;">
					<button class="btn-export" onclick={exportImage} title="Export PNG">
						<Download size={16} />
						Export
					</button>
					<button class="btn-export" onclick={toggleLayout} style="width: auto; padding: 8px;" title="Rotate Layout">
						<RotateCw size={16} />
					</button>
				</div>
				<label class="filter-label" style="display: inline-flex; align-items: center; gap: 8px; cursor: pointer; background-color: var(--bg-hover); padding: 6px 12px; border-radius: 4px; border: 1px solid var(--border); margin-bottom: 8px;">
					<input type="checkbox" bind:checked={heatmapMode} />
					<span style="font-weight: 500; color: var(--accent);">Heatmap Mode</span>
				</label>
				<label class="filter-label">
					<input type="checkbox" bind:checked={showOnlyBreaking} />
					Show Only BREAKING Changes
				</label>
				<label class="filter-label">
					<input type="checkbox" bind:checked={hideOrphans} />
					Hide Orphaned Nodes
				</label>
				<select bind:value={protocolFilter} class="filter-select">
					<option value="All">All Protocols</option>
					<option value="openapi">OpenAPI</option>
					<option value="graphql">GraphQL</option>
					<option value="protobuf">Protobuf</option>
					<option value="avro">Avro</option>
				</select>
				<input type="text" bind:value={searchQuery} placeholder="Search repository..." class="filter-input" />
				<label class="filter-label">
					<input type="checkbox" bind:checked={includeNeighbors} />
					Highlight connected neighbors
				</label>
			</div>
			{#if selectedNode}
				<button class="btn-clear-selection" onclick={() => selectedNode = null}>
					Clear Selection
				</button>
			{/if}
			<!-- Removed zoom controls since SvelteFlow provides its own <Controls /> -->
		</div>

		<div bind:this={cyContainer} style="flex: 1; width: 100%; min-height: 0; position: relative; z-index: 10;"></div>

		<!-- Time Travel Scrubber -->
		<div class="scrubber-wrapper">
			<TimeTravelScrubber />
		</div>
	</main>

	<!-- Side Panel (Detail View) -->
	{#if selectedNode}
	<aside class="detail-panel">
		<header class="detail-header">
			<div>
				<h3 class="detail-title">{selectedNode.label}</h3>
				<div class="detail-meta">
					<span class="meta-badge">{selectedNode.version || 'v1.0.0'}</span>
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

			<!-- Taxonomy Metadata -->
			{#if selectedNode?.metadata}
			<section class="detail-section">
				<h4 class="section-title">Taxonomy</h4>
				<div class="metadata-grid">
					{#if selectedNode.metadata.team}
						<div class="meta-label">Team</div>
						<div class="meta-value">{selectedNode.metadata.team}</div>
					{/if}
					{#if selectedNode.metadata.type}
						<div class="meta-label">Type</div>
						<div class="meta-value" style="text-transform: capitalize;">{selectedNode.metadata.type}</div>
					{/if}
					{#if selectedNode.metadata.databases && selectedNode.metadata.databases.length > 0}
						<div class="meta-label">Databases</div>
						<div class="meta-value">
							{#each selectedNode.metadata.databases as db}
								<span style="background: #2D3240; padding: 2px 6px; border-radius: 4px; font-size: 11px; margin-right: 4px;">{db}</span>
							{/each}
						</div>
					{/if}
				</div>
			</section>
			{/if}

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

			<!-- Impact Analysis -->
			<section class="detail-section">
				<h4 class="section-title">Impact Analysis</h4>
				<div class="impact-lists">
					<div class="impact-col">
						<div class="meta-label">Downstream Consumers</div>
						{#if selectedDownstream && selectedDownstream.length > 0}
							<ul class="impact-ul">
								{#each selectedDownstream as p}
									<li class="meta-value">{p}</li>
								{/each}
							</ul>
						{:else}
							<div class="meta-value">None</div>
						{/if}
					</div>
					<div class="impact-col" style="margin-top: 12px;">
						<div class="meta-label">Upstream Providers</div>
						{#if selectedUpstream && selectedUpstream.length > 0}
							<ul class="impact-ul">
								{#each selectedUpstream as p}
									<li class="meta-value">{p}</li>
								{/each}
							</ul>
						{:else}
							<div class="meta-value">None</div>
						{/if}
					</div>
				</div>
			</section>
		</div>

		<div class="detail-footer">
			<a href={`https://github.com/${selectedNode.id}/actions`} target="_blank" rel="noopener noreferrer" class="btn-secondary" style="display: flex; justify-content: center; align-items: center; text-decoration: none;">View Logs</a>
			<a href={`vscode://vscode.git/clone?url=https://github.com/${selectedNode.id}.git`} class="btn-primary" style="display: flex; justify-content: center; align-items: center; text-decoration: none;">Open in IDE</a>
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
		flex-direction: column;
		gap: 12px;
		z-index: 20;
		align-items: flex-end;
	}

	.filter-panel {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.filter-label {
		color: var(--text-main);
		font-size: 13px;
		display: flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
	}

	.filter-select, .filter-input {
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		color: var(--text-main);
		padding: 6px 8px;
		border-radius: 4px;
		font-size: 13px;
		width: 100%;
		box-sizing: border-box;
	}

	.scrubber-wrapper {
		position: absolute;
		bottom: 24px;
		left: 0;
		right: 0;
		z-index: 20;
		pointer-events: none;
		display: flex;
		justify-content: center;
		padding: 0 24px;
	}
	.btn-clear-selection {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		color: var(--text-main);
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 13px;
		font-weight: 500;
		box-shadow: 0 2px 4px rgba(0,0,0,0.2);
		transition: background-color 0.2s;
	}

	.btn-clear-selection:hover {
		background-color: var(--bg-hover, #2D3240);
	}

	.btn-export {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		color: var(--text-main);
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 13px;
		font-weight: 500;
		box-shadow: 0 2px 4px rgba(0,0,0,0.2);
		transition: background-color 0.2s;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		width: 100%;
	}

	.btn-export:hover {
		background-color: var(--bg-hover, #2D3240);
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

	.error-text {
		color: #ef4444;
	}

	.empty-state {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		display: flex;
		flex-direction: column;
		align-items: center;
		color: var(--text-muted, #94a3b8);
		text-align: center;
		z-index: 15;
		pointer-events: none;
	}
	.empty-state h2 {
		color: var(--text-main, #f8fafc);
		margin-top: 16px;
		margin-bottom: 8px;
		font-size: 1.2rem;
	}
</style>
