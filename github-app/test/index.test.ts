import { describe, it, expect, vi, beforeEach } from 'vitest';
import worker from '../src/index.js';
import * as githubClient from '../src/github-client.js';
const ctx: any = { waitUntil: vi.fn(), passThroughOnException: vi.fn() };

// Mock the github-client functions
vi.mock('../src/github-client.js', () => ({
  generateInstallationToken: vi.fn(),
  fetchFileContent: vi.fn(),
  postPRComment: vi.fn(),
  setCommitStatus: vi.fn(),
  fetchPRFiles: vi.fn()
}));

// Mock global fetch for the container service call
globalThis.fetch = vi.fn();

const MOCK_ENV = {
  GITHUB_APP_ID: '123',
  GITHUB_APP_PRIVATE_KEY: 'mock-key',
  GITHUB_WEBHOOK_SECRET: 'test-secret',
  CONTAINER_SERVICE_URL: 'http://localhost:8080'
};

async function signWebhook(payload: string, secret: string): Promise<string> {
  const encoder = new TextEncoder();
  const key = await crypto.subtle.importKey(
    'raw',
    encoder.encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  );
  const signature = await crypto.subtle.sign('HMAC', key, encoder.encode(payload));
  const hex = Array.from(new Uint8Array(signature))
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
  return `sha256=${hex}`;
}

describe('Worker Handler', () => {
  const ctx = {
    waitUntil: vi.fn(),
    passThroughOnException: vi.fn()
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (githubClient.generateInstallationToken as any).mockResolvedValue('mock-token');
  });

  it('1. Non-POST method -> 405', async () => {
    const request = new Request('http://localhost', { method: 'GET' });
    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(405);
  });

  it('2. Missing X-Hub-Signature-256 header -> 401', async () => {
    const request = new Request('http://localhost', { method: 'POST', body: '{}' });
    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(401);
  });

  it('3. Invalid HMAC signature -> 401', async () => {
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': 'sha256=invalid' },
      body: '{}'
    });
    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(401);
  });

  it('4. Valid signature, push event -> 200, no GitHub API calls made', async () => {
    const payload = JSON.stringify({});
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'push'
      },
      body: payload
    });
    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);
    expect(githubClient.generateInstallationToken).not.toHaveBeenCalled();
  });

  it('5. Valid signature, pull_request.closed -> 200, no GitHub API calls made', async () => {
    const payload = JSON.stringify({ action: 'closed' });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'pull_request'
      },
      body: payload
    });
    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);
    expect(githubClient.generateInstallationToken).not.toHaveBeenCalled();
  });

  it('6. Valid signature, pull_request.opened, no substrate.yaml -> missing config comment', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo', full_name: 'owner/repo' },
      pull_request: { head: { sha: 'headsha' }, base: { ref: 'main' }, number: 1 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'pull_request'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any).mockResolvedValueOnce(null); // local config
    (githubClient.fetchPRFiles as any).mockResolvedValueOnce(['some-random-file.txt']); // PR files

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.generateInstallationToken).toHaveBeenCalled();
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'pending', 'Substrate is checking...'
    );
    expect(githubClient.fetchFileContent).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'substrate.yaml', 'headsha'
    );
    expect(githubClient.fetchPRFiles).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1
    );
    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('substrate init')
    );
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'pending', expect.stringContaining('substrate.yaml')
    );
  });

  it('7. Valid signature, pull_request.opened, substrate.yaml present, base missing -> first time setup', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo', full_name: 'owner/repo' },
      pull_request: { head: { sha: 'headsha' }, base: { ref: 'main' }, number: 1 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'pull_request'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml') // substrate.yaml
      .mockResolvedValueOnce(null) // base.yaml
      .mockResolvedValueOnce('head content'); // head.yaml

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'success', expect.stringContaining('First-time setup detected')
    );
    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('Welcome to Substrate!')
    );
  });

  it('8. Valid signature, full happy path, Container Service returns 500 -> engine error', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo', full_name: 'owner/repo' },
      pull_request: { head: { sha: 'headsha' }, base: { ref: 'main' }, number: 1 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'pull_request'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: false,
      status: 500
    });

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', expect.stringContaining('engine error')
    );
  });

  it('9. Valid signature, full happy path, 0 breaking changes', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo', full_name: 'owner/repo' },
      pull_request: { head: { sha: 'headsha' }, base: { ref: 'main' }, number: 1 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'pull_request'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        breaking_changes: [], warnings: [], safe_changes: [],
        summary: { breaking_count: 0, warning_count: 0, info_count: 0 }
      })
    });

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('All Clear')
    );
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'success', 'All clear — no breaking changes'
    );
  });

  it('10. Valid signature, full happy path, 2 breaking changes', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo', full_name: 'owner/repo' },
      pull_request: { head: { sha: 'headsha' }, base: { ref: 'main' }, number: 1 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'pull_request'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockImplementation(async (url: string, init?: RequestInit) => {
      if (url.endsWith('/diff')) {
        return {
          ok: true,
          json: async () => ({
            breaking_changes: [{ rule_id: 'rule1', severity: 'BREAKING', path: 'path1', description: 'msg1' },
                       { rule_id: 'rule2', severity: 'BREAKING', path: 'path2', description: 'msg2' }],
            warnings: [], safe_changes: [],
            summary: { breaking_count: 2, warning_count: 0, info_count: 0 }
          })
        };
      }
      if (url.endsWith('/api/v1/ai/autofix')) {
        return {
          ok: true,
          json: async () => ({
            explanation: 'Mock AI explanation',
            safe_patch: 'mock:\n  safe: patch',
            patch_language: 'yaml',
            mock_mode: true
          })
        };
      }
      return { ok: true, json: async () => ({}) };
    });

    const crossRepoRes = { total_consumers: 0, broken_consumers: 0, is_safe: true, results: [] };
    (crossRepoCheck as any).mockResolvedValueOnce(crossRepoRes);

    const envWithReg = { ...MOCK_ENV, REGISTRY_API_URL: 'http://reg.api', REGISTRY_API_TOKEN: 'token' };

    const response = await worker.fetch(request, envWithReg as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('Mock AI explanation')
    );
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', expect.stringContaining('2 breaking change(s) detected')
    );
  });
});

import { parseConsumersFromYaml, syncToRegistry } from '../src/registry-client.js';

vi.mock('../src/registry-client.js', () => ({
  parseConsumersFromYaml: vi.fn(),
  syncToRegistry: vi.fn(),
  crossRepoCheck: vi.fn()
}));

import { crossRepoCheck } from '../src/registry-client.js';

describe('Worker Handler Push Event', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (githubClient.generateInstallationToken as any).mockResolvedValue('mock-token');
  });

  it('11. push to main with consumers -> calls syncToRegistry, returns 200', async () => {
    const payload = JSON.stringify({
      ref: 'refs/heads/main',
      after: 'sha123',
      deleted: false,
      installation: { id: 1 },
      repository: { owner: { login: 'owner', id: 99 }, name: 'repo', full_name: 'owner/repo', id: 456 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'push'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('consumers yaml content')
      .mockResolvedValueOnce('provider spec content');

    (parseConsumersFromYaml as any).mockResolvedValueOnce([{
      name: 'c', provider_repo: 'org/prov', schema_type: 'openapi', provider_spec_path: 's.yaml', provider_branch: 'main'
    }]);

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ owner: { login: 'org' }, name: 'prov', id: 123 })
    });

    (syncToRegistry as any).mockResolvedValueOnce({ synced: 1 });

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);

    expect(syncToRegistry).toHaveBeenCalledWith(undefined, undefined, expect.objectContaining({
      dependencies: expect.arrayContaining([expect.objectContaining({
        raw_content: 'provider spec content'
      })])
    }));
  });

  it('12. push to main, no substrate.yaml -> returns 200 Ignored', async () => {
    const payload = JSON.stringify({
      ref: 'refs/heads/main',
      after: 'sha123',
      deleted: false,
      installation: { id: 1 },
      repository: { owner: { login: 'owner', id: 99 }, name: 'repo', full_name: 'owner/repo', id: 456 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'push'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any).mockResolvedValueOnce(null);

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);
    expect(await response.text()).toBe('Ignored');
  });

  it('13. push to main, no consumers block -> returns 200 Ignored', async () => {
    const payload = JSON.stringify({
      ref: 'refs/heads/main',
      after: 'sha123',
      deleted: false,
      installation: { id: 1 },
      repository: { owner: { login: 'owner', id: 99 }, name: 'repo', full_name: 'owner/repo', id: 456 }
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'push'
      },
      body: payload
    });

    (githubClient.fetchFileContent as any).mockResolvedValueOnce('content');
    (parseConsumersFromYaml as any).mockResolvedValueOnce([]);

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);
    expect(await response.text()).toBe('Ignored');
  });

  it('14. push to non-main branch -> returns 200 Ignored', async () => {
    const payload = JSON.stringify({
      ref: 'refs/heads/feature',
      deleted: false
    });
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: {
        'X-Hub-Signature-256': sig,
        'X-GitHub-Event': 'push'
      },
      body: payload
    });

    const response = await worker.fetch(request, MOCK_ENV as any, ctx as any);
    expect(response.status).toBe(200);
    expect(await response.text()).toBe('Ignored');
  });
});

describe('Worker Handler Cross Repo PR Events', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (githubClient.generateInstallationToken as any).mockResolvedValue('mock-token');
  });

  const basePayload = {
    action: 'opened',
    installation: { id: 1 },
    repository: { owner: { login: 'owner' }, name: 'repo', full_name: 'owner/repo' },
    pull_request: { head: { sha: 'headsha' }, base: { ref: 'main' }, number: 1 }
  };

  it('PR with breaking change + 1 broken consumer → status check FAILS, comment includes cross-repo section', async () => {
    const payload = JSON.stringify(basePayload);
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': sig, 'X-GitHub-Event': 'pull_request' },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        breaking_changes: [{ rule_id: 'rule1', severity: 'BREAKING', path: 'path1', description: 'msg1' }],
        warnings: [], safe_changes: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      })
    });

    const crossRepoRes = {
      total_consumers: 1, broken_consumers: 1, is_safe: false,
      results: [{ consumer_repo: 'org/consumer', status: 'breaking', diff_report: { breaking_changes: [{ rule_id: 'rule', path: 'path', description: 'msg' }], summary: { breaking_count: 1 } } }]
    };
    (crossRepoCheck as any).mockResolvedValueOnce(crossRepoRes);

    const envWithReg = { ...MOCK_ENV, REGISTRY_API_URL: 'http://reg.api', REGISTRY_API_TOKEN: 'token' };
    const response = await worker.fetch(request, envWithReg as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('Cross-Repo Impact')
    );
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', expect.stringContaining('1 breaking change(s) detected — 1 consumer(s) affected')
    );
  });

  it('PR with breaking change + registry not configured (no REGISTRY_API_URL) → cross-repo check skipped', async () => {
    const payload = JSON.stringify(basePayload);
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': sig, 'X-GitHub-Event': 'pull_request' },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        breaking_changes: [{ rule_id: 'rule1', severity: 'BREAKING', path: 'path1', description: 'msg1' }],
        warnings: [], safe_changes: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      })
    });

    const envWithoutReg = { ...MOCK_ENV, REGISTRY_API_URL: '' };
    const response = await worker.fetch(request, envWithoutReg as any, ctx as any);
    expect(response.status).toBe(200);

    expect(crossRepoCheck).not.toHaveBeenCalled();
    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.not.stringContaining('Cross-Repo Impact')
    );
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', '1 breaking change(s) detected'
    );
  });

  it('PR with no single-repo breaking changes + 1 broken consumer → status check FAILS (cross-repo is the blocker)', async () => {
    const payload = JSON.stringify(basePayload);
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': sig, 'X-GitHub-Event': 'pull_request' },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ breaking_changes: [], warnings: [], safe_changes: [], summary: { breaking_count: 0, warning_count: 0, info_count: 0 } })
    });

    const crossRepoRes = {
      total_consumers: 1, broken_consumers: 1, is_safe: false,
      results: [{ consumer_repo: 'org/consumer', status: 'breaking', diff_report: { breaking_changes: [{ rule_id: 'rule', path: 'path', description: 'msg' }], summary: { breaking_count: 1 } } }]
    };
    (crossRepoCheck as any).mockResolvedValueOnce(crossRepoRes);

    const envWithReg = { ...MOCK_ENV, REGISTRY_API_URL: 'http://reg.api', REGISTRY_API_TOKEN: 'token' };
    const response = await worker.fetch(request, envWithReg as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', '1 downstream consumer(s) affected by this change'
    );
  });

  it('PR with no breaking changes + all consumers safe → status check PASSES, no cross-repo callout in comment', async () => {
    const payload = JSON.stringify(basePayload);
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': sig, 'X-GitHub-Event': 'pull_request' },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ breaking_changes: [], warnings: [], safe_changes: [], summary: { breaking_count: 0, warning_count: 0, info_count: 0 } })
    });

    const crossRepoRes = {
      total_consumers: 1, broken_consumers: 0, is_safe: true,
      results: [{ consumer_repo: 'org/consumer', status: 'safe', diff_report: { breaking_changes: [], summary: { breaking_count: 0 } } }]
    };
    (crossRepoCheck as any).mockResolvedValueOnce(crossRepoRes);

    const envWithReg = { ...MOCK_ENV, REGISTRY_API_URL: 'http://reg.api', REGISTRY_API_TOKEN: 'token' };
    const response = await worker.fetch(request, envWithReg as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'success', 'All clear — no breaking changes'
    );
  });

  it('PR comment includes dashboard link when DASHBOARD_URL is set', async () => {
    const payload = JSON.stringify(basePayload);
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': sig, 'X-GitHub-Event': 'pull_request' },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        breaking_changes: [{ rule_id: 'rule1', severity: 'BREAKING', path: 'path1', description: 'msg1' }],
        warnings: [], safe_changes: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      })
    });

    const envWithDashboard = { ...MOCK_ENV, DASHBOARD_URL: 'https://substrate.example.com', REGISTRY_API_URL: '' };
    const response = await worker.fetch(request, envWithDashboard as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('View in Dashboard →')
    );
    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('https://substrate.example.com/diff?owner=owner&repo=repo&pr=1')
    );
  });

  it('PR comment renders correctly when DASHBOARD_URL is not set', async () => {
    const payload = JSON.stringify(basePayload);
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': sig, 'X-GitHub-Event': 'pull_request' },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        breaking_changes: [{ rule_id: 'rule1', severity: 'BREAKING', path: 'path1', description: 'msg1' }],
        warnings: [], safe_changes: [],
        summary: { breaking_count: 1, warning_count: 0, info_count: 0 }
      })
    });

    const envWithoutDashboard = { ...MOCK_ENV, DASHBOARD_URL: undefined, REGISTRY_API_URL: '' };
    const response = await worker.fetch(request, envWithoutDashboard as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('Powered by [Substrate]')
    );
    expect(githubClient.postPRComment).not.toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('View in Dashboard')
    );
  });

  it('Registry API returns 500 → cross-repo silently skipped, PR handler does not throw', async () => {
    const payload = JSON.stringify(basePayload);
    const sig = await signWebhook(payload, MOCK_ENV.GITHUB_WEBHOOK_SECRET);
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': sig, 'X-GitHub-Event': 'pull_request' },
      body: payload
    });

    (githubClient.fetchFileContent as any)
      .mockResolvedValueOnce('base_schema: base.yaml\nhead_schema: head.yaml')
      .mockResolvedValueOnce('base content')
      .mockResolvedValueOnce('head content');

    (globalThis.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ breaking_changes: [], warnings: [], safe_changes: [], summary: { breaking_count: 0, warning_count: 0, info_count: 0 } })
    });

    const crossRepoRes = { total_consumers: 0, broken_consumers: 0, is_safe: true, results: [] };
    (crossRepoCheck as any).mockResolvedValueOnce(crossRepoRes);

    const envWithReg = { ...MOCK_ENV, REGISTRY_API_URL: 'http://reg.api', REGISTRY_API_TOKEN: 'token' };
    const response = await worker.fetch(request, envWithReg as any, ctx as any);
    expect(response.status).toBe(200);

    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'success', 'All clear — no breaking changes'
    );
  });

});
