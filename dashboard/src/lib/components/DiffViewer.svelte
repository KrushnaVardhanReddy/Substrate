// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	import { diffLines } from 'diff';

	let { before = '', after = '' } = $props<{ before?: string; after?: string }>();

	type Row = {
		leftNum: number | null;
		rightNum: number | null;
		leftContent: string | null;
		rightContent: string | null;
		leftType: 'normal' | 'removed' | 'empty';
		rightType: 'normal' | 'added' | 'empty';
	};

	let rows: Row[] = $derived.by(() => {
		const diffs = diffLines(before || '', after || '');
		let rows: Row[] = [];
		let leftNum = 1;
		let rightNum = 1;

		for (let i = 0; i < diffs.length; i++) {
			const chunk = diffs[i];
			const lines = chunk.value.endsWith('\n') ? chunk.value.slice(0, -1).split('\n') : chunk.value.split('\n');

			if (!chunk.added && !chunk.removed) {
				lines.forEach((line: string) => {
					rows.push({
						leftNum: leftNum++,
						rightNum: rightNum++,
						leftContent: line,
						rightContent: line,
						leftType: 'normal',
						rightType: 'normal'
					});
				});
			} else if (chunk.removed) {
				let nextChunk = (i + 1 < diffs.length && diffs[i + 1].added) ? diffs[i + 1] : null;
				let addedLines: string[] = [];
				if (nextChunk) {
					addedLines = nextChunk.value.endsWith('\n') ? nextChunk.value.slice(0, -1).split('\n') : nextChunk.value.split('\n');
					i++;
				}

				const maxLen = Math.max(lines.length, addedLines.length);
				for (let j = 0; j < maxLen; j++) {
					rows.push({
						leftNum: j < lines.length ? leftNum++ : null,
						rightNum: j < addedLines.length ? rightNum++ : null,
						leftContent: j < lines.length ? lines[j] : null,
						rightContent: j < addedLines.length ? addedLines[j] : null,
						leftType: j < lines.length ? 'removed' : 'empty',
						rightType: j < addedLines.length ? 'added' : 'empty'
					});
				}
			} else if (chunk.added) {
				lines.forEach((line: string) => {
					rows.push({
						leftNum: null,
						rightNum: rightNum++,
						leftContent: null,
						rightContent: line,
						leftType: 'empty',
						rightType: 'added'
					});
				});
			}
		}
		return rows;
	});

	import { diffSchemas } from '$lib/utils/wasmLoader';
	import { onMount } from 'svelte';

	let semanticDiff: any = $state(null);
	let isWasmLoading = $state(true);
	let wasmError: string | null = $state(null);

	onMount(async () => {
		if (!before || !after) {
			isWasmLoading = false;
			return;
		}
		try {
			isWasmLoading = true;
			semanticDiff = await diffSchemas(before, after);
		} catch (err: any) {
			console.error("WASM Semantic Diff Error:", err);
			wasmError = err.message || "Failed to analyze schema semantics.";
		} finally {
			isWasmLoading = false;
		}
	});
</script>

<div class="diff-viewer">
	{#if isWasmLoading}
		<div class="semantic-diff-panel loading">
			<span class="spinner"></span> Analyzing structural changes with WASM Engine...
		</div>
	{:else if wasmError}
		<div class="semantic-diff-panel error">
			<strong>WASM Engine Error:</strong> {wasmError}
		</div>
	{:else if semanticDiff}
		<div class="semantic-diff-panel">
			<div class="semantic-header">
				<h3>Semantic Diff Analysis</h3>
				<span class="badge {semanticDiff.summary?.overallSeverity?.toLowerCase() || 'safe'}">
					{semanticDiff.summary?.overallSeverity || 'Safe'}
				</span>
			</div>
			<div class="semantic-stats">
				<span class="stat breaking">{semanticDiff.summary?.breakingCount || 0} Breaking</span>
				<span class="stat warning">{semanticDiff.summary?.warningCount || 0} Warnings</span>
				<span class="stat safe">{semanticDiff.summary?.safeCount || 0} Safe</span>
			</div>
			{#if semanticDiff.breakingChanges?.length > 0}
				<div class="semantic-list">
					<h4>Breaking Changes</h4>
					<ul>
						{#each semanticDiff.breakingChanges as change}
							<li><strong>{change.path}:</strong> {change.description}</li>
						{/each}
					</ul>
				</div>
			{/if}
		</div>
	{/if}

	<div class="diff-header">
		<div class="diff-header-side">Base</div>
		<div class="diff-header-side">Head</div>
	</div>
	<div class="diff-body">
		{#each rows as row}
			<div class="diff-row">
				<div class="diff-side {row.leftType}">
					<div class="line-number">{row.leftNum !== null ? row.leftNum : ''}</div>
					<div class="line-marker">{row.leftType === 'removed' ? '-' : ' '}</div>
					<div class="line-content">{row.leftContent !== null ? row.leftContent : ''}</div>
				</div>
				<div class="diff-side {row.rightType}">
					<div class="line-number">{row.rightNum !== null ? row.rightNum : ''}</div>
					<div class="line-marker">{row.rightType === 'added' ? '+' : ' '}</div>
					<div class="line-content">{row.rightContent !== null ? row.rightContent : ''}</div>
				</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.diff-viewer {
		background: var(--bg-main, #000000);
		border: 1px solid var(--border, #333333);
		border-radius: 8px;
		overflow: hidden;
		font-family: Menlo, Monaco, 'Lucida Console', monospace;
		font-size: 13px;
		color: var(--text-main, #ffffff);
		display: flex;
		flex-direction: column;
	}

	.semantic-diff-panel {
		padding: 16px;
		background: var(--bg-hover, #111);
		border-bottom: 1px solid var(--border, #333);
	}
	.semantic-diff-panel.loading { color: #888; display: flex; align-items: center; gap: 8px; }
	.semantic-diff-panel.error { color: #ef4444; background: rgba(239, 68, 68, 0.1); }
	
	.semantic-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
	.semantic-header h3 { margin: 0; font-size: 14px; font-weight: 600; color: #fff; }
	
	.badge { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; text-transform: uppercase; }
	.badge.breaking { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
	.badge.warning { background: rgba(245, 158, 11, 0.2); color: #f59e0b; }
	.badge.safe { background: rgba(16, 185, 129, 0.2); color: #10b981; }

	.semantic-stats { display: flex; gap: 16px; margin-bottom: 12px; }
	.stat { font-size: 13px; font-weight: 500; }
	.stat.breaking { color: #ef4444; }
	.stat.warning { color: #f59e0b; }
	.stat.safe { color: #10b981; }

	.semantic-list h4 { margin: 0 0 8px 0; font-size: 13px; color: #ef4444; }
	.semantic-list ul { margin: 0; padding-left: 20px; color: #ddd; font-size: 12px; }
	.semantic-list li { margin-bottom: 4px; }

	.spinner { width: 14px; height: 14px; border: 2px solid #555; border-top-color: #fff; border-radius: 50%; animation: spin 1s linear infinite; }
	@keyframes spin { to { transform: rotate(360deg); } }

	.diff-header {
		display: flex;
		background: var(--bg-header, #111111);
		border-bottom: 1px solid var(--border, #333333);
		font-weight: bold;
		color: var(--text-muted, #666666);
	}

	.diff-header-side {
		flex: 1;
		padding: 8px 16px;
	}

	.diff-header-side:first-child {
		border-right: 1px solid var(--border, #333333);
	}

	.diff-body {
		overflow-x: auto;
		display: flex;
		flex-direction: column;
	}

	.diff-row {
		display: flex;
		min-width: 100%;
	}

	.diff-side {
		flex: 1;
		display: flex;
		min-width: 0;
	}

	.diff-side:first-child {
		border-right: 1px solid var(--border, #333333);
	}

	.line-number {
		width: 45px;
		min-width: 45px;
		text-align: right;
		padding-right: 10px;
		color: var(--text-muted, #666666);
		user-select: none;
		border-right: 1px solid var(--border, #333333);
		opacity: 0.5;
	}

	.line-marker {
		width: 25px;
		min-width: 25px;
		text-align: center;
		user-select: none;
		font-weight: bold;
	}

	.line-content {
		padding-left: 8px;
		white-space: pre;
	}

	.removed {
		background: rgba(239, 68, 68, 0.15);
		color: var(--danger, #EF4444);
	}

	.removed .line-number {
		color: var(--danger, #EF4444);
		background: rgba(239, 68, 68, 0.1);
		opacity: 0.8;
		border-right-color: rgba(239, 68, 68, 0.3);
	}

	.added {
		background: rgba(16, 185, 129, 0.15);
		color: #10B981;
	}

	.added .line-number {
		color: #10B981;
		background: rgba(16, 185, 129, 0.1);
		opacity: 0.8;
		border-right-color: rgba(16, 185, 129, 0.3);
	}

	.empty {
		background: transparent;
	}

	.normal {
		background: transparent;
	}
</style>