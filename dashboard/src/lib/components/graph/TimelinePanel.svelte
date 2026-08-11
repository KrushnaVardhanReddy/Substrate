// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	let events: any[] = $state([]);
		let { selectedEventRepo = $bindable(), selectedEventType = $bindable() } = $props<{ selectedEventRepo?: string | null, selectedEventType?: string | null }>();

	onMount(async () => {
		const org = $page.params.org;
		try {
			const res = await fetch(`/api/v1/events/${org}`, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('local-dev-token') || 'test-token'}`
				}
			});
			if (res.ok) {
				events = await res.json();
			} else {
				console.error('Failed to fetch events', res.status);
			}
		} catch (e) {
			console.error('Error fetching events:', e);
		}
	});

		function selectEvent(repo: string | null, type: string | null) {
		selectedEventRepo = repo;
		selectedEventType = type;
	}
</script>

<div class="timeline-panel">
	<h3 class="timeline-title">Event Timeline</h3>
	<div class="events-list">
		{#if events && events.length > 0}
			{#each events as event}
				<div
					class="event-item {event.event_type}"
					onmouseenter={() => selectEvent(event.repo, event.event_type)}
					onmouseleave={() => selectEvent(null, null)}
					role="listitem"
				>
					<div class="event-time">{new Date(event.event_time).toLocaleString()}</div>
					<div class="event-details">
						<span class="event-badge {event.event_type}">{event.event_type}</span>
						<span class="event-repo">{event.repo}</span>
						<span class="event-desc">{event.description}</span>
					</div>
				</div>
			{/each}
		{:else}
			<div class="no-events">No events found in the selected timeframe.</div>
		{/if}
	</div>
</div>

<style>
	.timeline-panel {
		position: absolute;
		bottom: 24px;
		left: 50%;
		transform: translateX(-50%);
		width: 80%;
		max-width: 800px;
		background-color: var(--bg-card, #1E222C);
		border: 1px solid var(--border, #2D3240);
		border-radius: 8px;
		padding: 16px;
		z-index: 40;
		display: flex;
		flex-direction: column;
		box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.5);
		max-height: 200px;
	}

	.timeline-title {
		font-size: 14px;
		font-weight: 600;
		color: var(--text-main, #F8FAFC);
		margin-bottom: 12px;
		font-family: 'Inter', sans-serif;
	}

	.events-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		overflow-y: auto;
		flex: 1;
	}

	.event-item {
		display: flex;
		align-items: center;
		padding: 8px 12px;
		border-radius: 6px;
		background-color: var(--bg-dark, #0F172A);
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.event-item:hover {
		background-color: var(--bg-hover, #334155);
	}

	.event-item.deployment:hover {
		border-left: 3px solid #3B82F6;
	}

	.event-item.incident:hover {
		border-left: 3px solid #EF4444;
	}

	.event-time {
		font-size: 12px;
		color: var(--text-muted, #94A3B8);
		min-width: 150px;
		font-family: 'JetBrains Mono', monospace;
	}

	.event-details {
		display: flex;
		align-items: center;
		gap: 12px;
		flex: 1;
	}

	.event-badge {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 12px;
		text-transform: uppercase;
		font-weight: 600;
	}

	.event-badge.deployment {
		background-color: rgba(59, 130, 246, 0.2);
		color: #3B82F6;
	}

	.event-badge.incident {
		background-color: rgba(239, 68, 68, 0.2);
		color: #EF4444;
	}

	.event-repo {
		font-size: 13px;
		font-weight: 500;
		color: var(--text-main, #F8FAFC);
	}

	.event-desc {
		font-size: 13px;
		color: var(--text-muted, #94A3B8);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		flex: 1;
	}

	.no-events {
		color: var(--text-muted, #94A3B8);
		font-size: 13px;
		text-align: center;
		padding: 20px;
	}
</style>
