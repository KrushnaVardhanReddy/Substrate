// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import * as yaml from 'js-yaml';

	let yamlContent = $state(
		'openapi: 3.0.0\ninfo:\n  title: Sample API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      summary: Returns a list of users.\n      responses:\n        "200":\n          description: A JSON array of user names\n'
	);

	let parseError = $state<string | null>(null);
	let endpoints = $state<any[]>([]);

	$effect(() => {
		try {
			const parsed = yaml.load(yamlContent);
			parseError = null;
			const extracted = [];

			if (parsed && typeof parsed === 'object' && 'paths' in (parsed as any)) {
				const paths = (parsed as any).paths;
				for (const [path, methods] of Object.entries(paths)) {
					if (methods && typeof methods === 'object') {
						for (const [method, details] of Object.entries(methods)) {
							extracted.push({
								path,
								method: method.toUpperCase(),
								summary: (details as any).summary || ''
							});
						}
					}
				}
			}
			endpoints = extracted;
		} catch (e: any) {
			parseError = e.message;
			endpoints = [];
		}
	});

	// Form states
	let newMethod = $state('GET');
	let newPath = $state('/new-endpoint');
	let newSummary = $state('A new endpoint');

	function addEndpoint() {
		try {
			let parsed: any = yaml.load(yamlContent);
			if (!parsed || typeof parsed !== 'object') {
				parsed = { openapi: '3.0.0', info: { title: 'API', version: '1.0.0' }, paths: {} };
			}
			if (!parsed.paths) {
				parsed.paths = {};
			}
			if (!parsed.paths[newPath]) {
				parsed.paths[newPath] = {};
			}

			const methodLower = newMethod.toLowerCase();
			parsed.paths[newPath][methodLower] = {
				summary: newSummary,
				responses: {
					"200": {
						description: "OK"
					}
				}
			};

			yamlContent = yaml.dump(parsed);

			// Reset form
			newPath = '/';
			newSummary = '';
		} catch (e) {
			console.error("Failed to add endpoint", e);
		}
	}

	function getMethodColor(method: string) {
		switch (method) {
			case 'GET': return '#FFAB40'; // Usually blue or green, fallback to safe
			case 'POST': return '#00BFA5';
			case 'PUT': return '#3B82F6';
			case 'DELETE': return '#FF5252';
			default: return '#9BA1B0';
		}
	}
</script>

<div class="flex flex-col h-full w-full p-8 box-border" style="height: calc(100vh - 64px); overflow: hidden;">
	<div class="page-header shrink-0 mb-6">
		<h1 class="page-title">Visual API Design Studio</h1>
		<p class="page-subtitle">Design your OpenAPI contracts visually or via code.</p>
	</div>

	<div class="flex gap-6 grow min-h-0 min-w-0">
		<!-- Left Pane: Editor -->
		<div class="card flex-1 flex flex-col m-0 min-w-0 h-full overflow-hidden">
			<h2 class="card-title shrink-0">OpenAPI YAML</h2>
			{#if parseError}
				<div class="text-[var(--danger)] text-sm mb-2 shrink-0">{parseError}</div>
			{/if}
			<textarea
				class="w-full grow p-4 font-mono text-sm bg-[#090A0F] text-[#F8F9FA] border border-[#2A2E3D] rounded resize-none focus:outline-none focus:border-[#00BFA5]"
				bind:value={yamlContent}
				spellcheck="false"
			></textarea>
		</div>

		<!-- Right Pane: Visual Studio -->
		<div class="card flex-1 flex flex-col m-0 min-w-0 h-full overflow-hidden">
			<h2 class="card-title shrink-0">Visual Designer</h2>

			<!-- Add Endpoint Form -->
			<div class="bg-[#1D202D] p-4 rounded mb-4 shrink-0 border border-[#2A2E3D]">
				<h3 class="text-sm font-medium mb-3 text-[#F8F9FA]">Add New Endpoint</h3>
				<div class="flex gap-2">
					<select bind:value={newMethod} class="bg-[#090A0F] text-[#F8F9FA] border border-[#2A2E3D] rounded px-2 py-1 text-sm focus:outline-none focus:border-[#00BFA5]">
						<option>GET</option>
						<option>POST</option>
						<option>PUT</option>
						<option>DELETE</option>
						<option>PATCH</option>
					</select>
					<input
						type="text"
						bind:value={newPath}
						placeholder="/path"
						class="flex-1 bg-[#090A0F] text-[#F8F9FA] border border-[#2A2E3D] rounded px-2 py-1 text-sm focus:outline-none focus:border-[#00BFA5]"
					/>
					<input
						type="text"
						bind:value={newSummary}
						placeholder="Summary"
						class="flex-1 bg-[#090A0F] text-[#F8F9FA] border border-[#2A2E3D] rounded px-2 py-1 text-sm focus:outline-none focus:border-[#00BFA5]"
					/>
					<button onclick={addEndpoint} class="btn m-0 py-1 px-3 text-sm shrink-0 whitespace-nowrap">Add</button>
				</div>
			</div>

			<!-- Endpoints List -->
			<div class="flex-1 overflow-y-auto">
				{#if endpoints.length === 0 && !parseError}
					<div class="empty-state h-full">
						<div class="empty-icon">🔌</div>
						<p>No endpoints found in the YAML.</p>
					</div>
				{:else}
					<div class="flex flex-col gap-2">
						{#each endpoints as endpoint}
							<div class="bg-[#090A0F] border border-[#2A2E3D] rounded p-3 flex items-center gap-3">
								<span
									class="text-xs font-bold px-2 py-1 rounded w-16 text-center"
									style="background-color: {getMethodColor(endpoint.method)}20; color: {getMethodColor(endpoint.method)}"
								>
									{endpoint.method}
								</span>
								<span class="font-mono text-sm text-[#F8F9FA] flex-1 truncate">{endpoint.path}</span>
								{#if endpoint.summary}
									<span class="text-sm text-[#9BA1B0] truncate max-w-[200px]">{endpoint.summary}</span>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
