<script lang="ts">
    import { page } from '$app/stores';
    import { env } from '$env/dynamic/public';

    let prompt = $state("");
    let cel = $state("");
    let error = $state("");
    let loading = $state(false);

    async function generateRule() {
        if (!prompt.trim()) {
            error = "Please enter a rule description.";
            return;
        }

        error = "";
        cel = "";
        loading = true;

        try {
            const apiUrl = env.PUBLIC_API_URL || '';
            const res = await fetch(`${apiUrl}/api/governance/generate-cel`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ prompt })
            });

            const data = await res.json();

            if (!res.ok) {
                error = data.error || "Failed to generate rule";
            } else {
                cel = data.cel;
            }
        } catch (err) {
            error = "Network error while generating rule.";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>Governance Rules - Substrate</title>
</svelte:head>

<div class="content" style="max-width: 800px; margin: 0 auto;">
    <div class="page-header">
        <h1 class="page-title">Governance Rules</h1>
        <p class="page-subtitle">Generate and manage custom CEL rules using Natural Language.</p>
    </div>

    <div class="card" style="display: flex; flex-direction: column; gap: 16px;">
        <div style="display: flex; flex-direction: column; gap: 8px;">
            <label for="prompt" style="font-size: 0.9rem; font-weight: 500; color: var(--text-muted);">Rule Description</label>
            <textarea
                id="prompt"
                bind:value={prompt}
                style="width: 100%; border-radius: 6px; background-color: var(--bg-dark); border: 1px solid var(--border); color: var(--text-main); padding: 12px; font-family: inherit; font-size: 0.9rem; resize: vertical;"
                rows="4"
                placeholder="e.g. All payment APIs must require authentication"
            ></textarea>
        </div>

        <div style="align-self: flex-start;">
            <button
                onclick={generateRule}
                disabled={loading}
                class="btn"
                style={loading ? "opacity: 0.5; cursor: not-allowed;" : ""}
            >
                {loading ? 'Generating...' : 'Generate Rule'}
            </button>
        </div>

        {#if error}
            <div style="padding: 12px; background-color: rgba(255, 82, 82, 0.1); color: var(--danger); border: 1px solid rgba(255, 82, 82, 0.3); border-radius: 6px; font-size: 0.9rem;">
                {error}
            </div>
        {/if}

        {#if cel}
            <div style="margin-top: 16px;">
                <h3 style="font-size: 0.9rem; font-weight: 500; color: var(--text-muted); margin-bottom: 8px;">Generated CEL Expression:</h3>
                <pre style="background-color: var(--bg-dark); padding: 16px; border-radius: 6px; overflow-x: auto; border: 1px solid var(--border); font-size: 0.9rem; font-family: monospace;"><code>{cel}</code></pre>
            </div>
        {/if}
    </div>
</div>
