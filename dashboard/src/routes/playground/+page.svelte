<script lang="ts">
    import { env } from '$env/dynamic/public';

    const templates = {
        graphql: {
            current: `type User {\n  id: ID!\n  user_id: String!\n  name: String\n  email: String\n}`,
            proposed: `type User {\n  id: ID!\n  name: String\n  email: String\n}`
        },
        openapi: {
            current: `openapi: 3.0.0\ninfo:\n  title: Sample API\n  version: 1.0.0\npaths:\n  /users/{id}:\n    get:\n      summary: Get a user by ID\n      responses:\n        '200':\n          description: OK`,
            proposed: `openapi: 3.0.0\ninfo:\n  title: Sample API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      summary: Get all users\n      responses:\n        '200':\n          description: OK`
        },
        sql: {
            current: `CREATE TABLE users (\n  id SERIAL PRIMARY KEY,\n  user_id VARCHAR(255) NOT NULL,\n  name VARCHAR(255),\n  email VARCHAR(255)\n);`,
            proposed: `CREATE TABLE users (\n  id SERIAL PRIMARY KEY,\n  name VARCHAR(255),\n  email VARCHAR(255)\n);`
        },
        protobuf: {
            current: `syntax = "proto3";\npackage users;\nmessage User {\n  string id = 1;\n  string email = 2;\n  string name = 3;\n}`,
            proposed: `syntax = "proto3";\npackage users;\nmessage User {\n  string id = 1;\n  string name = 3;\n}`
        },
        asyncapi: {
            current: `asyncapi: 2.6.0\ninfo:\n  title: User Events\n  version: 1.0.0\nchannels:\n  user.created:\n    publish:\n      message:\n        payload:\n          type: object\n          required: [id, email]\n          properties:\n            id: { type: string }\n            email: { type: string }`,
            proposed: `asyncapi: 2.6.0\ninfo:\n  title: User Events\n  version: 1.0.0\nchannels:\n  user.created:\n    publish:\n      message:\n        payload:\n          type: object\n          required: [id]\n          properties:\n            id: { type: string }`
        },
        terraform: {
            current: `resource "aws_s3_bucket" "data" {\n  bucket = "company-data"\n  force_destroy = false\n}`,
            proposed: `resource "aws_s3_bucket" "data" {\n  bucket = "company-data"\n  force_destroy = true\n}`
        }
    };

    let schemaType = $state<'graphql' | 'openapi' | 'sql' | 'protobuf' | 'asyncapi' | 'terraform'>('graphql');
    let currentSchema = $state(templates.graphql.current);
    let proposedSchema = $state(templates.graphql.proposed);

    function handleTypeChange() {
        currentSchema = templates[schemaType].current;
        proposedSchema = templates[schemaType].proposed;
    }
    let isAnalyzing = $state(false);
    let analysisResult = $state<null | 'success'>(null);

    let aiThinking = $state('');
    let aiFindings = $state<{severity: string, message: string}[]>([]);
    let aiFixCode = $state('');
    let aiFixLanguage = $state('');

    async function analyzeWithAI() {
        isAnalyzing = true;
        analysisResult = null;
        aiThinking = '';
        aiFindings = [];
        aiFixCode = '';
        aiFixLanguage = '';

        try {
            const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080';
            const res = await fetch(`${apiUrl}/api/v1/ai/analyze`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    org: 'playground',
                    current_schema: currentSchema,
                    proposed_schema: proposedSchema,
                    schema_type: schemaType
                })
            });

            if (!res.ok) {
                console.error("Failed to analyze", await res.text());
                isAnalyzing = false;
                return;
            }

            const reader = res.body?.getReader();
            const decoder = new TextDecoder();
            if (!reader) return;

            analysisResult = 'success';

            let buffer = '';
            while (true) {
                const { value, done } = await reader.read();
                if (done) break;
                buffer += decoder.decode(value, { stream: true });
                
                let newlineIndex;
                while ((newlineIndex = buffer.indexOf('\n')) >= 0) {
                    const line = buffer.slice(0, newlineIndex).trim();
                    buffer = buffer.slice(newlineIndex + 1);
                    
                    if (line.startsWith('data: ')) {
                        const dataStr = line.slice(6);
                        if (dataStr === '[DONE]') continue;
                        try {
                            const event = JSON.parse(dataStr);
                            if (event.type === 'thinking') {
                                aiThinking += event.content || '';
                            } else if (event.type === 'finding') {
                                aiFindings = [...aiFindings, { severity: event.severity, message: event.content }];
                            } else if (event.type === 'fix') {
                                aiFixCode += event.code || '';
                                if (event.language) aiFixLanguage = event.language;
                            } else if (event.type === 'done') {
                                isAnalyzing = false;
                            } else if (event.type === 'error') {
                                console.error('AI Error:', event.content);
                                isAnalyzing = false;
                            }
                        } catch (e) {
                            // partial JSON parse error
                        }
                    }
                }
            }
            isAnalyzing = false;
        } catch (e) {
            console.error("Stream error", e);
            isAnalyzing = false;
        }
    }

    function applyFix() {
        if (aiFixCode) {
            proposedSchema = aiFixCode;
        }
    }
</script>

<div class="content">
    <div class="page-header flex justify-between items-center">
        <div>
            <h1 class="page-title">AI Schema Validator Playground</h1>
            <p class="page-subtitle">Simulate cross-repo intelligence and auto-remediation with Substrate MCP tools.</p>
        </div>
        <div class="flex items-center gap-4">
            <select class="schema-select" bind:value={schemaType} onchange={handleTypeChange}>
                <option value="graphql">GraphQL</option>
                <option value="openapi">OpenAPI (REST)</option>
                <option value="sql">SQL (PostgreSQL)</option>
                <option value="protobuf">Protobuf (gRPC)</option>
                <option value="asyncapi">AsyncAPI (Kafka)</option>
                <option value="terraform">Terraform</option>
            </select>
            <button class="btn analyze-btn" onclick={analyzeWithAI} disabled={isAnalyzing}>
                {#if isAnalyzing}
                    <span class="sparkle-spin">✨</span> AI is analyzing cross-repo impact...
                {:else}
                    ✨ Analyze with Substrate AI
                {/if}
            </button>
        </div>
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
                <h3 class="card-title text-[var(--danger)]">🚨 Analysis Results</h3>
                
                {#if aiThinking}
                    <div class="mb-4 text-[var(--text-muted)] italic text-sm">
                        <span class="sparkle-spin mr-1">✨</span> {aiThinking}
                    </div>
                {/if}

                {#each aiFindings as finding}
                    <p class="mb-4 text-[var(--text-main)]">
                        <strong>{finding.severity}:</strong> {finding.message}
                    </p>
                {/each}

                {#if aiFixCode}
                <div class="auto-fix-section">
                    <div class="flex justify-between items-center mb-2">
                        <h4 class="font-medium text-[var(--safe)]">Suggested Auto-Fix</h4>
                        <button class="btn-sm btn-safe" onclick={applyFix}>Apply Fix</button>
                    </div>
                    <pre class="code-block"><code>{aiFixCode}</code></pre>
                </div>
                {/if}
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
        color: white;
        border: none;
        border-radius: 6px;
        cursor: pointer;
    }

    .analyze-btn:hover:not(:disabled) {
        background: linear-gradient(135deg, #4f46e5, #9333ea);
    }

    .analyze-btn:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }

    .schema-select {
        background-color: var(--bg-card);
        color: var(--text-main);
        border: 1px solid var(--border);
        padding: 8px 12px;
        border-radius: 6px;
        font-size: 0.95rem;
        outline: none;
        cursor: pointer;
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
        white-space: pre-wrap;
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
    .mr-1 { margin-right: 0.25rem; }
    .gap-4 { gap: 1rem; }
    .items-center { align-items: center; }
    .justify-between { justify-content: space-between; }
    .italic { font-style: italic; }
    .text-sm { font-size: 0.875rem; }
</style>
