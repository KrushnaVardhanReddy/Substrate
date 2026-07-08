import { describe, it, expect, vi, beforeEach } from 'vitest';
import worker from '../src/index.js';
import * as githubClient from '../src/github-client.js';

// Mock the github-client functions
vi.mock('../src/github-client.js', () => ({
  generateInstallationToken: vi.fn(),
  fetchFileContent: vi.fn(),
  postPRComment: vi.fn(),
  setCommitStatus: vi.fn()
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
  beforeEach(() => {
    vi.clearAllMocks();
    (githubClient.generateInstallationToken as any).mockResolvedValue('mock-token');
  });

  it('1. Non-POST method -> 405', async () => {
    const request = new Request('http://localhost', { method: 'GET' });
    const response = await worker.fetch(request, MOCK_ENV as any);
    expect(response.status).toBe(405);
  });

  it('2. Missing X-Hub-Signature-256 header -> 401', async () => {
    const request = new Request('http://localhost', { method: 'POST', body: '{}' });
    const response = await worker.fetch(request, MOCK_ENV as any);
    expect(response.status).toBe(401);
  });

  it('3. Invalid HMAC signature -> 401', async () => {
    const request = new Request('http://localhost', {
      method: 'POST',
      headers: { 'X-Hub-Signature-256': 'sha256=invalid' },
      body: '{}'
    });
    const response = await worker.fetch(request, MOCK_ENV as any);
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
    const response = await worker.fetch(request, MOCK_ENV as any);
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
    const response = await worker.fetch(request, MOCK_ENV as any);
    expect(response.status).toBe(200);
    expect(githubClient.generateInstallationToken).not.toHaveBeenCalled();
  });

  it('6. Valid signature, pull_request.opened, no substrate.yaml -> missing config comment', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo' },
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

    (githubClient.fetchFileContent as any).mockResolvedValueOnce(null);

    const response = await worker.fetch(request, MOCK_ENV as any);
    expect(response.status).toBe(200);

    expect(githubClient.generateInstallationToken).toHaveBeenCalled();
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'pending', 'Substrate is checking...'
    );
    expect(githubClient.fetchFileContent).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'substrate.yaml', 'headsha'
    );
    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('substrate init')
    );
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'pending', expect.stringContaining('substrate.yaml')
    );
  });

  it('7. Valid signature, pull_request.opened, substrate.yaml present, base missing -> config error', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo' },
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

    const response = await worker.fetch(request, MOCK_ENV as any);
    expect(response.status).toBe(200);

    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', expect.stringContaining('config error')
    );
    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('Config Error')
    );
  });

  it('8. Valid signature, full happy path, Container Service returns 500 -> engine error', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo' },
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

    const response = await worker.fetch(request, MOCK_ENV as any);
    expect(response.status).toBe(200);

    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', expect.stringContaining('engine error')
    );
  });

  it('9. Valid signature, full happy path, 0 breaking changes', async () => {
    const payload = JSON.stringify({
      action: 'opened',
      installation: { id: 1 },
      repository: { owner: { login: 'owner' }, name: 'repo' },
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
        breaking: [], warning: [], info: [],
        summary: { breaking_count: 0, warning_count: 0, info_count: 0 }
      })
    });

    const response = await worker.fetch(request, MOCK_ENV as any);
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
      repository: { owner: { login: 'owner' }, name: 'repo' },
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
        breaking: [{ rule: 'rule1', severity: 'BREAKING', path: 'path1', message: 'msg1' },
                   { rule: 'rule2', severity: 'BREAKING', path: 'path2', message: 'msg2' }],
        warning: [], info: [],
        summary: { breaking_count: 2, warning_count: 0, info_count: 0 }
      })
    });

    const response = await worker.fetch(request, MOCK_ENV as any);
    expect(response.status).toBe(200);

    expect(githubClient.postPRComment).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 1, expect.stringContaining('Breaking Changes Detected')
    );
    expect(githubClient.setCommitStatus).toHaveBeenCalledWith(
      'mock-token', 'owner', 'repo', 'headsha', 'failure', expect.stringContaining('2 breaking change(s) detected')
    );
  });
});
