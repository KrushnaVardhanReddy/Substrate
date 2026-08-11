// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { page } from '$app/stores';

	let { data } = $props();
	let isAdmin = $state(false);

	$effect(() => {
		if (typeof window !== 'undefined') {
			const token = localStorage.getItem('github_token');
			if (token) {
				try {
					const payload = JSON.parse(atob(token.split('.')[1]));
					const orgs: Record<string, string> = payload.orgs || {};
					const org = $page.params?.org;
					if (org && orgs[org] !== 'admin') {
						isAdmin = false;
					} else {
						isAdmin = true;
					}
				} catch (e) {
					isAdmin = false;
				}
			}
		}
	});

	function handleSave() {
		alert('Settings saved!');
	}
</script>

<div class="settings-page">
	<header class="page-header">
		<h1 class="page-title">Organization Settings</h1>
		<p class="page-description">Manage configuration for {$page.params.org}</p>
	</header>

	<div class="settings-content">
		{#if !isAdmin}
			<div class="alert-box warning">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
				<p>You do not have administrative privileges for this organization. Settings are read-only.</p>
			</div>
		{/if}

		<section class="settings-section">
			<h3>General</h3>
			<div class="form-group">
				<label for="org-name">Organization Name</label>
				<input type="text" id="org-name" value={$page.params.org} disabled class="input-field disabled" />
			</div>
			<div class="form-group">
				<label for="visibility">Default Repository Visibility</label>
				<select id="visibility" class="input-field" disabled={!isAdmin}>
					<option value="private">Private</option>
					<option value="internal">Internal</option>
					<option value="public">Public</option>
				</select>
			</div>
		</section>

		<section class="settings-section">
			<h3>Webhooks & Integrations</h3>
			<div class="form-group">
				<label for="webhook-url">Global Webhook URL</label>
				<input type="url" id="webhook-url" placeholder="https://api.example.com/webhook" class="input-field" disabled={!isAdmin} />
			</div>
			<div class="form-group flex-row">
				<input type="checkbox" id="slack-notify" disabled={!isAdmin} />
				<label for="slack-notify">Enable Slack Notifications on Breaking Changes</label>
			</div>
		</section>

		{#if isAdmin}
			<div class="form-actions">
				<button class="btn-primary" onclick={handleSave}>Save Changes</button>
			</div>
		{/if}
	</div>
</div>

<style>
	.settings-page {
		padding: 24px;
		max-width: 800px;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 32px;
	}

	.page-title {
		font-size: 1.8rem;
		color: var(--text-main);
		margin-bottom: 8px;
	}

	.page-description {
		color: var(--text-muted);
		font-size: 1rem;
	}

	.settings-content {
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.settings-section {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 24px;
	}

	.settings-section h3 {
		font-size: 1.2rem;
		margin-bottom: 20px;
		color: var(--text-main);
		border-bottom: 1px solid var(--border);
		padding-bottom: 12px;
	}

	.form-group {
		margin-bottom: 20px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.form-group.flex-row {
		flex-direction: row;
		align-items: center;
		gap: 12px;
	}

	.form-group label {
		color: var(--text-main);
		font-size: 0.95rem;
		font-weight: 500;
	}

	.input-field {
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		color: var(--text-main);
		padding: 10px 12px;
		border-radius: 4px;
		font-size: 0.95rem;
		width: 100%;
		box-sizing: border-box;
	}

	.input-field:focus {
		outline: none;
		border-color: var(--accent);
	}

	.input-field.disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
	}

	.btn-primary {
		background-color: var(--accent);
		color: white;
		border: none;
		padding: 10px 20px;
		border-radius: 4px;
		font-weight: 600;
		cursor: pointer;
		transition: opacity 0.2s;
	}

	.btn-primary:hover {
		opacity: 0.9;
	}

	.alert-box {
		padding: 16px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		gap: 12px;
		font-size: 0.95rem;
	}

	.alert-box.warning {
		background-color: rgba(255, 171, 64, 0.1);
		border: 1px solid rgba(255, 171, 64, 0.3);
		color: #FFAB40;
	}
</style>
