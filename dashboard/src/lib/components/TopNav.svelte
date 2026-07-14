<script lang="ts">
	import { env } from '$env/dynamic/public';
</script>

<header class="top-nav">
	<div class="breadcrumbs">
		Dashboard / Overview
	</div>
	<div class="user-profile">
		<span style="font-size: 0.9rem; color: var(--text-muted);">{env.PUBLIC_ORG_NAME || 'Demo Org'}</span>
		<div class="avatar">{env.PUBLIC_ORG_NAME ? env.PUBLIC_ORG_NAME.charAt(0).toUpperCase() : 'M'}</div>
	</div>
</header>
