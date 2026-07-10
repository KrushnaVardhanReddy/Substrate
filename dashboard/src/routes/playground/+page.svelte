<script lang="ts">
    let currentSchema = $state(`type User {
  id: ID!
  user_id: String!
  name: String
  email: String
}`);
    let proposedSchema = $state(`type User {
  id: ID!
  name: String
  email: String
}`);
    let isAnalyzing = $state(false);
    let analysisResult = $state<null | 'success'>(null);

    const autoFixSchema = `type User {
  id: ID!
  user_id: String! @deprecated(reason: "Use id instead")
  name: String
  email: String
}`;

    function analyzeWithAI() {
        isAnalyzing = true;
        analysisResult = null;

        // Simulate streaming / analyzing delay
        setTimeout(() => {
            isAnalyzing = false;
            analysisResult = 'success';
        }, 1500);
    }

    function applyFix() {
        proposedSchema = autoFixSchema;
    }
</script>

<div class="content">
    <div class="page-header flex justify-between items-center">
        <div>
            <h1 class="page-title">AI Schema Validator Playground</h1>
            <p class="page-subtitle">Simulate cross-repo intelligence and auto-remediation with Substrate MCP tools.</p>
        </div>
        <button class="btn analyze-btn" onclick={analyzeWithAI} disabled={isAnalyzing}>
            {#if isAnalyzing}
                <span class="sparkle-spin">✨</span> AI is analyzing cross-repo impact...
            {:else}
                ✨ Analyze with Substrate AI
            {/if}
        </button>
    </div>

    <div class="playground-container">
        <div class="split-pane">
            <div class="editor-pane">
                <h3 class="pane-title">Current Schema</h3>
                <textarea class="code-editor" bind:value={currentSchema}></textarea>
            </div>
            <div class="editor-pane">
                <h3 class="pane-title">Proposed Schema</h3>
                <textarea class="code-editor" bind:value={proposedSchema}></textarea>
            </div>
        </div>

        {#if analysisResult === 'success'}
            <div class="analysis-panel card">
                <h3 class="card-title text-[var(--danger)]">🚨 Breaking Change Detected</h3>
                <p class="mb-4 text-[var(--text-main)]">
                    Contextual Impact: Removing <code>user_id</code> breaks downstream consumer <strong>Billing API v2</strong>.
                </p>

                <div class="auto-fix-section">
                    <div class="flex justify-between items-center mb-2">
                        <h4 class="font-medium text-[var(--safe)]">Suggested Auto-Fix</h4>
                        <button class="btn-sm btn-safe" onclick={applyFix}>Apply Fix</button>
                    </div>
                    <pre class="code-block"><code>{autoFixSchema}</code></pre>
                    <p class="text-sm text-[var(--text-muted)] mt-2">
                        Safely remediated by adding <code>@deprecated</code> instead of deleting the field.
                    </p>
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    .playground-container {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .split-pane {
        display: flex;
        gap: 20px;
        height: 400px;
    }

    .editor-pane {
        flex: 1;
        display: flex;
        flex-direction: column;
        background-color: var(--bg-card);
        border: 1px solid var(--border);
        border-radius: 8px;
        overflow: hidden;
    }

    .pane-title {
        padding: 10px 15px;
        background-color: rgba(0, 0, 0, 0.2);
        border-bottom: 1px solid var(--border);
        font-size: 0.9rem;
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .code-editor {
        flex: 1;
        width: 100%;
        background-color: transparent;
        color: var(--text-main);
        border: none;
        padding: 15px;
        font-family: monospace;
        font-size: 14px;
        resize: none;
        outline: none;
        line-height: 1.5;
    }

    .code-editor:focus {
        box-shadow: inset 0 0 0 1px var(--accent);
    }

    .analyze-btn {
        background: linear-gradient(135deg, #6366f1, #a855f7);
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 1rem;
        padding: 10px 20px;
    }

    .analyze-btn:hover:not(:disabled) {
        background: linear-gradient(135deg, #4f46e5, #9333ea);
    }

    .analyze-btn:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }

    @keyframes spin {
        100% { transform: rotate(360deg); }
    }

    .sparkle-spin {
        display: inline-block;
        animation: spin 2s linear infinite;
    }

    .analysis-panel {
        border-left: 4px solid var(--danger);
    }

    .auto-fix-section {
        background-color: rgba(0, 0, 0, 0.2);
        padding: 15px;
        border-radius: 6px;
        border: 1px solid var(--border);
    }

    .code-block {
        background-color: #0d1117;
        padding: 12px;
        border-radius: 4px;
        overflow-x: auto;
        font-family: monospace;
        color: #c9d1d9;
        border: 1px solid var(--border);
    }

    .btn-sm {
        padding: 4px 10px;
        font-size: 0.85rem;
        border-radius: 4px;
        cursor: pointer;
        font-weight: 500;
        border: none;
        transition: opacity 0.2s;
    }

    .btn-safe {
        background-color: var(--safe);
        color: white;
    }

    .btn-safe:hover {
        opacity: 0.9;
    }

    .mb-4 { margin-bottom: 1rem; }
    .mb-2 { margin-bottom: 0.5rem; }
    .mt-2 { margin-top: 0.5rem; }
    .justify-between { justify-content: space-between; }
</style>
