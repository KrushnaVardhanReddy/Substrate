<script>
	let { data } = $props();

	let diffLines = $derived.by(() => {
		const oldLines = (data.oldText || '').split('\n');
		const newLines = (data.newText || '').split('\n');
		const result = [];

		let i = 0;
		let j = 0;

		while (i < oldLines.length || j < newLines.length) {
			if (i < oldLines.length && j < newLines.length && oldLines[i] === newLines[j]) {
				result.push({ type: 'normal', oldNum: i + 1, newNum: j + 1, text: oldLines[i] });
				i++;
				j++;
			} else if (j < newLines.length && (i >= oldLines.length || !oldLines.slice(i).includes(newLines[j]))) {
				result.push({ type: 'add', oldNum: '', newNum: j + 1, text: newLines[j] });
				j++;
			} else {
				result.push({ type: 'delete', oldNum: i + 1, newNum: '', text: oldLines[i] });
				i++;
			}
		}
		return result;
	});
</script>

<div class="vercel-diff">
	<div class="diff-header">
		<span class="dot red"></span>
		<span class="dot yellow"></span>
		<span class="dot green"></span>
		<span class="title">Schema Diff Viewer</span>
	</div>
	<div class="diff-viewport">
		{#each diffLines as line}
			<div class="diff-line {line.type}">
				<div class="line-number old">{line.oldNum}</div>
				<div class="line-number new">{line.newNum}</div>
				<div class="line-marker">
					{#if line.type === 'add'}+{/if}
					{#if line.type === 'delete'}-{/if}
				</div>
				<div class="line-content">{line.text}</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.vercel-diff {
		--bg-main: #000000;
		--bg-header: #111111;
		--border: #333333;
		--text-muted: #666666;
		--text-main: #ffffff;
		--add-bg: rgba(0, 224, 140, 0.15);
		--add-text: #00e08c;
		--del-bg: rgba(255, 0, 85, 0.15);
		--del-text: #ff0055;
		font-family: Menlo, Monaco, Lucida Console, monospace;
		font-size: 13px;
		background: var(--bg-main);
		border: 1px solid var(--border);
		border-radius: 8px;
		overflow: hidden;
		max-width: 800px;
		margin: 20px auto;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
	}
	.diff-header {
		background: var(--bg-header);
		border-bottom: 1px solid var(--border);
		padding: 12px;
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.dot { width: 10px; height: 10px; border-radius: 50%; }
	.red { background: #ff5f56; }
	.yellow { background: #ffbd2e; }
	.green { background: #27c93f; }
	.title { color: var(--text-muted); margin-left: 8px; font-size: 12px; }
	.diff-viewport { overflow-x: auto; max-height: 450px; }
	.diff-line {
		display: grid; grid-template-columns: 45px 45px 25px 1fr; align-items: center;
		color: var(--text-main); white-space: pre; line-height: 20px;
	}
	.line-number {
		color: var(--text-muted); text-align: right; padding-right: 12px; user-select: none; border-right: 1px solid var(--border);
	}
	.line-marker { text-align: center; user-select: none; font-weight: bold; }
	.line-content { padding-left: 8px; }
	.add { background: var(--add-bg); color: var(--add-text); }
	.add .line-number { color: rgba(0, 224, 140, 0.4); }
	.delete { background: var(--del-bg); color: var(--del-text); }
	.delete .line-number { color: rgba(255, 0, 85, 0.4); }
</style>
