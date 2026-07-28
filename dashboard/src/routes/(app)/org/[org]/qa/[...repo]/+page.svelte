<script lang="ts">
  import { page } from '$app/stores';

  let replaying = $state(false);
  let coverageResults: any = $state(null);
  let fuzzingGaps: any = $state([]);
  let selectedTimestamp = $state('2026-07-01'); // Default

  const org = $page.params.org;
  const repo = $page.params.repo;

  $effect(() => {
    fetch(`/api/v1/fuzzer/gaps`)
      .then(r => r.json())
      .then(data => fuzzingGaps = data)
      .catch(() => {});
  });

  async function exportPostman() {
    const res = await fetch(`/api/v1/qa/postman/${org}/${repo}`);
    if (res.ok) {
      const data = await res.json();
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `postman-${(repo || "").replace(/\//g, '-')}.json`;
      a.click();
    }
  }

  async function replayTraffic() {
    replaying = true;
    try {
      const res = await fetch(`/api/v1/qa/shadow/replay`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ org, repo, timestamp: selectedTimestamp })
      });

      if (res.ok) {
        // Simulate delay then fetch coverage
        setTimeout(async () => {
           const covRes = await fetch(`/api/v1/qa/coverage/${org}/${repo}`);
           if (covRes.ok) {
              coverageResults = await covRes.json();
           }
           replaying = false;
        }, 1000);
      } else {
        replaying = false;
      }
    } catch (e) {
      replaying = false;
    }
  }
</script>

<div class="p-6">
  <h1 class="text-2xl font-bold mb-4">QA Dashboard: {repo}</h1>

  <button
    class="bg-blue-600 text-white px-4 py-2 rounded mb-6"
    onclick={exportPostman}
  >
    Export Postman Collection
  </button>

  <div class="border rounded p-4 mb-6">
    <h2 class="text-xl font-bold mb-2">Time Machine (Shadow API Replay)</h2>
    <div class="mb-4">
      <label class="block mb-1">Select Timestamp:</label>
      <input type="text" bind:value={selectedTimestamp} class="border p-2 rounded" />
    </div>
    <button
      class="bg-green-600 text-white px-4 py-2 rounded"
      onclick={replayTraffic}
      disabled={replaying}
    >
      {replaying ? 'Replaying...' : 'Replay Traffic'}
    </button>
  </div>

  <div class="border rounded p-4 border-red-500">
    <h2 class="text-xl font-bold mb-2 text-red-600">Schema Validation Gaps</h2>
    {#if fuzzingGaps && fuzzingGaps.length > 0}
      <ul>
        {#each fuzzingGaps as gap}
          <li class="mb-2 p-2 bg-red-50 rounded">
            <strong>{gap.Severity}</strong> - {gap.Method} {gap.Path}<br />
            {gap.Issue}<br />
            <pre class="text-xs bg-gray-100 p-1 mt-1">{JSON.stringify(gap.Payload, null, 2)}</pre>
          </li>
        {/each}
      </ul>
    {:else}
      <p>No gaps found.</p>
    {/if}
  </div>

  {#if coverageResults}
    <div class="mt-6 p-4 bg-gray-100 rounded">
      <h3 class="font-bold">Coverage Results</h3>
      <pre>{JSON.stringify(coverageResults, null, 2)}</pre>
    </div>
  {/if}
</div>
