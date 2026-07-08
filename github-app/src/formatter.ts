import type { DiffReport, SubstrateConfig } from './types.js';

function escapeMarkdown(text: string): string {
  // Escape pipes and backticks in markdown tables
  return text.replace(/\|/g, '\\|').replace(/`/g, '\\`');
}

export function formatPRComment(report: DiffReport, config: SubstrateConfig): string {
  const breakingCount = report.summary?.breaking_count || 0;
  const warningCount = report.summary?.warning_count || 0;
  const infoCount = report.summary?.info_count || 0;

  const breakingChanges = report.breaking || [];
  const warningChanges = report.warning || [];
  const infoChanges = report.info || [];

  let comment = '';

  if (breakingCount > 0) {
    comment += `## 🔴 Substrate — Breaking Changes Detected\n\n`;
    comment += `This PR introduces **${breakingCount} breaking change(s)** to your OpenAPI contract.\n`;
    comment += `Consumers of this API may break if this PR is merged without coordination.\n\n`;

    comment += `| Severity | Rule | Path |\n`;
    comment += `|---|---|---|\n`;

    for (const change of breakingChanges) {
      comment += `| 🔴 BREAKING | \`${escapeMarkdown(change.rule)}\` | \`${escapeMarkdown(change.path)}\` |\n`;
    }
    for (const change of warningChanges) {
      comment += `| 🟡 WARNING | \`${escapeMarkdown(change.rule)}\` | \`${escapeMarkdown(change.path)}\` |\n`;
    }

    comment += '\n';

    if (infoCount > 0) {
      // The spec uses "1 warning, 5 informational changes (click to expand)"
      let warningStr = warningCount === 1 ? '1 warning' : `${warningCount} warnings`;
      let infoStr = infoCount === 1 ? '1 informational change' : `${infoCount} informational changes`;

      let summaryStr = warningCount > 0 ? `${warningStr}, ${infoStr}` : infoStr;

      comment += `<details>\n`;
      comment += `<summary>${summaryStr} (click to expand)</summary>\n`;
      comment += `...\n`;
      comment += `</details>\n\n`;
    }

    comment += `---\n`;
    comment += `*To acknowledge a breaking change, add an override to your \`substrate.yaml\`.*\n`;
    comment += `*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*`;

  } else if (warningCount > 0) {
    comment += `## 🟡 Substrate — Warnings Only\n\n`;
    comment += `No breaking changes, but ${warningCount} warning(s) detected.\n\n`;

    comment += `| Severity | Rule | Path |\n`;
    comment += `|---|---|---|\n`;

    for (const change of warningChanges) {
      comment += `| 🟡 WARNING | \`${escapeMarkdown(change.rule)}\` | \`${escapeMarkdown(change.path)}\` |\n`;
    }

    comment += '\n';

    if (infoCount > 0) {
      let summaryStr = infoCount === 1 ? `1 informational change` : `${infoCount} informational changes`;
      comment += `<details>\n`;
      comment += `<summary>${summaryStr} (click to expand)</summary>\n`;
      comment += `...\n`;
      comment += `</details>\n\n`;
    }

    comment += `---\n`;
    comment += `*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*`;

  } else {
    comment += `## ✅ Substrate — All Clear\n\n`;
    comment += `No breaking changes detected in this PR. Safe to merge. 🎉\n\n`;

    comment += `---\n`;
    comment += `*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*`;
  }

  return comment;
}

export function formatMissingConfigComment(): string {
  return `## 👋 Substrate is installed but not configured

We noticed this repo has OpenAPI/SQL/GraphQL spec files but no \`substrate.yaml\`.

Run this in your repo root to activate breaking change protection:
\`\`\`bash
substrate init
\`\`\`
This generates a \`substrate.yaml\` in 30 seconds. [View setup guide →](https://github.com/KrushnaVardhanReddy/Substrate#quick-start)

*This comment will not appear again once \`substrate.yaml\` is added.*

---
*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*`;
}

export function getCommitStatusState(report: DiffReport, config: SubstrateConfig): 'success' | 'failure' {
  const breakingCount = report.summary?.breaking_count || 0;

  if (breakingCount === 0) {
    return 'success';
  }

  const onBreakingChange = config.on_breaking_change || 'block';

  if (onBreakingChange === 'warn') {
    return 'success';
  }

  return 'failure';
}

export function getCommitStatusDescription(report: DiffReport): string {
  const breakingCount = report.summary?.breaking_count || 0;

  if (breakingCount === 0) {
    return 'All clear — no breaking changes';
  }

  return `${breakingCount} breaking change(s) detected`;
}
