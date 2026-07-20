import { describe, it, expect } from 'vitest';
import {
  formatPRComment,
  formatMissingConfigComment,
  getCommitStatusState,
  getCommitStatusDescription,
  formatCrossRepoImpact
} from '../src/formatter.js';
import type { DiffReport, SubstrateConfig, CrossRepoCheckResponse } from '../src/types.js';

describe('Formatter', () => {
  const emptyReport: DiffReport = {
    breaking_changes: [],
    warnings: [],
    safe_changes: [],
    summary: { breaking_count: 0, warning_count: 0, info_count: 0 }
  };
  const emptyConfig: SubstrateConfig = {};

  describe('formatPRComment', () => {
    it('returns all clear if no breaking or warnings', () => {
      const comment = formatPRComment(emptyReport, emptyConfig);
      expect(comment).toContain('All Clear');
      expect(comment).toContain('No breaking changes detected');
      expect(comment).toContain('Powered by [Substrate]');
    });

    it('returns breaking changes properly', () => {
      const report: DiffReport = {
        breaking_changes: [
          { rule_id: 'R1', path: '/path', description: 'desc', severity: 'BREAKING' }
        ],
        warnings: [],
        safe_changes: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('Breaking Changes Detected');
      expect(comment).toContain('R1');
      expect(comment).toContain('/path');
      expect(comment).toContain('desc');
    });

    it('returns warnings properly', () => {
      const report: DiffReport = {
        breaking_changes: [],
        warnings: [
          { rule_id: 'W1', path: '/path2', description: 'desc2', severity: 'WARNING' }
        ],
        safe_changes: [],
        summary: { breaking_count: 0, warning_count: 1, info_count: 0 }
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('Warnings Only');
      expect(comment).toContain('W1');
    });

    it('escapes markdown properly', () => {
      const report: DiffReport = {
        breaking_changes: [
          { rule_id: 'R|1', path: '`path`', description: 'desc', severity: 'BREAKING' }
        ],
        warnings: [],
        safe_changes: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('R\\|1');
      expect(comment).toContain('\\`path\\`');
    });

    it('shows compliance alerts', () => {
      const report: DiffReport = {
        ...emptyReport,
        compliance_alerts: [
          { path: '/x', compliance_type: 'PII', message: 'alert msg' }
        ]
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('Compliance Flags');
      expect(comment).toContain('PII');
      expect(comment).toContain('alert msg');
    });

    it('includes AI explanation and safe patch', () => {
      const report: DiffReport = {
        ...emptyReport,
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      };
      const aiExplanation = "This is an explanation";
      const aiSafePatch = "patch content";
      const comment = formatPRComment(report, emptyConfig, undefined, undefined, undefined, undefined, aiExplanation, aiSafePatch);
      expect(comment).toContain('AI Impact Analysis');
      expect(comment).toContain(aiExplanation);
      expect(comment).toContain('Suggested Safe Remediation');
      expect(comment).toContain(aiSafePatch);
    });

    it('includes risk score if provided', () => {
      const comment = formatPRComment(emptyReport, emptyConfig, undefined, undefined, undefined, undefined, undefined, undefined, undefined, 'HIGH');
      expect(comment).toContain('🟠 DEPLOYMENT RISK: HIGH');
    });

    it('handles critical risk score color', () => {
      const comment = formatPRComment(emptyReport, emptyConfig, undefined, undefined, undefined, undefined, undefined, undefined, undefined, 'CRITICAL');
      expect(comment).toContain('🔴 DEPLOYMENT RISK: CRITICAL');
    });

    it('includes risk score if provided', () => {
      const comment = formatPRComment(emptyReport, emptyConfig, undefined, undefined, undefined, undefined, undefined, undefined, undefined, 'HIGH');
      expect(comment).toContain('🟠 DEPLOYMENT RISK: HIGH');
    });

    it('handles critical risk score color', () => {
      const comment = formatPRComment(emptyReport, emptyConfig, undefined, undefined, undefined, undefined, undefined, undefined, undefined, 'CRITICAL');
      expect(comment).toContain('🔴 DEPLOYMENT RISK: CRITICAL');
    });
  });


  describe('formatMissingConfigComment', () => {
    it('returns setup guide', () => {
      const comment = formatMissingConfigComment();
      expect(comment).toContain('substrate.yaml');
      expect(comment).toContain('substrate init');
    });
  });

  describe('getCommitStatusState', () => {
    it('returns success for audit mode', () => {
      expect(getCommitStatusState(emptyReport, { mode: 'audit' })).toBe('success');
    });

    it('returns success for no breaking changes', () => {
      expect(getCommitStatusState(emptyReport, emptyConfig)).toBe('success');
    });

    it('returns failure for breaking changes', () => {
      const report: DiffReport = { ...emptyReport, summary: { breaking_count: 1, warning_count: 0, info_count: 0 } };
      expect(getCommitStatusState(report, emptyConfig)).toBe('failure');
    });

    it('returns success for breaking changes with warn on_breaking_change', () => {
      const report: DiffReport = { ...emptyReport, summary: { breaking_count: 1, warning_count: 0, info_count: 0 } };
      expect(getCommitStatusState(report, { on_breaking_change: 'warn' })).toBe('success');
    });

    it('returns failure if crossRepo check fails', () => {
      const cr: CrossRepoCheckResponse = { total_consumers: 1, broken_consumers: 1, is_safe: false, results: [] };
      expect(getCommitStatusState(emptyReport, emptyConfig, cr)).toBe('failure');
    });
  });

  describe('getCommitStatusDescription', () => {
    it('returns all clear', () => {
      expect(getCommitStatusDescription(emptyReport)).toBe('All clear — no breaking changes');
    });

    it('returns breaking count', () => {
      const report: DiffReport = { ...emptyReport, summary: { breaking_count: 2, warning_count: 0, info_count: 0 } };
      expect(getCommitStatusDescription(report)).toBe('2 breaking change(s) detected');
    });

    it('returns cross repo failure when no direct breaking', () => {
      const cr: CrossRepoCheckResponse = { total_consumers: 1, broken_consumers: 2, is_safe: false, results: [] };
      expect(getCommitStatusDescription(emptyReport, cr)).toBe('2 downstream consumer(s) affected by this change');
    });

    it('combines direct and cross repo failure', () => {
      const report: DiffReport = { ...emptyReport, summary: { breaking_count: 3, warning_count: 0, info_count: 0 } };
      const cr: CrossRepoCheckResponse = { total_consumers: 1, broken_consumers: 4, is_safe: false, results: [] };
      expect(getCommitStatusDescription(report, cr)).toBe('3 breaking change(s) detected — 4 consumer(s) affected');
    });

    it('appends Audit Mode text', () => {
      const report: DiffReport = { ...emptyReport, summary: { breaking_count: 1, warning_count: 0, info_count: 0 } };
      expect(getCommitStatusDescription(report, undefined, { mode: 'audit' })).toContain('(Audit Mode: Non-blocking)');
    });
  });

  describe('formatCrossRepoImpact', () => {
    it('returns empty if 0 consumers', () => {
      expect(formatCrossRepoImpact({ total_consumers: 0, broken_consumers: 0, is_safe: true, results: [] })).toBe('');
    });

    it('formats safe consumers', () => {
      const cr: CrossRepoCheckResponse = {
        total_consumers: 1,
        broken_consumers: 0,
        is_safe: true,
        results: [
          {
            consumer_repo: 'org/c1',
            status: 'safe',
            diff_report: { breaking_changes: [], warnings: [], safe_changes: [], summary: { breaking_count: 0, warning_count: 0, info_count: 0 } }
          }
        ]
      };
      const text = formatCrossRepoImpact(cr);
      expect(text).toContain('Cross-Repo Impact');
      expect(text).toContain('org/c1');
      expect(text).toContain('Safe');
    });

    it('formats breaking consumers', () => {
      const cr: CrossRepoCheckResponse = {
        total_consumers: 1,
        broken_consumers: 1,
        is_safe: false,
        results: [
          {
            consumer_repo: 'org/c2',
            status: 'breaking',
            diff_report: {
              breaking_changes: [{ rule_id: '1', path: 'p', description: 'd' }],
              warnings: [],
              safe_changes: [],
              summary: { breaking_count: 2, warning_count: 0, info_count: 0 }
            }
          }
        ]
      };
      const text = formatCrossRepoImpact(cr);
      expect(text).toContain('org/c2');
      expect(text).toContain('BREAKING');
      expect(text).toContain('(+1 more)');
      expect(text).toContain('Action required');
    });

    it('shows SLA breaches', () => {
      const cr: CrossRepoCheckResponse = {
        total_consumers: 1,
        broken_consumers: 1,
        is_safe: false,
        results: [
          {
            consumer_repo: 'org/c2',
            status: 'breaking',
            diff_report: {
              breaking_changes: [], warnings: [], safe_changes: [], summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
            }
          }
        ],
        sla_breaches: [
          { consumer: 'org/c2', required_days: 14 }
        ]
      };
      const text = formatCrossRepoImpact(cr);
      expect(text).toContain('SLA Breach');
      expect(text).toContain('requires 14 days notice');
    });
  });
});
