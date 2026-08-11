// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
  import { page } from '$app/stores';

  let orgId = $page.params.org;

  let policy: any = $state(null);
  let claims: any[] = $state([]);
  let loading = $state(true);
  let error = $state('');

  let newClaimPR = $state('');
  let newClaimDate = $state('');
  let newClaimAmount = $state(0);

  let filingClaim = $state(false);
  let fileClaimError = $state('');

  let fetched = false;

  $effect(() => {
    if (!fetched) {
      fetched = true;
      fetchPolicyAndClaims();
    }
  });

  async function fetchPolicyAndClaims() {
    loading = true;
    error = '';

    try {
      const token = window.localStorage.getItem('substrate-token') || window.localStorage.getItem('auth_token') || 'local-dev-token';
      const headers = { 'Authorization': `Bearer ${token}` };

      const policyRes = await fetch(`/api/v1/org/${orgId}/insurance/policy`, { headers });
      if (policyRes.ok) {
        policy = await policyRes.json();
      } else if (policyRes.status !== 404) {
        error = 'Failed to load policy';
      }

      const claimsRes = await fetch(`/api/v1/org/${orgId}/insurance/claims`, { headers });
      if (claimsRes.ok) {
        claims = await claimsRes.json();
      } else {
        if (!error) error = 'Failed to load claims';
      }
    } catch (e: any) {
      error = e.message || 'An error occurred';
    } finally {
      loading = false;
    }
  }

  async function fileClaim() {
    if (!newClaimPR || !newClaimDate || newClaimAmount <= 0) {
      fileClaimError = 'Please fill out all fields correctly';
      return;
    }

    filingClaim = true;
    fileClaimError = '';

    try {
      const token = window.localStorage.getItem('substrate-token') || window.localStorage.getItem('auth_token') || 'local-dev-token';
      const headers = { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}` 
      };

      const res = await fetch(`/api/v1/org/${orgId}/insurance/claims`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          github_pr_url: newClaimPR,
          incident_date: new Date(newClaimDate).toISOString(),
          amount_cents: newClaimAmount * 100 // convert to cents
        })
      });

      if (res.ok) {
        const newClaim = await res.json();
        claims = [newClaim, ...claims];
        newClaimPR = '';
        newClaimDate = '';
        newClaimAmount = 0;
      } else {
        const data = await res.text();
        fileClaimError = data || 'Failed to file claim';
      }
    } catch (e: any) {
      fileClaimError = e.message || 'An error occurred';
    } finally {
      filingClaim = false;
    }
  }
</script>

<div class="p-6 max-w-4xl mx-auto">
  <h1 class="text-2xl font-bold mb-6">Schema Insurance</h1>

  {#if loading}
    <p>Loading...</p>
  {:else if error}
    <p class="text-red-500">{error}</p>
  {:else}
    <div class="bg-white p-6 rounded shadow mb-8">
      <h2 class="text-xl font-semibold mb-4">Your Policy</h2>
      {#if policy}
        <p><strong>Policy ID:</strong> {policy.id}</p>
        <p><strong>Limit:</strong> ${(policy.policy_limit_cents / 100).toFixed(2)}</p>
      {:else}
        <p>No active insurance policy found for this organization.</p>
      {/if}
    </div>

    <div class="bg-white p-6 rounded shadow mb-8">
      <h2 class="text-xl font-semibold mb-4">File a Claim</h2>
      {#if policy}
        {#if fileClaimError}
          <p class="text-red-500 mb-4">{fileClaimError}</p>
        {/if}
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1" for="prUrl">GitHub PR URL</label>
            <input id="prUrl" class="w-full border p-2 rounded" type="url" bind:value={newClaimPR} placeholder="https://github.com/owner/repo/pull/123" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1" for="incidentDate">Incident Date</label>
            <input id="incidentDate" class="w-full border p-2 rounded" type="date" bind:value={newClaimDate} />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1" for="amount">Amount ($)</label>
            <input id="amount" class="w-full border p-2 rounded" type="number" min="0" step="0.01" bind:value={newClaimAmount} />
          </div>
          <button
            class="bg-blue-600 text-white px-4 py-2 rounded disabled:opacity-50"
            onclick={fileClaim}
            disabled={filingClaim}>
            {filingClaim ? 'Filing...' : 'File Claim'}
          </button>
        </div>
      {:else}
        <p class="text-gray-500">You need an active policy to file a claim.</p>
      {/if}
    </div>

    <div class="bg-white p-6 rounded shadow">
      <h2 class="text-xl font-semibold mb-4">Claim History</h2>
      {#if claims.length === 0}
        <p>No claims filed yet.</p>
      {:else}
        <table class="w-full text-left">
          <thead>
            <tr>
              <th class="pb-2">Date</th>
              <th class="pb-2">PR URL</th>
              <th class="pb-2">Amount</th>
              <th class="pb-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {#each claims as claim}
              <tr class="border-t">
                <td class="py-2">{claim.incident_date ? new Date(claim.incident_date).toLocaleDateString() : ''}</td>
                <td class="py-2"><a href={claim.github_pr_url} target="_blank" class="text-blue-600 hover:underline">Link</a></td>
                <td class="py-2">${(claim.amount_cents / 100).toFixed(2)}</td>
                <td class="py-2">
                  <span class="px-2 py-1 rounded text-xs font-semibold
                    {claim.status === 'APPROVED' ? 'bg-green-100 text-green-800' :
                     claim.status === 'REJECTED' ? 'bg-red-100 text-red-800' :
                     'bg-yellow-100 text-yellow-800'}">
                    {claim.status}
                  </span>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>
  {/if}
</div>
