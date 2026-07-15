<script lang="ts">
	import { env } from '$env/dynamic/public';
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
		<button class="signout-btn" onclick={handleSignOut} aria-label="Sign Out">
			<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
				<polyline points="16 17 21 12 16 7"></polyline>
				<line x1="21" y1="12" x2="9" y2="12"></line>
			</svg>
		</button>
	</div>
</header>

<style>
	.signout-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 6px;
		border-radius: 4px;
		transition: background-color 0.2s ease, color 0.2s ease;
		margin-left: 8px;
	}
	.signout-btn:hover {
		background-color: var(--bg-hover);
		color: var(--danger);
	}
</style>
