// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { onMount, createEventDispatcher } from 'svelte';
	import { page } from '$app/stores';

	let { repos = [] }: { repos?: any[] } = $props();

	let visible = $state(false);
	let searchQuery = $state('');
	let selectedIndex = $state(0);

	let nodes = $derived(repos.map(r => ({ id: r.full_name, label: r.name, type: 'service' })).concat([
		{ id: 'settings', label: 'Settings', type: 'settings' }
	]));

	let filteredNodes = $derived(nodes.filter(node =>
		node.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
		node.label.toLowerCase().includes(searchQuery.toLowerCase())
	));

	const dispatch = createEventDispatcher();

	function toggle() {
		visible = !visible;
		if (visible) {
			searchQuery = '';
			selectedIndex = 0;
			setTimeout(() => {
				document.getElementById('cmd-palette-input')?.focus();
			}, 0);
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			toggle();
		}

		if (!visible) return;

		if (e.key === 'Escape') {
			visible = false;
		} else if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = (selectedIndex + 1) % filteredNodes.length;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = (selectedIndex - 1 + filteredNodes.length) % filteredNodes.length;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (filteredNodes.length > 0) {
				selectNode(filteredNodes[selectedIndex]);
			}
		}
	}

	function selectNode(node: any) {
		dispatch('select', node);
		visible = false;

		if (node.id === 'settings') {
			window.location.href = `/org/${$page.params.org}/settings`;
		} else {
			window.location.href = `/org/${$page.params.org}/repo/${node.label}`;
		}
	}

	onMount(() => {
		window.addEventListener('keydown', handleKeydown);
		return () => {
			window.removeEventListener('keydown', handleKeydown);
		};
	});
</script>

{#if visible}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="palette-overlay" onclick={() => visible = false}>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="palette-modal" onclick={(e) => e.stopPropagation()}>
			<input
				id="cmd-palette-input"
				type="text"
				bind:value={searchQuery}
				placeholder="Search for a microservice, team, or settings..."
				autocomplete="off"
			/>

			<ul class="results-list">
				{#each filteredNodes as node, i}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
					<li
						class="result-item {i === selectedIndex ? 'selected' : ''}"
						onmouseenter={() => selectedIndex = i}
						onclick={() => selectNode(node)}
					>
						{node.label} <span class="node-id">({node.id})</span>
					</li>
				{/each}
				{#if filteredNodes.length === 0}
					<li class="no-results">No results found</li>
				{/if}
			</ul>
		</div>
	</div>
{/if}

<style>
	.palette-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(15, 17, 23, 0.75);
		backdrop-filter: blur(8px);
		display: flex;
		justify-content: center;
		align-items: flex-start;
		padding-top: 15vh;
		z-index: 1000;
	}

	.palette-modal {
		background: rgba(30, 34, 44, 0.9);
		border: 1px solid #2A2E3D;
		border-radius: 12px;
		width: 600px;
		max-width: 90vw;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
		overflow: hidden;
	}

	#cmd-palette-input {
		width: 100%;
		padding: 16px 20px;
		background: transparent;
		border: none;
		border-bottom: 1px solid #2A2E3D;
		color: #F8F9FA;
		font-size: 1.1rem;
		outline: none;
		font-family: 'Inter', system-ui, sans-serif;
	}

	#cmd-palette-input::placeholder {
		color: #9BA1B0;
	}

	.results-list {
		list-style: none;
		margin: 0;
		padding: 8px;
		max-height: 400px;
		overflow-y: auto;
	}

	.result-item {
		padding: 12px 16px;
		border-radius: 8px;
		cursor: pointer;
		color: #9BA1B0;
		font-size: 0.95rem;
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.result-item.selected, .result-item:hover {
		background: #00BFA522;
		color: #00BFA5;
	}

	.node-id {
		font-size: 0.8rem;
		opacity: 0.6;
		font-family: monospace;
	}

	.no-results {
		padding: 20px;
		text-align: center;
		color: #9BA1B0;
	}
</style>
