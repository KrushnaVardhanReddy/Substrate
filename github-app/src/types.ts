export interface DiffChange {
  rule: string;
  severity: 'BREAKING' | 'WARNING' | 'INFO';
  path: string;
  message: string;
}

export interface DiffSummary {
  breaking_count: number;
  warning_count: number;
  info_count: number;
}

export interface DiffReport {
  breaking: DiffChange[];
  warning: DiffChange[];
  info: DiffChange[];
  summary: DiffSummary;
}

export interface SubstrateConfig {
  on_breaking_change?: 'block' | 'warn';  // default: 'block'
  base_schema?: string;
  head_schema?: string;
}
