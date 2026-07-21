import type { DiffReport, SubstrateConfig, CrossRepoCheckResponse } from './types.js';

export async function getCommitStatusState(
  containerUrl: string,
  report: DiffReport,
  config: SubstrateConfig,
  crossRepo?: CrossRepoCheckResponse
): Promise<'success' | 'failure'> {
  if (config.mode === 'audit') {
    return 'success';
  }

  const breakingCount = report.summary?.breaking_count || 0;
  const warningCount = report.summary?.warning_count || 0;
  const isCrossRepoSafe = crossRepo?.is_safe ?? true;
  const gate = config.quality_gate || 'standard';

  try {
    const response = await fetch(`${containerUrl}/evaluate-gate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        gate: gate,
        breaking_count: breakingCount,
        warning_count: warningCount,
        cross_repo_safe: isCrossRepoSafe
      })
    });

    if (response.ok) {
      const data = await response.json() as { pass: boolean };

      if (data.pass) {
        // Fallback to legacy warn mode check if gate passes but breaking changes exist
        const onBreakingChange = config.on_breaking_change || 'block';
        if (breakingCount > 0 && onBreakingChange === 'warn') {
           return 'success';
        }
        return 'success';
      }
      return 'failure';
    } else {
      console.error(`Gate evaluation failed with status ${response.status}`);
      // Fallback logic if container fails
      if (gate === 'permissive' && !isCrossRepoSafe) return 'failure';
      if (gate === 'strict' && (breakingCount > 0 || warningCount > 0 || !isCrossRepoSafe)) return 'failure';
      if (breakingCount > 0 || !isCrossRepoSafe) return 'failure';
      return 'success';
    }
  } catch (err) {
    console.error("Error evaluating gate:", err);
    // Fallback logic if container fails
    if (gate === 'permissive' && !isCrossRepoSafe) return 'failure';
    if (gate === 'strict' && (breakingCount > 0 || warningCount > 0 || !isCrossRepoSafe)) return 'failure';
    if (breakingCount > 0 || !isCrossRepoSafe) return 'failure';
    return 'success';
  }
}

export function getCommitStatusDescription(
  report: DiffReport,
  crossRepo?: CrossRepoCheckResponse,
  config?: SubstrateConfig
): string {
  const breakingCount = report.summary?.breaking_count || 0;
  const gate = config?.quality_gate || 'standard';

  let description = '';

  if (breakingCount > 0 && crossRepo?.is_safe === false) {
    description = `${breakingCount} breaking change(s) detected — ${crossRepo.broken_consumers} consumer(s) affected`;
  } else if (breakingCount > 0) {
    description = `${breakingCount} breaking change(s) detected`;
  } else if (crossRepo?.is_safe === false) {
    description = `${crossRepo.broken_consumers} downstream consumer(s) affected by this change`;
  } else if (gate === 'strict' && (report.summary?.warning_count || 0) > 0) {
    description = `${report.summary?.warning_count} warning(s) detected (Strict Mode)`;
  } else {
    description = 'All clear — no breaking changes';
  }

  if (config?.mode === 'audit' && breakingCount > 0) {
    description += ' (Audit Mode: Non-blocking)';
  } else if (config?.tier) {
    description += ` [Tier: ${config.tier}]`;
  } else if (gate !== 'standard') {
    description += ` [Gate: ${gate}]`;
  }

  return description;
}
