<script lang="ts">
	import * as Diff from 'diff';

	let { before = '', after = '' } = $props<{ before?: string; after?: string }>();

	type DiffLine = {
		type: 'normal' | 'add' | 'delete' | 'empty';
		num?: number;
		text: string;
	};

	type DiffRow = {
		left: DiffLine;
		right: DiffLine;
	};

	let diffRows = $derived.by(() => {
		const parts = Diff.diffLines(before || '', after || '');
		let rows: DiffRow[] = [];
		let leftNum = 1;
		let rightNum = 1;

		for (let i = 0; i < parts.length; i++) {
			const part = parts[i];
			// parts.value includes newlines. Remove the trailing newline for accurate line count
			const lines = part.value.replace(/\n$/, '').split('\n');
            if (lines.length === 1 && lines[0] === '') {
                // If it's completely empty, skip it unless it's the only part
                if (parts.length === 1 && before === '' && after === '') {
                    // Do nothing
                } else {
                    continue;
                }
            }

			if (part.removed) {
				// Check if the next part is an addition to align them side-by-side
				if (i + 1 < parts.length && parts[i + 1].added) {
					const nextPart = parts[i + 1];
					const nextLines = nextPart.value.replace(/\n$/, '').split('\n');
					const maxLines = Math.max(lines.length, nextLines.length);

					for (let j = 0; j < maxLines; j++) {
						rows.push({
							left: j < lines.length
								? { type: 'delete', num: leftNum++, text: lines[j] }
								: { type: 'empty', text: '' },
							right: j < nextLines.length
								? { type: 'add', num: rightNum++, text: nextLines[j] }
								: { type: 'empty', text: '' }
						});
					}
					i++; // Skip the next part as we've processed it
				} else {
					// Just removed lines
					for (const line of lines) {
						rows.push({
							left: { type: 'delete', num: leftNum++, text: line },
							right: { type: 'empty', text: '' }
						});
					}
				}
			} else if (part.added) {
				// Just added lines
				for (const line of lines) {
					rows.push({
						left: { type: 'empty', text: '' },
						right: { type: 'add', num: rightNum++, text: line }
					});
				}
			} else {
				// Unchanged lines
				for (const line of lines) {
					rows.push({
						left: { type: 'normal', num: leftNum++, text: line },
						right: { type: 'normal', num: rightNum++, text: line }
					});
				}
			}
		}
		return rows;
	});
</script>

<div class="diff-container">
	<div class="diff-table">
		{#each diffRows as row}
			<div class="diff-row">
				<!-- Left Side -->
				<div class="diff-cell left-num {row.left.type}">
					{row.left.num ?? ''}
				</div>
				<div class="diff-cell left-marker {row.left.type}">
					{#if row.left.type === 'delete'}-{/if}
				</div>
				<div class="diff-cell left-content {row.left.type}">
					{row.left.text}
				</div>

				<!-- Right Side -->
				<div class="diff-cell right-num {row.right.type}">
					{row.right.num ?? ''}
				</div>
				<div class="diff-cell right-marker {row.right.type}">
					{#if row.right.type === 'add'}+{/if}
				</div>
				<div class="diff-cell right-content {row.right.type}">
					{row.right.text}
				</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.diff-container {
		border: 1px solid var(--border);
		border-radius: 8px;
		background-color: var(--bg-card);
		overflow-x: auto;
		font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
		font-size: 13px;
		line-height: 1.5;
	}

	.diff-table {
		display: flex;
		flex-direction: column;
		min-width: 800px;
	}

	.diff-row {
		display: grid;
		grid-template-columns: 50px 20px 1fr 50px 20px 1fr;
	}

	.diff-cell {
		padding: 2px 4px;
		white-space: pre-wrap;
		word-break: break-all;
	}

	.left-num, .right-num {
		color: var(--text-muted);
		text-align: right;
		user-select: none;
		border-right: 1px solid var(--border);
		padding-right: 8px;
	}

	.left-marker, .right-marker {
		user-select: none;
		text-align: center;
		font-weight: bold;
	}

	/* Normal */
	.normal {
		background-color: transparent;
		color: var(--text-main);
	}

	/* Add */
	.add {
		background-color: color-mix(in srgb, var(--accent) 15%, transparent);
	}
	.left-num.add, .right-num.add {
		background-color: color-mix(in srgb, var(--accent) 25%, transparent);
	}
	.left-marker.add, .right-marker.add {
		color: var(--accent);
	}

	/* Delete */
	.delete {
		background-color: color-mix(in srgb, var(--danger) 15%, transparent);
	}
	.left-num.delete, .right-num.delete {
		background-color: color-mix(in srgb, var(--danger) 25%, transparent);
	}
	.left-marker.delete, .right-marker.delete {
		color: var(--danger);
	}

	/* Empty */
	.empty {
		background-color: transparent;
	}
	.left-num.empty, .right-num.empty {
		border-right: 1px solid var(--border);
	}
</style>
