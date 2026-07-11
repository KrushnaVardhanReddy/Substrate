import type { DiffReport, SubstrateConfig, CrossRepoCheckResponse } from './types.js';

function escapeMarkdown(text: string): string {
  // Escape pipes and backticks in markdown tables
  return text.replace(/\|/g, '\\|').replace(/`/g, '\\`');
}

function formatFooter(dashboardUrl?: string, owner?: string, repo?: string, prNumber?: number): string {
  const poweredBy = `*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*`;
  if (dashboardUrl && owner && repo && prNumber) {
    const link = `${dashboardUrl}/diff?owner=${owner}&repo=${repo}&pr=${prNumber}`;
    return `[View in Dashboard →](${link})\n${poweredBy}`;
  }
  return poweredBy;
}

export function formatPRComment(
  report: DiffReport,
  config: SubstrateConfig,
  dashboardUrl?: string,
  owner?: string,
  repo?: string,
  prNumber?: number,
  aiExplanation?: string,
  aiSafePatch?: string
): string {
  const breakingCount = report.summary?.breaking_count || 0;
  const warningCount = report.summary?.warning_count || 0;
  const infoCount = report.summary?.info_count || 0;

  const breakingChanges = report.breaking_changes || [];
  const warningChanges = report.warnings || [];
  const infoChanges = report.safe_changes || [];

  let comment = '';

  if (breakingCount > 0) {
    comment += `## 🔴 Substrate — Breaking Changes Detected\n\n`;
    const schemaName = (report as any).schema_type === 'sql' ? 'SQL' : 'OpenAPI';
    comment += `This PR introduces **${breakingCount} breaking change(s)** to your ${schemaName} contract.\n`;
    comment += `Consumers of this API may break if this PR is merged without coordination.\n\n`;

    comment += `| Severity | Rule | Path | Description |\n`;
    comment += `|---|---|---|---|\n`;

    for (const change of breakingChanges) {
      comment += `| 🔴 BREAKING | \`${escapeMarkdown(change.rule_id || '')}\` | \`${escapeMarkdown(change.path || '')}\` | ${escapeMarkdown(change.description || '')} |\n`;
    }
    for (const change of warningChanges) {
      comment += `| 🟡 WARNING | \`${escapeMarkdown(change.rule_id || '')}\` | \`${escapeMarkdown(change.path || '')}\` | ${escapeMarkdown(change.description || '')} |\n`;
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

    if (aiExplanation) {
      comment += `\n---\n### 🤖 AI Impact Analysis\n> ${aiExplanation}\n`;
    }
    if (aiSafePatch) {
      comment += `\n### 🔧 Suggested Safe Remediation\nApply the following change to unblock this PR:\n\`\`\`yaml\n${aiSafePatch}\n\`\`\`\n`;
    }

    comment += `\n*To acknowledge a breaking change, add an override to your \`substrate.yaml\`.*\n\n`;
  } else if (warningCount > 0) {
    comment += `## 🟡 Substrate — Warnings Only\n\n`;
    comment += `No breaking changes, but ${warningCount} warning(s) detected.\n\n`;

    comment += `| Severity | Rule | Path | Description |\n`;
    comment += `|---|---|---|---|\n`;

    for (const change of warningChanges) {
      comment += `| 🟡 WARNING | \`${escapeMarkdown(change.rule_id || '')}\` | \`${escapeMarkdown(change.path || '')}\` | ${escapeMarkdown(change.description || '')} |\n`;
    }

    comment += '\n';

    if (infoCount > 0) {
      let summaryStr = infoCount === 1 ? `1 informational change` : `${infoCount} informational changes`;
      comment += `<details>\n`;
      comment += `<summary>${summaryStr} (click to expand)</summary>\n`;
      comment += `...\n`;
      comment += `</details>\n\n`;
    }

  } else {
    comment += `## ✅ Substrate — All Clear\n\n`;
    comment += `No breaking changes detected in this PR. Safe to merge. 🎉\n\n`;

  }

  if (report.compliance_alerts && report.compliance_alerts.length > 0) {
    comment += `\n### 🛡️ Compliance & PII Alerts\n\n`;
    comment += `| Type | Path | Message |\n`;
    comment += `|---|---|---|\n`;
    for (const alert of report.compliance_alerts) {
      comment += `| \`${escapeMarkdown(alert.compliance_type)}\` | \`${escapeMarkdown(alert.path)}\` | ${escapeMarkdown(alert.message)} |\n`;
    }
    comment += '\n';
  }

  comment += `---\n`;
  comment += `${formatFooter(dashboardUrl, owner, repo, prNumber)}`;

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



export function getCommitStatusState(
  report: DiffReport,
  config: SubstrateConfig,
  crossRepo?: CrossRepoCheckResponse
): 'success' | 'failure' {
  const breakingCount = report.summary?.breaking_count || 0;

  if (crossRepo?.is_safe === false) {
    return 'failure';
  }

  if (breakingCount === 0) {
    return 'success';
  }

  const onBreakingChange = config.on_breaking_change || 'block';

  if (onBreakingChange === 'warn') {
    return 'success';
  }

  return 'failure';
}

export function getCommitStatusDescription(
  report: DiffReport,
  crossRepo?: CrossRepoCheckResponse
): string {
  const breakingCount = report.summary?.breaking_count || 0;

  if (breakingCount > 0 && crossRepo?.is_safe === false) {
    return `${breakingCount} breaking change(s) detected — ${crossRepo.broken_consumers} consumer(s) affected`;
  }

  if (breakingCount > 0) {
    return `${breakingCount} breaking change(s) detected`;
  }

  if (crossRepo?.is_safe === false) {
    return `${crossRepo.broken_consumers} downstream consumer(s) affected by this change`;
  }

  return 'All clear — no breaking changes';
}

export function formatCrossRepoImpact(response: CrossRepoCheckResponse): string {
  if (response.total_consumers === 0) {
    return '';
  }

  let text = `\n---\n\n## 🌐 Cross-Repo Impact\n\nThis change affects **${response.total_consumers} registered consumer(s)**:\n\n| Consumer | Status | Breaking Changes |\n|---|---|---|\n`;

  for (const result of response.results) {
    let statusText = '';
    if (result.status === 'breaking') statusText = '❌ BREAKING';
    else if (result.status === 'safe') statusText = '✅ Safe';
    else if (result.status === 'warning') statusText = '⚠️ Warning';
    else statusText = '❓ Unknown';

    let breakingText = '';
    const breakingCount = result.diff_report.summary.breaking_count || 0;
    if (breakingCount === 0) {
      breakingText = 'No breaking changes detected';
    } else {
      if (result.diff_report.breaking_changes && result.diff_report.breaking_changes.length > 0) {
        const firstChange = result.diff_report.breaking_changes[0];
        breakingText = `\`${escapeMarkdown(firstChange.path || '')}\` — ${escapeMarkdown(firstChange.description || '')}`;
      } if (breakingCount > 1) {
        breakingText += ` (+${breakingCount - 1} more)`;
      }
    }

    text += `| \`${escapeMarkdown(result.consumer_repo)}\` | ${statusText} | ${breakingText} |\n`;
  }

  if (response.broken_consumers > 0) {
    const brokenRepos = response.results
      .filter(r => r.status === 'breaking')
      .map(r => `\`${escapeMarkdown(r.consumer_repo)}\``)
      .join(', ');

    text += `\n> ⚠️ **Action required:** Coordinate with the ${brokenRepos} team before merging.\n> The \`substrate/breaking-changes\` check is now **FAILING**.\n`;
  } else {
    text += `\n> ✅ All registered consumers are compatible with this change.\n`;
  }

  return text;
}
