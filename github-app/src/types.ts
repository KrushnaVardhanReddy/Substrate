export interface DiffChange {
  rule_id: string;
  severity: 'BREAKING' | 'WARNING' | 'INFO';
  path: string;
  description: string;
}

export interface DiffSummary {
  breaking_count: number;
  warning_count: number;
  info_count: number;
}

export interface ComplianceAlert {
  path: string;
  compliance_type: string;
  message: string;
}

export interface DiffReport {
  breaking_changes: DiffChange[];
  warnings: DiffChange[];
  safe_changes: DiffChange[];
  summary: DiffSummary;
  compliance_alerts?: ComplianceAlert[];
}

export interface SubstrateConfig {
  on_breaking_change?: 'block' | 'warn';
  base_schema?: string;
  head_schema?: string;
  mode?: 'strict' | 'legacy' | 'audit';
}

export interface Env {
  GITHUB_APP_ID: string;
  GITHUB_APP_PRIVATE_KEY: string;
  GITHUB_TOKEN?: string;
  GITHUB_WEBHOOK_SECRET: string;
  GITEA_API_URL?: string;
  GITEA_TOKEN?: string;
  CONTAINER_SERVICE_URL: string;
  REGISTRY_API_URL: string;
  REGISTRY_API_TOKEN: string;
  DASHBOARD_URL?: string;
}

export interface ConsumerEntry {
  name: string;
  provider_repo: string;
  schema_type: string;
  provider_spec_path: string;
  provider_branch: string;
  required_notice_days?: number;
}

export interface SubstrateConsumerConfig {
  consumers?: ConsumerEntry[];
}

export interface SyncDependency {
  provider_repo: string;
  provider_github_repo_id: number;
  schema_type: string;
  spec_path: string;
  branch: string;
  raw_content: string;
  required_notice_days?: number;
}

export interface SyncRequest {
  installation_id: number;
  org: string;
  consumer_repo: string;
  consumer_github_repo_id: number;
  commit_sha: string;
  dependencies: SyncDependency[];
}

export interface PushEvent {
  ref: string;
  after: string;
  installationId: number;
  owner: string;
  repo: string;
  fullName: string;
  githubRepoId: number;
  installationOrgId: number;
}

export interface CrossRepoBreakingChange {
  rule_id: string;
  path: string;
  description: string;
}

export interface CrossRepoDiffSummary {
  breaking_count: number;
  warning_count: number;
  info_count: number;
}

export interface CrossRepoDiffReport {
  breaking_changes: CrossRepoBreakingChange[];
  warnings: CrossRepoBreakingChange[];
  safe_changes: CrossRepoBreakingChange[];
  summary: CrossRepoDiffSummary;
}

export interface ConsumerResult {
  consumer_repo: string;
  status: 'breaking' | 'safe' | 'warning' | 'unknown';
  diff_report: CrossRepoDiffReport;
}

export interface CrossRepoCheckRequest {
  installation_id: number;
  org: string;
  provider_repo: string;
  head_schema_content: string;
  schema_type: string;
  config_content?: string;
}

export interface SLABreach {
  consumer: string;
  required_days: number;
}

export interface CrossRepoCheckResponse {
  total_consumers: number;
  broken_consumers: number;
  is_safe: boolean;
  results: ConsumerResult[];
  sla_breaches?: SLABreach[];
}

export interface AIAutofixRequest {
  provider_repo: string;
  schema_type: string;
  current_schema: string;
  proposed_schema: string;
  breaking_changes: any[];
}

export interface AIAutofixResponse {
  explanation: string;
  safe_patch: string;
  patch_language: string;
  mock_mode: boolean;
}

export interface InstallationRepo {
  id: number;
  name: string;
  full_name: string;
  owner: {
    login: string;
    id: number;
  };
}

export interface InstallationRepositoriesEvent {
  action: 'added' | 'removed';
  installation: {
    id: number;
  };
  repositories_added: InstallationRepo[];
  repositories_removed: InstallationRepo[];
}

export interface InstallationEvent {
  action: 'created' | 'deleted' | 'suspend' | 'unsuspend' | 'new_permissions_accepted';
  installation: {
    id: number;
  };
  repositories?: InstallationRepo[];
}

export interface VCSClient {
  fetchFileContent(owner: string, repo: string, path: string, ref: string): Promise<string | null>;
  postPRComment(owner: string, repo: string, prNumber: number, body: string): Promise<void>;
  setCommitStatus(owner: string, repo: string, sha: string, state: 'success' | 'failure' | 'pending', description: string): Promise<void>;
  fetchPRFiles(owner: string, repo: string, prNumber: number): Promise<string[]>;
}

export interface StandardPREvent {
  owner: string;
  repo: string;
  fullName: string;
  prNumber: number;
  headSha: string;
  baseBranch: string;
  installationId: number;
}

export interface StandardPushEvent {
  ref: string;
  after: string;
  installationId: number;
  owner: string;
  repo: string;
  fullName: string;
  githubRepoId: number;
  installationOrgId: number;
}
