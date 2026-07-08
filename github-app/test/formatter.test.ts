import { describe, it, expect } from 'vitest';
import {
  formatPRComment,
  formatMissingConfigComment,
  getCommitStatusState,
  getCommitStatusDescription
} from '../src/formatter.js';
import type { DiffReport, SubstrateConfig } from '../src/types.js';

describe('formatter', () => {
  const emptyReport: DiffReport = {
    breaking: [],
    warning: [],
    info: [],
    summary: { breaking_count: 0, warning_count: 0, info_count: 0 }
  };

  const emptyConfig: SubstrateConfig = {};

  describe('formatPRComment', () => {
    it('0 breaking, 0 warning -> output contains "All Clear" and "✅", no table', () => {
      const comment = formatPRComment(emptyReport, emptyConfig);
      expect(comment).toContain('✅');
      expect(comment).toContain('All Clear');
      expect(comment).not.toContain('|---|---|---|');
      expect(comment).toContain('Powered by [Substrate]');
    });

    it('0 breaking, 2 warnings -> output contains "Warnings Only" and "🟡"', () => {
      const report: DiffReport = {
        breaking: [],
        warning: [
          { rule: 'rule-1', path: 'path-1', severity: 'WARNING', message: '' },
          { rule: 'rule-2', path: 'path-2', severity: 'WARNING', message: '' }
        ],
        info: [],
        summary: { breaking_count: 0, warning_count: 2, info_count: 0 }
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('🟡');
      expect(comment).toContain('Warnings Only');

      const tableRows = comment.split('\n').filter((line: string) => line.startsWith('| 🟡 WARNING |'));
      expect(tableRows).toHaveLength(2);
      expect(comment).toContain('Powered by [Substrate]');
    });

    it('3 breaking, 2 warning -> output contains "Breaking Changes Detected" and "🔴"', () => {
      const report: DiffReport = {
        breaking: [
          { rule: 'b-1', path: 'p-1', severity: 'BREAKING', message: '' },
          { rule: 'b-2', path: 'p-2', severity: 'BREAKING', message: '' },
          { rule: 'b-3', path: 'p-3', severity: 'BREAKING', message: '' }
        ],
        warning: [
          { rule: 'w-1', path: 'p-4', severity: 'WARNING', message: '' },
          { rule: 'w-2', path: 'p-5', severity: 'WARNING', message: '' }
        ],
        info: [],
        summary: { breaking_count: 3, warning_count: 2, info_count: 0 }
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('🔴');
      expect(comment).toContain('Breaking Changes Detected');

      const breakingRows = comment.split('\n').filter((line: string) => line.startsWith('| 🔴 BREAKING |'));
      const warningRows = comment.split('\n').filter((line: string) => line.startsWith('| 🟡 WARNING |'));

      expect(breakingRows).toHaveLength(3);
      expect(warningRows).toHaveLength(2);
      expect(comment).toContain('Powered by [Substrate]');
    });

    it('handles empty arrays properly', () => {
      // Intentionally omitting arrays if possible, but TypeScript interface requires them.
      // Test the `|| []` defaults by casting.
      const comment = formatPRComment({ summary: { breaking_count: 1 } } as any, emptyConfig);
      expect(comment).toContain('Breaking Changes Detected');
    });

    it('handles rule name with backticks without breaking markdown table', () => {
      const report: DiffReport = {
        breaking: [
          { rule: 'rule`name', path: 'path', severity: 'BREAKING', message: '' }
        ],
        warning: [],
        info: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('| 🔴 BREAKING | `rule\\`name` | `path` |');
    });

    it('handles path with pipe character | without breaking markdown table', () => {
      const report: DiffReport = {
        breaking: [
          { rule: 'rule', path: 'path|pipe', severity: 'BREAKING', message: '' }
        ],
        warning: [],
        info: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      };
      const comment = formatPRComment(report, emptyConfig);
      expect(comment).toContain('| 🔴 BREAKING | `rule` | `path\\|pipe` |');
    });
  });

  describe('formatMissingConfigComment', () => {
    it('contains "substrate init" and "substrate.yaml"', () => {
      const comment = formatMissingConfigComment();
      expect(comment).toContain('substrate init');
      expect(comment).toContain('substrate.yaml');
      expect(comment).toContain('Powered by [Substrate]');
    });
  });

  describe('getCommitStatusState', () => {
    it('0 breaking -> success', () => {
      expect(getCommitStatusState(emptyReport, emptyConfig)).toBe('success');
    });

    it('1 breaking, block -> failure', () => {
      const report = { ...emptyReport, summary: { ...emptyReport.summary, breaking_count: 1 } };
      expect(getCommitStatusState(report, { on_breaking_change: 'block' })).toBe('failure');
    });

    it('1 breaking, warn -> success', () => {
      const report = { ...emptyReport, summary: { ...emptyReport.summary, breaking_count: 1 } };
      expect(getCommitStatusState(report, { on_breaking_change: 'warn' })).toBe('success');
    });

    it('1 breaking, undefined config -> failure (default)', () => {
      const report = { ...emptyReport, summary: { ...emptyReport.summary, breaking_count: 1 } };
      expect(getCommitStatusState(report, {})).toBe('failure');
    });
  });

  describe('getCommitStatusDescription', () => {
    it('0 breaking -> "All clear"', () => {
      const desc = getCommitStatusDescription(emptyReport);
      expect(desc).toContain('All clear');
    });

    it('3 breaking -> contains "3"', () => {
      const report = { ...emptyReport, summary: { ...emptyReport.summary, breaking_count: 3 } };
      const desc = getCommitStatusDescription(report);
      expect(desc).toContain('3');
      expect(desc).toContain('breaking change(s) detected');
    });
  });
});
