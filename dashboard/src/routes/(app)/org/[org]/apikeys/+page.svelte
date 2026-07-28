<script lang="ts">
	import { page } from '$app/stores';
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
				<button class="btn-primary">Generate New Key</button>
			</div>

			<div class="empty-state">
				<p>No API keys generated yet.</p>
			</div>
		</section>
	</div>
</div>

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

	.btn-primary:hover {
		opacity: 0.9;
	}

	.empty-state {
		text-align: center;
		padding: 40px 20px;
		color: var(--text-muted);
		background-color: var(--bg-dark);
		border-radius: 8px;
		border: 1px dashed var(--border);
	}
</style>
