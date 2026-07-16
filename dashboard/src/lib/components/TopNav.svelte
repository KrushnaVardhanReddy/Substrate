<script lang="ts">
	import { env } from '$env/dynamic/public';
	import { LogOut } from '@lucide/svelte';
	import { goto } from '$app/navigation';

	function handleSignOut() {
		localStorage.removeItem('github_token');
		goto('/onboarding');
	}
</script>

<header class="top-nav">
	<div class="breadcrumbs">
		Dashboard / Overview
	</div>
	<div class="user-profile">
		<span style="font-size: 0.9rem; color: var(--text-muted);">{env.PUBLIC_ORG_NAME || 'Demo Org'}</span>
		<div class="avatar">{env.PUBLIC_ORG_NAME ? env.PUBLIC_ORG_NAME.charAt(0).toUpperCase() : 'M'}</div>
		<button class="sign-out-btn" onclick={handleSignOut} aria-label="Sign Out">
			<LogOut size={18} />
		</button>
	</div>
</header>

<style>
	.top-nav {
		height: 64px;
		background-color: var(--bg-card);
		border-bottom: 1px solid var(--border);
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 32px;
	}

	.breadcrumbs {
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.user-profile {
		display: flex;
		align-items: center;
		gap: 12px;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.avatar {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background-color: var(--accent);
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		font-weight: bold;
	}

	.sign-out-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 4px;
		border-radius: 4px;
		transition: color 0.2s ease, background-color 0.2s ease;
	}

	.sign-out-btn:hover {
		color: var(--danger);
		background-color: rgba(255, 82, 82, 0.1);
	}
</style>