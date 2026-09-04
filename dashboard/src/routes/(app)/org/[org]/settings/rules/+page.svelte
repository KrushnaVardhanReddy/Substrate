<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { invalidateAll } from '$app/navigation';

  let rules: { id: string, rule_text: string }[] = $state([]);
  let newRuleText = $state('');
  let loading = $state(true);
  let error = $state('');

  const org = $page.params.org;

  async function fetchRules() {
    loading = true;
    error = '';
    try {
      const res = await fetch(`/api/v1/org/${org}/rules`);
      if (!res.ok) throw new Error('Failed to fetch rules');
      rules = await res.json();
    } catch (e: any) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function addRule() {
    if (!newRuleText.trim()) return;
    try {
      const res = await fetch(`/api/v1/org/${org}/rules`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ rule_text: newRuleText })
      });
      if (!res.ok) throw new Error('Failed to add rule');
      newRuleText = '';
      await fetchRules();
    } catch (e: any) {
      error = e.message;
    }
  }

  async function deleteRule(id: string) {
    try {
      const res = await fetch(`/api/v1/org/${org}/rules/${id}`, {
        method: 'DELETE'
      });
      if (!res.ok) throw new Error('Failed to delete rule');
      await fetchRules();
    } catch (e: any) {
      error = e.message;
    }
  }

  onMount(() => {
    fetchRules();
  });
</script>

<div class="p-6 max-w-4xl mx-auto space-y-8">
  <div>
    <h1 class="text-3xl font-bold tracking-tight text-[var(--text-primary)]">Governance Rules</h1>
    <p class="text-[var(--text-muted)]">Manage API governance rules for your organization.</p>
  </div>

  {#if error}
    <div class="p-4 bg-red-50 text-red-600 rounded-md">
      {error}
    </div>
  {/if}

  <div class="bg-[var(--bg-card)] rounded-xl border border-[var(--border-subtle)] overflow-hidden shadow-sm">
    <div class="p-6">
      <form onsubmit={(e) => { e.preventDefault(); addRule(); }} class="flex gap-4 mb-6">
        <input
          type="text"
          bind:value={newRuleText}
          placeholder="e.g., All endpoints must use camelCase"
          class="flex-1 bg-[var(--bg-primary)] border border-[var(--border-subtle)] text-[var(--text-primary)] rounded-md px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          type="submit"
          disabled={!newRuleText.trim()}
          class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md font-medium disabled:opacity-50 transition-colors"
        >
          Add Rule
        </button>
      </form>

      {#if loading}
        <div class="flex justify-center p-8">
          <div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
        </div>
      {:else if rules && rules.length === 0}
        <div class="text-center p-8 text-[var(--text-muted)] border-t border-[var(--border-subtle)] border-dashed">
          No governance rules found. Add one above.
        </div>
      {:else if rules && rules.length > 0}
        <ul class="divide-y divide-[var(--border-subtle)] border-t border-[var(--border-subtle)]">
          {#each rules as rule (rule.id)}
            <li class="py-4 flex justify-between items-center group">
              <span class="text-[var(--text-primary)]">{rule.rule_text}</span>
              <button
                onclick={() => deleteRule(rule.id)}
                class="text-red-500 hover:text-red-700 opacity-0 group-hover:opacity-100 transition-opacity p-2 rounded-md hover:bg-red-50"
                aria-label="Delete rule"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 6h18"></path>
                  <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
                  <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
                </svg>
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>
</div>
