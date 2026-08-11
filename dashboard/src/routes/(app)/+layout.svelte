// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import '../../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import TopNav from '$lib/components/TopNav.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import CommandPalette from '$lib/components/CommandPalette.svelte';
	import CopilotWidget from '$lib/components/CopilotWidget.svelte';
	import { page } from '$app/stores';

	let { data, children } = $props();
	let isAdmin = $state(true);

	$effect(() => {
		if (typeof window !== 'undefined') {
			const token = localStorage.getItem('github_token');
			if (token) {
				try {
					const payload = JSON.parse(atob(token.split('.')[1]));
					const orgs = payload.orgs || {};
					const org = $page.params.org;
					isAdmin = !org || orgs[org] === 'admin';
				} catch (e) {
					isAdmin = false;
				}
			}
		}
	});
</script>

<CommandPalette repos={data.repos} />

<!-- Sidebar -->
<Sidebar repos={data.repos} org={$page.params.org} pathname={$page.url.pathname} isAdmin={isAdmin} />

<!-- Main Wrapper -->
<div class="main-wrapper">
	<!-- Top Navigation -->
	<TopNav />

	<!-- Content Area -->
	{@render children()}

	<Footer />
</div>

<CopilotWidget />
