<script lang="ts">
	import { env } from '$env/dynamic/public';
</script>

<div class="topnav">
	<div class="breadcrumbs">
		Dashboard / Overview
	</div>
	<div class="user-profile">
		<span>{env.PUBLIC_ORG_NAME}</span>
		<div class="avatar">{env.PUBLIC_ORG_NAME ? env.PUBLIC_ORG_NAME.charAt(0).toUpperCase() : 'M'}</div>
	</div>
</div>
