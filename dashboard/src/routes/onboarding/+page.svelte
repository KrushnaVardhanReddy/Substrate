<script lang="ts">
    import { fade } from 'svelte/transition';
    import { goto } from '$app/navigation';
    import { tick } from 'svelte';

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
        goto('/');
    }
</script>

<svelte:head>
    <title>Nexus Enterprise - Onboarding</title>
    <script src="https://cdn.tailwindcss.com?plugins=forms,container-queries"></script>
    <link href="https://fonts.googleapis.com/css2?family=Public+Sans:wght@300;400;500;600;700&display=swap" rel="stylesheet"/>
    <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap" rel="stylesheet"/>
    
    <!-- Using raw HTML for the script tag to ensure it works correctly inside svelte:head without Vite interfering -->
    {@html `
    <script id="tailwind-config">
        tailwind.config = {
            darkMode: "class",
            theme: {
                extend: {
                    colors: {
                        primary: "#0f172a",
                        "primary-fixed": "#334155",
                        "primary-fixed-dim": "#475569",
                        "on-primary-container": "#f8fafc",
                        "primary-container": "#e2e8f0",
                        "surface": "#ffffff",
                        "surface-container-low": "#f8fafc",
                        "surface-container-highest": "#f1f5f9",
                        "on-surface-variant": "#64748b",
                        "outline-variant": "#e2e8f0"
                    },
                    borderRadius: {
                        "DEFAULT": "0.25rem",
                        "lg": "0.5rem",
                        "xl": "0.75rem",
                        "full": "9999px"
                    },
                    fontFamily: {
                        "headline": ["Public Sans", "sans-serif"],
                        "display": ["Public Sans", "sans-serif"],
                        "body": ["Public Sans", "sans-serif"],
                        "label": ["Public Sans", "sans-serif"]
                    },
                    spacing: {
                        "sm": "0.5rem",
                        "md": "1rem",
                        "lg": "1.5rem",
                        "xl": "2rem"
                    }
                },
            },
        }
    </script>`}
</svelte:head>

<div class="bg-primary text-slate-100 min-h-screen font-body flex overflow-hidden relative w-full h-screen" style="font-family: 'Public Sans', sans-serif;">
    <!-- Background Effect -->
    {#if step === 1}
        <div class="absolute inset-0 overflow-hidden pointer-events-none z-0">
            <div class="absolute top-[-20%] left-[-10%] w-[50%] h-[50%] rounded-full bg-blue-500/10 blur-[120px]"></div>
            <div class="absolute bottom-[-20%] right-[-10%] w-[50%] h-[50%] rounded-full bg-purple-500/10 blur-[120px]"></div>
        </div>
    {/if}

    {#if step === 3}
        <div class="absolute inset-0 pointer-events-none bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-cyan-50/50 via-transparent to-transparent"></div>
    {/if}
    
    <!-- Shared SideNavBar -->
    <nav class="w-[240px] h-screen fixed left-0 top-0 border-r border-outline-variant flex flex-col py-lg px-md gap-xl z-20 shadow-none {step === 3 ? 'bg-surface-container-low' : 'bg-surface-container-low'}">
        <div class="flex items-center gap-3 px-2 mt-4 mb-4">
            <div class="w-10 h-10 rounded {step === 1 ? 'bg-primary-fixed' : step === 2 ? 'bg-blue-600/10' : 'bg-slate-900'} flex items-center justify-center {step === 1 ? 'text-white' : step === 2 ? 'text-blue-600' : 'text-white'} font-bold text-xl overflow-hidden">
                {#if step === 2}
                    <img alt="Nexus Enterprise Logo" class="object-cover w-full h-full" src="https://lh3.googleusercontent.com/aida-public/AB6AXuCi5IUbGT6J0g7tevpF1wxabuywlh9-CLn17Nh-6yysABncd4jqGESny_1YuyDvG00SeaAx-NIhnKPkqjPty-RF2IU6jyJY1HW5alfluGNc8_HO8bwnKUNKISCFOgG0iQ9bssniUl63eU8QZ2v9L-xoYhYRfemO5f4Bf98an27BF242sJKg7aog4J3MaE7-Gakl9UABHnbS8-SqALihTyoZqbQOx4xC3wdrgH6lxZPn0HSuF4IWL9Y2h30ZbilK1mf3yWmXp6nsb7M"/>
                {:else if step === 3}
                    <span class="material-symbols-outlined text-white text-sm" style="font-variation-settings: 'FILL' 1;">dataset</span>
                {:else}
                    N
                {/if}
            </div>
            <div>
                <h1 class="font-headline font-bold tracking-tighter text-primary-fixed text-lg leading-tight">NEXUS</h1>
                <p class="text-xs text-on-surface-variant font-medium">Enterprise Setup</p>
            </div>
        </div>
        
        <div class="flex flex-col gap-2 flex-grow mt-8">
            <!-- Connection Tab -->
            <div class="flex items-center gap-3 px-4 py-3 rounded transition-all {step === 1 ? 'text-primary font-bold border-r-2 border-primary bg-primary/5 opacity-80 scale-[0.99]' : 'text-green-500 font-medium hover:bg-slate-100'}">
                <span class="material-symbols-outlined" style="font-variation-settings: 'FILL' 1;">{step > 1 ? 'check' : 'hub'}</span>
                <span class="font-body text-sm">Connection</span>
            </div>
            
            <!-- Scanning Tab -->
            <div class="flex items-center gap-3 px-4 py-3 rounded transition-all {step === 2 ? 'text-blue-600 font-bold border-r-2 border-blue-600 bg-blue-600/5 opacity-80 scale-[0.99]' : step > 2 ? 'text-green-500 font-medium hover:bg-slate-100' : 'text-on-surface-variant font-medium hover:bg-surface-container-highest'}">
                <span class="material-symbols-outlined" style="font-variation-settings: 'FILL' 1;">{step > 2 ? 'check' : 'radar'}</span>
                <span class="font-body text-sm">Scanning</span>
            </div>
            
            <!-- Completion Tab -->
            <div class="flex items-center gap-3 px-4 py-3 rounded transition-all {step === 3 ? 'text-primary font-bold border-r-2 border-primary bg-slate-200/50 opacity-80 scale-[0.99]' : 'text-on-surface-variant font-medium hover:bg-surface-container-highest'}">
                <span class="material-symbols-outlined" style="font-variation-settings: 'FILL' 1;">check_circle</span>
                <span class="font-body text-sm">Completion</span>
            </div>
        </div>
        
        {#if step === 1}
        <div class="mt-auto px-2 mb-4">
            <div class="flex items-center gap-3 px-2 py-2">
                <img alt="User Avatar" class="w-8 h-8 rounded-full bg-slate-200 object-cover" src="https://lh3.googleusercontent.com/aida-public/AB6AXuBefaEon10QBgaZCaeAxgftZd1RpJ9bbsRgWai-pcVRxeGkPI_6tBYEB1FKf5FwTFFnLAmyMYvSqVKR1gxp3UIHOJJUWPyB3S_oSCQk9oLPLfvZpr5zNIydmKZ3eiNQJFnS3kd3Z2RRujityfbGwAIQK6vF9Uf4j4BZtDzLDkOflS7taj6iHaHC-T75q2mmGsC6eojArlivgF3ZVhRJFkpFuKF_rlhfdwY8GrjgRxpHzvW-pUyWL1EV77W73ieaGirHqC1vthI1-OE"/>
                <div class="flex flex-col">
                    <span class="text-xs font-bold text-primary-fixed">Setup Admin</span>
                    <span class="text-[10px] text-on-surface-variant">Step 1 of 3</span>
                </div>
            </div>
        </div>
        {/if}
    </nav>
    
    <!-- Shared TopAppBar -->
    <header class="fixed top-0 flex justify-between items-center h-16 ml-[240px] px-lg w-[calc(100%-240px)] border-b border-outline-variant z-10 shadow-none {step === 1 ? 'bg-surface' : step === 2 ? 'bg-surface' : 'bg-surface'}">
        <div class="flex items-center gap-2">
            <span class="font-headline font-bold text-primary-fixed text-sm">Nexus Enterprise</span>
            {#if step === 1}
                <span class="text-on-surface-variant material-symbols-outlined text-sm">chevron_right</span>
                <span class="text-primary font-medium text-sm">Onboarding</span>
            {/if}
        </div>
        <div class="flex items-center gap-4 text-on-surface-variant">
            <button class="hover:text-primary-fixed-dim transition-colors p-2 rounded-full hover:bg-slate-100 flex items-center justify-center">
                <span class="material-symbols-outlined">settings</span>
            </button>
            <button class="hover:text-primary-fixed-dim transition-colors p-2 rounded-full hover:bg-slate-100 flex items-center justify-center">
                <span class="material-symbols-outlined">help_outline</span>
            </button>
            {#if step > 1}
            <div class="w-8 h-8 rounded-full bg-slate-200 overflow-hidden border border-slate-300 ml-2">
                <img alt="User Avatar" class="w-full h-full object-cover" src="https://lh3.googleusercontent.com/aida-public/AB6AXuDdZZPyE0niQwYsNuNB_LCPUVgkgpAebzXPeFVzvF93DDs8rqqxHwOPD8A2HAXqwP3Qeb9WwYfwNYj5DwEaZLwWsLJLtmj6q-svm-5jZsTxIHWdfiUVqfKQxZPCcqfo5AzDrf59SuIlTy83T8z5gCcYCmxJeqsbJu5zocJVE0aEUIbuBfw7v_XtFMvYCKhu15aV0m2f2J9AQOIBuT49exx-0dNEFYYQVITlV5x5U7bkdM7zFholDQd0yGbB-Qg5eNTpc8QHeq35pO0"/>
            </div>
            {/if}
        </div>
    </header>

    <!-- Main Content Canvas -->
    <main class="ml-[240px] mt-16 p-8 flex-grow flex items-center justify-center relative z-10 w-[calc(100%-240px)] min-h-[calc(100vh-64px)] {step === 2 || step === 3 ? 'bg-surface' : ''}">
        
        {#if step === 1}
            <div transition:fade={{ duration: 300 }} class="absolute w-full max-w-2xl bg-slate-900/40 backdrop-blur-md border border-slate-700/50 rounded-2xl p-10 shadow-2xl flex flex-col items-center text-center transform transition-all duration-500 hover:shadow-blue-900/20 hover:border-slate-600/50">
                <div class="mb-8 relative">
                    <div class="w-24 h-24 rounded-full bg-blue-500/10 flex items-center justify-center border border-blue-500/20 relative z-10">
                        <span class="material-symbols-outlined text-blue-400 text-5xl" style="font-variation-settings: 'FILL' 1;">hub</span>
                    </div>
                    <div class="absolute inset-0 bg-blue-400/20 blur-xl rounded-full z-0 animate-pulse"></div>
                </div>
                
                <h2 class="text-3xl md:text-4xl font-headline font-bold text-white mb-4 tracking-tight">Welcome to Nexus Enterprise</h2>
                <p class="text-slate-400 text-base md:text-lg max-w-md mx-auto mb-10 leading-relaxed">
                    Let's get your infrastructure connected. Start by linking your GitHub account.
                </p>
                
                <button 
                    class="group flex items-center gap-3 bg-white text-slate-900 px-8 py-4 rounded-xl font-bold text-lg hover:bg-slate-100 transition-all duration-200 hover:scale-105 active:scale-95 shadow-[0_0_20px_rgba(255,255,255,0.1)] hover:shadow-[0_0_30px_rgba(255,255,255,0.2)]"
                    onclick={connectGithub}
                    aria-label="Connect GitHub"
                >
                    <svg aria-hidden="true" class="w-6 h-6 fill-current" viewBox="0 0 24 24">
                        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"></path>
                    </svg>
                    Connect GitHub
                    <span class="material-symbols-outlined opacity-0 group-hover:opacity-100 transition-opacity translate-x-[-10px] group-hover:translate-x-0 duration-200">arrow_forward</span>
                </button>
                
                <div class="mt-6 flex items-center justify-center gap-2 text-sm text-slate-500 bg-slate-800/50 px-4 py-2 rounded-full border border-slate-700/50">
                    <span class="material-symbols-outlined text-[16px]">lock</span>
                    <p>Secure connection via OAuth. We only request read access.</p>
                </div>
            </div>
        {/if}

        {#if step === 2}
            <div transition:fade={{ duration: 300 }} class="absolute max-w-xl w-full bg-surface-container-low rounded-2xl shadow-sm border border-outline-variant p-10 flex flex-col items-center text-center">
                <div class="w-20 h-20 rounded-full bg-blue-600/10 flex items-center justify-center mb-6 text-blue-600">
                    <span class="material-symbols-outlined text-4xl animate-spin" style="animation-duration: 3s;">radar</span>
                </div>
                
                <h2 class="font-headline text-2xl font-bold text-slate-900 mb-3 tracking-tight">Scanning your Repositories</h2>
                <p class="font-body text-base text-on-surface-variant mb-10 max-w-md mx-auto leading-relaxed">
                    We're analyzing your codebases to prepare your enterprise dashboard. This will only take a moment.
                </p>
                
                <div class="w-full max-w-md mb-4">
                    <div class="flex justify-between items-end mb-2">
                        <span class="font-label text-sm font-medium text-on-surface-variant">Progress</span>
                        <span class="font-label text-sm font-bold text-blue-600" data-testid="progress-text">{progress}%</span>
                    </div>
                    <div class="h-3 w-full bg-outline-variant rounded-full overflow-hidden">
                        <div class="h-full bg-blue-600 rounded-full transition-all duration-300 ease-out" style="width: {progress}%"></div>
                    </div>
                </div>
                
                <div class="flex items-center gap-2 text-on-surface-variant font-body text-sm mt-2">
                    <span class="material-symbols-outlined text-sm" style="font-variation-settings: 'FILL' 1;">folder_open</span>
                    <span>Found 12 repositories...</span>
                </div>
            </div>
        {/if}

        {#if step === 3}
            <div transition:fade={{ duration: 300 }} class="absolute bg-white/80 backdrop-blur-md rounded-2xl shadow-[0_8px_30px_rgb(0,0,0,0.04)] border border-slate-100 p-12 max-w-xl w-full text-center relative z-10">
                <div class="mb-8 relative inline-block">
                    <div class="absolute inset-0 bg-cyan-400 rounded-full blur-xl opacity-30 animate-pulse"></div>
                    <span class="material-symbols-outlined text-7xl text-cyan-500 relative z-10 drop-shadow-sm" style="font-variation-settings: 'FILL' 1;">
                        check_circle
                    </span>
                </div>
                
                <h2 class="text-3xl font-extrabold text-slate-900 tracking-tight mb-4 font-headline">Ready to Launch</h2>
                <p class="text-slate-500 text-lg leading-relaxed mb-10 max-w-md mx-auto">
                    Scanning complete. Your infrastructure is mapped, secured, and ready for centralized management.
                </p>
                
                <div class="flex flex-col sm:flex-row gap-4 justify-center items-center mb-8">
                    <button 
                        class="w-full sm:w-auto bg-slate-900 hover:bg-slate-800 text-white font-semibold py-3 px-8 rounded-lg shadow-sm transition-all active:scale-95 flex items-center justify-center gap-2"
                        onclick={enterDashboard}
                        data-testid="enter-dashboard-btn"
                        aria-label="Enter Dashboard"
                    >
                        <span>Enter Dashboard</span>
                        <span class="material-symbols-outlined text-sm">arrow_forward</span>
                    </button>
                    <button class="w-full sm:w-auto bg-white hover:bg-slate-50 text-slate-700 border border-slate-200 font-semibold py-3 px-8 rounded-lg transition-all active:scale-95 flex items-center justify-center gap-2">
                        <span class="material-symbols-outlined text-sm">download</span>
                        <span>Download Report</span>
                    </button>
                </div>
                
                <div class="pt-6 border-t border-slate-100 flex items-center justify-center gap-2 text-slate-400 text-sm font-medium">
                    <span class="material-symbols-outlined text-[16px]">lock</span>
                    <span>Session encrypted end-to-end. Connection secured.</span>
                </div>
            </div>
        {/if}
    </main>
</div>
