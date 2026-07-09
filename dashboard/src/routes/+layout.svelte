<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import TopNav from '$lib/components/TopNav.svelte';

	let { data, children } = $props();
</script>

<!-- Sidebar -->
<Sidebar repos={data.repos} />

<!-- Main Wrapper -->
<div class="main-wrapper">
	<!-- Top Navigation -->
	<TopNav />

	<!-- Content Area -->
	{@render children()}
</div>
