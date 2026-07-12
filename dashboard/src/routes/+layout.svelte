<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import TopNav from '$lib/components/TopNav.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { page } from '$app/stores';

	let { data, children } = $props();
</script>

<!-- Sidebar -->
<Sidebar repos={data.repos} org={$page.params.org} pathname={$page.url.pathname} />

<!-- Main Wrapper -->
<div class="main-wrapper">
	<!-- Top Navigation -->
	<TopNav />

	<!-- Content Area -->
	{@render children()}

	<Footer />
</div>
