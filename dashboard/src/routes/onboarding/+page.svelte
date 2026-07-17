<script lang="ts">
    import { fade, slide } from 'svelte/transition';
    import { goto } from '$app/navigation';
    import { env } from '$env/dynamic/public';
    import { Server, CheckCircle, ChevronRight, Lock, Loader2, ArrowRight, Download } from 'lucide-svelte';
    import Github from '$lib/components/icons/Github.svelte';
    import Gitlab from '$lib/components/icons/Gitlab.svelte';

    let step = $state(1);
    let progress = $state(0);
    
    let selectedProvider = $state('');
    let token = $state('');
    let baseUrl = $state('');

    function selectProvider(provider: string) {
        selectedProvider = provider;
        step = 2;
    }

    function connectProvider() {
        if (!token) return;
        if (selectedProvider === 'custom' && !baseUrl) return;

        step = 3;
        simulateProgress();
    }

    function simulateProgress() {
        progress = 0;
        const interval = setInterval(() => {
            progress += 10;
            if (progress >= 100) {
                progress = 100;
                clearInterval(interval);
                setTimeout(() => {
                    step = 4;
                }, 500); 
            }
        }, 300); 
    }

    function enterDashboard() {
        const org = env.PUBLIC_ORG_NAME || 'demo';
        goto(`/org/${org}/graph`);
    }
</script>

<svelte:head>
    <title>Substrate - Onboarding</title>
</svelte:head>

<div class="onboarding-container">
    <!-- Animated background effects -->
    {#if step === 1}
        <div class="bg-glow blue-glow"></div>
        <div class="bg-glow purple-glow"></div>
    {/if}

    {#if step >= 3}
        <div class="bg-glow cyan-glow"></div>
    {/if}
    
    <!-- Main Content Canvas -->
    <main class="onboarding-main">
        
        {#if step === 1}
            <div transition:fade={{ duration: 300 }} class="onboarding-card">
                <div class="icon-container">
                    <div class="icon-circle">
                        <Server size={40} color="var(--accent)" strokeWidth={1.5} />
                    </div>
                </div>

                <h2 class="card-title">Select Git Provider</h2>
                <p class="card-subtitle">
                    Choose where your repositories are hosted to connect your infrastructure.
                </p>

                <div class="provider-grid">
                    <button
                        class="provider-card"
                        onclick={() => selectProvider('github')}
                        data-testid="provider-github-btn"
                    >
                        <div class="provider-icon github">
                            <Github size={32} />
                        </div>
                        <div class="provider-info">
                            <span class="provider-name">GitHub</span>
                            <span class="provider-desc">Cloud or Enterprise</span>
                        </div>
                        <div class="provider-arrow">
                            <ChevronRight size={20} />
                        </div>
                    </button>

                    <button
                        class="provider-card"
                        onclick={() => selectProvider('gitlab')}
                        data-testid="provider-gitlab-btn"
                    >
                        <div class="provider-icon gitlab">
                            <Gitlab size={32} />
                        </div>
                        <div class="provider-info">
                            <span class="provider-name">GitLab</span>
                            <span class="provider-desc">SaaS or Self-Managed</span>
                        </div>
                        <div class="provider-arrow">
                            <ChevronRight size={20} />
                        </div>
                    </button>

                    <button
                        class="provider-card"
                        onclick={() => selectProvider('custom')}
                        data-testid="provider-custom-btn"
                    >
                        <div class="provider-icon custom">
                            <Server size={32} />
                        </div>
                        <div class="provider-info">
                            <span class="provider-name">Self-Hosted</span>
                            <span class="provider-desc">Gitea, Bitbucket, etc.</span>
                        </div>
                        <div class="provider-arrow">
                            <ChevronRight size={20} />
                        </div>
                    </button>
                </div>
            </div>
        {/if}

        {#if step === 2}
            <div transition:slide={{ duration: 400 }} class="onboarding-card">
                <div class="icon-container">
                    <div class="icon-circle">
                        {#if selectedProvider === 'github'}
                            <Github size={40} color="var(--accent)" strokeWidth={1.5} />
                        {:else if selectedProvider === 'gitlab'}
                            <Gitlab size={40} color="var(--accent)" strokeWidth={1.5} />
                        {:else}
                            <Server size={40} color="var(--accent)" strokeWidth={1.5} />
                        {/if}
                    </div>
                </div>
                
                <h2 class="card-title">
                    Connect
                    {#if selectedProvider === 'github'}GitHub
                    {:else if selectedProvider === 'gitlab'}GitLab
                    {:else}Self-Hosted Server{/if}
                </h2>
                <p class="card-subtitle">
                    Enter your credentials to securely connect.
                </p>

                <div class="form-container">
                    {#if selectedProvider === 'custom'}
                        <div class="input-group" transition:slide>
                            <label for="base-url">Server URL</label>
                            <input
                                id="base-url"
                                type="url"
                                bind:value={baseUrl}
                                placeholder="https://git.yourcompany.com"
                                data-testid="base-url-input"
                                class="auth-input"
                            />
                        </div>
                    {/if}

                    <div class="input-group">
                        <label for="token">Personal Access Token</label>
                        <input
                            id="token"
                            type="password"
                            bind:value={token}
                            placeholder="Enter Token"
                            data-testid="token-input"
                            class="auth-input"
                        />
                    </div>
                </div>
                
                <button 
                    class="action-button primary-action mt-4"
                    onclick={connectProvider}
                    disabled={!token || (selectedProvider === 'custom' && !baseUrl)}
                    aria-label="Connect Provider"
                    data-testid="connect-provider-btn"
                >
                    Connect
                    <ArrowRight size={18} style="margin-left: 8px;" />
                </button>

                <button
                    class="back-button mt-4"
                    onclick={() => step = 1}
                >
                    Back to Providers
                </button>
                
                <div class="secure-note">
                    <Lock size={16} />
                    <p>Secure connection via OAuth. We only request read access.</p>
                </div>
            </div>
        {/if}

        {#if step === 3}
            <div transition:fade={{ duration: 300 }} class="onboarding-card">
                <div class="icon-container">
                    <div class="icon-circle scanning-animation">
                        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--accent);">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="12" y1="12" x2="12" y2="2"></line>
                        </svg>
                    </div>
                </div>
                
                <h2 class="card-title">Scanning your Repositories</h2>
                <p class="card-subtitle">
                    We're analyzing your codebases to prepare your enterprise dashboard. This will only take a moment.
                </p>
                
                <div class="progress-container">
                    <div class="progress-header">
                        <span class="progress-label">Progress</span>
                        <span class="progress-value" data-testid="progress-text">{progress}%</span>
                    </div>
                    <div class="progress-bar-track">
                        <div class="progress-bar-fill" style="width: {progress}%"></div>
                    </div>
                </div>
                
                <div class="scan-status">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                    </svg>
                    <span>Found 12 repositories...</span>
                </div>
            </div>
        {/if}

        {#if step === 4}
            <div transition:fade={{ duration: 300 }} class="onboarding-card success-card">
                <div class="icon-container">
                    <div class="icon-circle success-animation">
                        <CheckCircle size={48} color="var(--accent)" strokeWidth={1.5} />
                    </div>
                </div>
                
                <h2 class="card-title">Ready to Launch</h2>
                <p class="card-subtitle">
                    Scanning complete. Your infrastructure is mapped, secured, and ready for centralized management.
                </p>
                
                <div class="action-group">
                    <button 
                        class="action-button primary-action"
                        onclick={enterDashboard}
                        data-testid="enter-dashboard-btn"
                        aria-label="Enter Dashboard"
                    >
                        Enter Dashboard
                        <ArrowRight size={18} style="margin-left: 8px;" />
                    </button>
                    <button class="action-button secondary-action">
                        <Download size={18} style="margin-right: 8px;" />
                        Download Report
                    </button>
                </div>
                
                <div class="secure-note">
                    <Lock size={16} />
                    <span>Session encrypted end-to-end. Connection secured.</span>
                </div>
            </div>
        {/if}
    </main>
</div>

<style>
    .onboarding-container {
        min-height: 100vh;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        position: relative;
        overflow: hidden;
        background-color: var(--bg-dark);
    }

    .bg-glow {
        position: absolute;
        width: 50vw;
        height: 50vh;
        border-radius: 50%;
        filter: blur(120px);
        opacity: 0.15;
        z-index: 0;
        pointer-events: none;
    }

    .blue-glow {
        top: -10%;
        left: -10%;
        background-color: #00BFA5;
    }

    .purple-glow {
        bottom: -10%;
        right: -10%;
        background-color: #7C4DFF;
    }

    .cyan-glow {
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        background-color: var(--accent);
        width: 60vw;
        height: 60vh;
        opacity: 0.1;
    }

    .onboarding-main {
        position: relative;
        z-index: 10;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 2rem;
    }

    .onboarding-card {
        width: 100%;
        max-width: 600px;
        background: var(--bg-card);
        border: 1px solid var(--border);
        border-radius: 16px;
        padding: 48px 40px;
        display: flex;
        flex-direction: column;
        align-items: center;
        text-align: center;
        box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
        backdrop-filter: blur(12px);
    }

    .icon-container {
        margin-bottom: 32px;
        position: relative;
    }

    .icon-circle {
        width: 96px;
        height: 96px;
        border-radius: 50%;
        background: rgba(0, 191, 165, 0.1);
        border: 1px solid rgba(0, 191, 165, 0.2);
        display: flex;
        align-items: center;
        justify-content: center;
        position: relative;
        z-index: 2;
    }

    .scanning-animation svg {
        animation: spin 3s linear infinite;
    }

    .success-animation {
        background: rgba(0, 191, 165, 0.15);
        box-shadow: 0 0 30px rgba(0, 191, 165, 0.2);
    }

    @keyframes spin {
        from { transform: rotate(0deg); }
        to { transform: rotate(360deg); }
    }

    .card-title {
        font-size: 2.25rem;
        font-weight: 700;
        color: var(--text-main);
        margin-bottom: 16px;
        letter-spacing: -0.02em;
    }

    .card-subtitle {
        font-size: 1.1rem;
        color: var(--text-muted);
        line-height: 1.6;
        margin-bottom: 40px;
        max-width: 480px;
    }

    .action-button {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        font-size: 1.1rem;
        font-weight: 600;
        padding: 16px 32px;
        border-radius: 12px;
        cursor: pointer;
        transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
        border: none;
    }

    .primary-action {
        background: var(--text-main);
        color: var(--bg-dark);
        box-shadow: 0 0 20px rgba(255, 255, 255, 0.1);
    }

    .primary-action:hover {
        transform: scale(1.02);
        box-shadow: 0 0 30px rgba(255, 255, 255, 0.2);
    }

    .primary-action:active {
        transform: scale(0.98);
    }

    .secondary-action {
        background: transparent;
        color: var(--text-main);
        border: 1px solid var(--border);
    }

    .secondary-action:hover {
        background: var(--bg-hover);
    }

    .action-group {
        display: flex;
        gap: 16px;
        margin-bottom: 32px;
    }

    .secure-note {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 0.9rem;
        color: var(--text-muted);
        margin-top: 32px;
        background: rgba(255, 255, 255, 0.03);
        padding: 10px 20px;
        border-radius: 20px;
        border: 1px solid var(--border);
    }

    /* Progress Bar */
    .progress-container {
        width: 100%;
        max-width: 400px;
        margin-bottom: 24px;
    }

    .progress-header {
        display: flex;
        justify-content: space-between;
        margin-bottom: 12px;
        font-size: 0.95rem;
    }

    .progress-label {
        color: var(--text-muted);
    }

    .progress-value {
        color: var(--accent);
        font-weight: 600;
    }

    .progress-bar-track {
        height: 12px;
        background: rgba(255, 255, 255, 0.05);
        border-radius: 6px;
        overflow: hidden;
        border: 1px solid var(--border);
    }

    .progress-bar-fill {
        height: 100%;
        background: var(--accent);
        border-radius: 6px;
        transition: width 0.3s ease-out;
        box-shadow: 0 0 10px rgba(0, 191, 165, 0.5);
    }

    .scan-status {
        display: flex;
        align-items: center;
        gap: 8px;
        color: var(--text-muted);
        font-size: 0.95rem;
    }

    .provider-grid {
        display: flex;
        flex-direction: column;
        gap: 16px;
        width: 100%;
        margin-top: 10px;
        margin-bottom: 24px;
    }

    .provider-card {
        display: flex;
        align-items: center;
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid var(--border);
        border-radius: 12px;
        padding: 16px 20px;
        cursor: pointer;
        transition: all 0.2s ease-out;
        width: 100%;
        text-align: left;
    }

    .provider-card:hover {
        background: rgba(255, 255, 255, 0.06);
        border-color: rgba(255, 255, 255, 0.2);
        transform: translateY(-2px);
        box-shadow: 0 10px 20px -10px rgba(0, 0, 0, 0.5);
    }

    .provider-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 48px;
        height: 48px;
        border-radius: 10px;
        margin-right: 16px;
    }

    .provider-icon.github {
        background: rgba(255, 255, 255, 0.1);
        color: #ffffff;
    }

    .provider-icon.gitlab {
        background: rgba(252, 109, 38, 0.1);
        color: #FC6D26;
    }

    .provider-icon.custom {
        background: rgba(0, 191, 165, 0.1);
        color: var(--accent);
    }

    .provider-info {
        display: flex;
        flex-direction: column;
        flex: 1;
    }

    .provider-name {
        font-weight: 600;
        font-size: 1.1rem;
        color: var(--text-main);
        margin-bottom: 4px;
    }

    .provider-desc {
        font-size: 0.9rem;
        color: var(--text-muted);
    }

    .provider-arrow {
        color: var(--text-muted);
        transition: transform 0.2s ease;
    }

    .provider-card:hover .provider-arrow {
        transform: translateX(4px);
        color: var(--text-main);
    }

    .form-container {
        width: 100%;
        display: flex;
        flex-direction: column;
        gap: 20px;
        margin-bottom: 8px;
    }

    .input-group {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        width: 100%;
    }

    .input-group label {
        font-size: 0.9rem;
        color: var(--text-muted);
        margin-bottom: 8px;
        margin-left: 4px;
    }

    .auth-input {
        width: 100%;
        padding: 14px 16px;
        border-radius: 10px;
        border: 1px solid var(--border);
        background: rgba(255, 255, 255, 0.04);
        color: var(--text-main);
        font-size: 1rem;
        transition: all 0.2s ease;
    }

    .auth-input:focus {
        outline: none;
        border-color: var(--accent);
        background: rgba(255, 255, 255, 0.08);
        box-shadow: 0 0 0 2px rgba(0, 191, 165, 0.2);
    }

    .back-button {
        background: transparent;
        border: none;
        color: var(--text-muted);
        font-size: 0.95rem;
        cursor: pointer;
        padding: 8px 16px;
        border-radius: 6px;
        transition: all 0.2s ease;
    }

    .back-button:hover {
        color: var(--text-main);
        background: rgba(255, 255, 255, 0.05);
    }

    .mt-4 {
        margin-top: 16px;
    }

    button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>
