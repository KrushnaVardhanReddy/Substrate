<script lang="ts">
    import { fade } from 'svelte/transition';
    import { goto } from '$app/navigation';
    import { env } from '$env/dynamic/public';

    let step = $state(1);
    let progress = $state(0);
    
    function connectGithub() {
        step = 2;
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
                    step = 3;
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

    {#if step === 3}
        <div class="bg-glow cyan-glow"></div>
    {/if}
    
    <!-- Main Content Canvas -->
    <main class="onboarding-main">
        
        {#if step === 1}
            <div transition:fade={{ duration: 300 }} class="onboarding-card">
                <div class="icon-container">
                    <div class="icon-circle">
                        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="color: var(--accent);">
                            <path d="M18 6H5a2 2 0 0 0-2 2v3a2 2 0 0 0 2 2h13l4-3.5L18 6Z"></path>
                            <path d="M12 13v9"></path>
                            <path d="M12 2v4"></path>
                        </svg>
                    </div>
                </div>
                
                <h2 class="card-title">Welcome to Substrate</h2>
                <p class="card-subtitle">
                    Let's get your infrastructure connected. Start by linking your GitHub account.
                </p>
                
                <button 
                    class="action-button primary-action"
                    onclick={connectGithub}
                    aria-label="Connect GitHub"
                >
                    <svg class="github-icon" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"></path>
                    </svg>
                    Connect GitHub
                </button>
                
                <div class="secure-note">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                        <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                    </svg>
                    <p>Secure connection via OAuth. We only request read access.</p>
                </div>
            </div>
        {/if}

        {#if step === 2}
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

        {#if step === 3}
            <div transition:fade={{ duration: 300 }} class="onboarding-card success-card">
                <div class="icon-container">
                    <div class="icon-circle success-animation">
                        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--accent);">
                            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                            <polyline points="22 4 12 14.01 9 11.01"></polyline>
                        </svg>
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
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-left: 8px;">
                            <line x1="5" y1="12" x2="19" y2="12"></line>
                            <polyline points="12 5 19 12 12 19"></polyline>
                        </svg>
                    </button>
                    <button class="action-button secondary-action">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                            <polyline points="7 10 12 15 17 10"></polyline>
                            <line x1="12" y1="15" x2="12" y2="3"></line>
                        </svg>
                        Download Report
                    </button>
                </div>
                
                <div class="secure-note">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                        <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                    </svg>
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

    .github-icon {
        width: 24px;
        height: 24px;
        margin-right: 12px;
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
</style>
