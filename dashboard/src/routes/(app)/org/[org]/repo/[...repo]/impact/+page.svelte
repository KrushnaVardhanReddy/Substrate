// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';

    let impactData: any = null;
    let loading = true;
    let error: string | null = null;

    onMount(async () => {
        const { org, repo } = $page.params;
        try {
            const res = await fetch(`/api/v1/impact/${org}/${repo}`);
            if (res.ok) {
                impactData = await res.json();
            } else {
                error = 'Failed to fetch impact data';
            }
        } catch (e: any) {
            error = e.message;
        } finally {
            loading = false;
        }
    });
</script>

<div class="page-content">
    <h1 class="page-title">Impact Analysis</h1>

    {#if loading}
        <p>Loading...</p>
    {:else if error}
        <p class="error-text">{error}</p>
    {:else if impactData}
        <div class="card">
            <h2>Can-Deploy: {impactData.can_deploy ? 'Yes' : 'No'}</h2>
            <h2>Can-Rollback: {impactData.can_rollback ? 'Yes' : 'No'}</h2>
        </div>
    {/if}
</div>
