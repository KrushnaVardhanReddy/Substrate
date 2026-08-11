// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import Book from '@lucide/svelte/icons/book';
	let { repos = [], org = '', pathname = '', isAdmin = true }: { repos?: { id: string; name: string; full_name: string }[], org?: string, pathname?: string, isAdmin?: boolean } = $props();
</script>

<div class="sidebar">
	<div class="sidebar-header">
		Substrate
	</div>
	<div class="sidebar-nav">
		{#if org}
			<a href="/org/{org}" class="nav-item {pathname === `/org/${org}` ? 'active' : ''}">Repositories</a>
			<a href="/org/{org}/graph" class="nav-item {pathname === `/org/${org}/graph` ? 'active' : ''}">Dependency Graph</a>
			<a href="/org/{org}/catalog" class="nav-item {pathname === `/org/${org}/catalog` ? 'active' : ''}">
				<Book size={16} class="mr-2 inline-block" />
				Catalog
			</a>
			<a href="/org/{org}/matrix" class="nav-item {pathname === `/org/${org}/matrix` ? 'active' : ''}">Compatibility Matrix</a>
			<a href="/org/{org}/marketplace" class="nav-item {pathname === `/org/${org}/marketplace` ? 'active' : ''}">Marketplace</a>
			<a href="/org/{org}/playground" class="nav-item {pathname === `/org/${org}/playground` ? 'active' : ''}" style="text-decoration:none;">✨ AI Playground</a>
			<a href="/org/{org}/apikeys" class="nav-item {pathname === `/org/${org}/apikeys` ? 'active' : ''}">API Keys</a>
		{:else}
			<a href="/" class="nav-item {pathname === '/' ? 'active' : ''}">Dashboard</a>
		{/if}

		{#if isAdmin}
			{#if org}
				<a href="/org/{org}/settings" class="nav-item {pathname === `/org/${org}/settings` ? 'active' : ''}">Settings</a>
			{:else}
				<div class="nav-item">Settings</div>
			{/if}
		{/if}

	</div>
</div>
