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
  on_breaking_change?: 'block' | 'warn';
  base_schema?: string;
  head_schema?: string;
}

export interface Env {
  GITHUB_APP_ID: string;
  GITHUB_APP_PRIVATE_KEY: string;
  GITHUB_WEBHOOK_SECRET: string;
  CONTAINER_SERVICE_URL: string;
  REGISTRY_API_URL: string;
  REGISTRY_API_TOKEN: string;
}

export interface ConsumerEntry {
  name: string;
  provider_repo: string;
  schema_type: string;
  provider_spec_path: string;
  provider_branch: string;
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
