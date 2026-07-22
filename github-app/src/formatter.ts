import type { DiffReport, SubstrateConfig, CrossRepoCheckResponse } from './types.js';

function escapeMarkdown(text: string): string {
  // Escape pipes and backticks in markdown tables
  return text.replace(/\|/g, '\\|').replace(/`/g, '\\`');
}

function formatFooter(dashboardUrl?: string, owner?: string, repo?: string, prNumber?: number, diffId?: string): string {
  const poweredBy = `*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*`;
  if (dashboardUrl && diffId) {
    const link = `${dashboardUrl}/diff/${diffId}`;
    return `[View in Dashboard →](${link})\n${poweredBy}`;
  } else if (dashboardUrl && owner && repo && prNumber) {
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
  aiSafePatch?: string,
  diffId?: string,
  riskScore?: string
): string {
  const breakingCount = report.summary?.breaking_count || 0;
  const warningCount = report.summary?.warning_count || 0;
  const infoCount = report.summary?.info_count || 0;

  const breakingChanges = report.breaking_changes || [];
  const warningChanges = report.warnings || [];
  const infoChanges = report.safe_changes || [];

  let comment = '';

  if (riskScore) {
    let riskColor = '🟢';
    if (riskScore === 'CRITICAL') {
      riskColor = '🔴';
    } else if (riskScore === 'HIGH') {
      riskColor = '🟠';
    } else if (riskScore === 'MEDIUM') {
      riskColor = '🟡';
    }
    comment += `## ${riskColor} DEPLOYMENT RISK: ${riskScore}\n\n`;
  }

  if (config.mode === 'audit') {
    comment += `> ℹ️ **Substrate is running in Audit Mode.** This breaking change has been recorded, but this PR is NOT blocked.\n\n`;
  }

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
    comment += `\n### ⚠️ Compliance Flags\n\n`;
    comment += `| Type | Path | Message |\n`;
    comment += `|---|---|---|\n`;
    for (const alert of report.compliance_alerts) {
      comment += `| \`${escapeMarkdown(alert.compliance_type)}\` | \`${escapeMarkdown(alert.path)}\` | ${escapeMarkdown(alert.message)} |\n`;
    }
    comment += '\n';
  }

  if (dashboardUrl && diffId) {
      comment = `[🔍 View Interactive Diff](${dashboardUrl}/diff/${diffId})\n\n` + comment;
  }

  comment += `---\n`;
  comment += `${formatFooter(dashboardUrl, owner, repo, prNumber, diffId)}`;

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

    if (response.sla_breaches && response.sla_breaches.length > 0) {
      for (const breach of response.sla_breaches) {
        text += `> ⚠️ **SLA Breach**: \`${escapeMarkdown(breach.consumer)}\` requires ${breach.required_days} days notice for breaking changes.\n`;
      }
    }

    if (response.affected_customers !== undefined && response.affected_mrr !== undefined) {
      const formattedMRR = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(response.affected_mrr);
      text += `\n### 💳 Customer Impact\n\n> 🔴 **${response.affected_customers} paying customers** (Total MRR: ${formattedMRR}) rely on the endpoints broken by this PR.\n`;
    }
  } else {
    text += `\n> ✅ All registered consumers are compatible with this change.\n`;
  }


  return text;
}
