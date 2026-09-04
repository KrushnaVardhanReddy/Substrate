<script lang="ts">
	import { page } from '$app/stores';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let keys = $state(data.keys);

	$effect(() => {
		keys = data.keys;
	});

	let showModal = $state(false);
	let newKeyName = $state('');
	let generatedKey = $state('');
	let isSubmitting = $state(false);

	async function createKey() {
		if (!newKeyName.trim()) return;

		isSubmitting = true;
		try {
			const res = await fetch(`/api/v1/org/${$page.params.org}/apikeys`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ name: newKeyName })
			});

			if (res.ok) {
				const result = await res.json();
				generatedKey = result.raw_token;
				keys = [result.key, ...keys];
				newKeyName = '';
			}
		} finally {
			isSubmitting = false;
		}
	}

	async function deleteKey(id: string) {
		const res = await fetch(`/api/v1/org/${$page.params.org}/apikeys/${id}`, {
			method: 'DELETE'
		});

		if (res.ok) {
			keys = keys.filter((k: any) => k.id !== id);
		}
	}

	function closeGeneratedKey() {
		generatedKey = '';
		showModal = false;
	}
</script>

<div class="apikeys-page">
	<header class="page-header">
		<h1 class="page-title">API Keys</h1>
		<p class="page-description">Manage API keys for {$page.params.org}</p>
	</header>

	<div class="apikeys-content">
		<section class="apikeys-section">
			<div class="section-header">
				<h3>Tokens</h3>
				<button class="btn-primary" onclick={() => showModal = true}>Generate New Key</button>
			</div>

			{#if keys.length === 0}
				<div class="empty-state">
					<p>No API keys generated yet.</p>
				</div>
			{:else}
				<table class="keys-table">
					<thead>
						<tr>
							<th>Name</th>
							<th>Prefix</th>
							<th>Created At</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each keys as key (key.id)}
							<tr>
								<td>{key.name}</td>
								<td><code>{key.prefix}</code></td>
								<td>{new Date(key.created_at).toLocaleDateString()}</td>
								<td>
									<button class="btn-danger" onclick={() => deleteKey(key.id)}>Revoke</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</section>
	</div>
</div>

{#if showModal}
	<div class="modal-backdrop">
		<div class="modal-content">
			{#if generatedKey}
				<h3>Your New API Key</h3>
				<p class="warning-text">Please copy this key now. You will not be able to see it again.</p>
				<div class="key-display">
					<code>{generatedKey}</code>
				</div>
				<button class="btn-primary full-width" onclick={closeGeneratedKey}>I have copied my key</button>
			{:else}
				<h3>Generate New API Key</h3>
				<div class="form-group">
					<label for="key-name">Key Name</label>
					<input
						id="key-name"
						type="text"
						bind:value={newKeyName}
						placeholder="e.g. CI/CD Pipeline"
						disabled={isSubmitting}
					/>
				</div>
				<div class="modal-actions">
					<button class="btn-secondary" onclick={() => { showModal = false; newKeyName = ''; }} disabled={isSubmitting}>Cancel</button>
					<button class="btn-primary" onclick={createKey} disabled={!newKeyName.trim() || isSubmitting}>Generate</button>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.apikeys-page {
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

	.apikeys-content {
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.apikeys-section {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 24px;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
		border-bottom: 1px solid var(--border);
		padding-bottom: 12px;
	}

	.section-header h3 {
		font-size: 1.2rem;
		color: var(--text-main);
		margin: 0;
	}

	.btn-primary {
		background-color: var(--accent);
		color: white;
		border: none;
		padding: 8px 16px;
		border-radius: 4px;
		font-weight: 600;
		cursor: pointer;
		transition: opacity 0.2s;
	}

	.btn-primary:hover:not(:disabled) {
		opacity: 0.9;
	}

	.btn-primary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: transparent;
		color: var(--text-main);
		border: 1px solid var(--border);
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
	}

	.btn-danger {
		background-color: #dc3545;
		color: white;
		border: none;
		padding: 6px 12px;
		border-radius: 4px;
		font-size: 0.85rem;
		cursor: pointer;
	}

	.btn-danger:hover {
		opacity: 0.9;
	}

	.full-width {
		width: 100%;
		margin-top: 16px;
	}

	.empty-state {
		text-align: center;
		padding: 40px 20px;
		color: var(--text-muted);
		background-color: var(--bg-dark);
		border-radius: 8px;
		border: 1px dashed var(--border);
	}

	.keys-table {
		width: 100%;
		border-collapse: collapse;
	}

	.keys-table th, .keys-table td {
		padding: 12px;
		text-align: left;
		border-bottom: 1px solid var(--border);
	}

	.keys-table th {
		color: var(--text-muted);
		font-weight: 500;
		font-size: 0.9rem;
	}

	.keys-table code {
		background-color: var(--bg-dark);
		padding: 4px 8px;
		border-radius: 4px;
		font-family: monospace;
	}

	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
	}

	.modal-content {
		background-color: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 24px;
		width: 100%;
		max-width: 400px;
	}

	.modal-content h3 {
		margin-top: 0;
		margin-bottom: 16px;
		color: var(--text-main);
	}

	.form-group {
		margin-bottom: 24px;
	}

	.form-group label {
		display: block;
		margin-bottom: 8px;
		color: var(--text-main);
		font-size: 0.9rem;
	}

	.form-group input {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--border);
		border-radius: 4px;
		background-color: var(--bg-dark);
		color: var(--text-main);
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
	}

	.warning-text {
		color: #e6a23c;
		font-size: 0.9rem;
		margin-bottom: 16px;
	}

	.key-display {
		background-color: var(--bg-dark);
		border: 1px solid var(--border);
		padding: 16px;
		border-radius: 4px;
		word-break: break-all;
	}

	.key-display code {
		font-family: monospace;
		font-size: 1.1rem;
		color: var(--text-main);
	}
</style>
