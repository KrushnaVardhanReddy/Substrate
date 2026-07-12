import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { validateWebhookSignature, parsePREvent, parsePushEvent, parseInstallationRepositoriesEvent, parseInstallationEvent } from '../src/webhook.js';

describe('webhook validation', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('validateWebhookSignature returns true for valid signature', async () => {
    const importKeyMock = vi.spyOn(crypto.subtle, 'importKey').mockResolvedValue('mock-key' as any);
    const verifyMock = vi.spyOn(crypto.subtle, 'verify').mockResolvedValue(true as any);

    const result = await validateWebhookSignature('sha256=123456', 'body', 'secret');
    expect(result).toBe(true);
    expect(verifyMock).toHaveBeenCalled();
  });

  it('validateWebhookSignature returns false for invalid signature prefix', async () => {
    const result = await validateWebhookSignature('sha1=123456', 'body', 'secret');
    expect(result).toBe(false);
  });

  it('validateWebhookSignature returns false when crypto.verify fails', async () => {
    const importKeyMock = vi.spyOn(crypto.subtle, 'importKey').mockResolvedValue('mock-key' as any);
    const verifyMock = vi.spyOn(crypto.subtle, 'verify').mockResolvedValue(false as any);

    const result = await validateWebhookSignature('sha256=123456', 'body', 'secret');
    expect(result).toBe(false);
  });
});

describe('PR event parsing', () => {
  it('parsePREvent returns null for push event', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'push' });
    const result = parsePREvent(headers, '{}');
    expect(result).toBeNull();
  });

  it('parsePREvent returns null for closed action', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'pull_request' });
    const body = JSON.stringify({ action: 'closed' });
    const result = parsePREvent(headers, body);
    expect(result).toBeNull();
  });

  it('parsePREvent returns PREvent for opened action', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'pull_request' });
    const body = JSON.stringify({
      action: 'opened',
      repository: {
        name: 'repo-name',
        owner: { login: 'owner-login' }
      },
      pull_request: {
        number: 42,
        head: { sha: 'abcdef' },
        base: { ref: 'main' }
      },
      installation: { id: 123 }
    });
    const result = parsePREvent(headers, body);
    expect(result).toEqual({
      owner: 'owner-login',
      repo: 'repo-name',
      prNumber: 42,
      headSha: 'abcdef',
      baseBranch: 'main',
      installationId: 123
    });
  });

  it('parsePREvent returns PREvent for synchronize action', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'pull_request' });
    const body = JSON.stringify({
      action: 'synchronize',
      repository: {
        name: 'repo-name',
        owner: { login: 'owner-login' }
      },
      pull_request: {
        number: 42,
        head: { sha: 'abcdef' },
        base: { ref: 'main' }
      },
      installation: { id: 123 }
    });
    const result = parsePREvent(headers, body);
    expect(result).not.toBeNull();
  });
});

describe('Push event parsing', () => {
  it('returns PushEvent for push to main', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'push' });
    const body = JSON.stringify({
      ref: 'refs/heads/main',
      after: 'abcdef',
      deleted: false,
      installation: { id: 123 },
      repository: {
        name: 'repo',
        full_name: 'owner/repo',
        id: 456,
        owner: { login: 'owner', id: 789 }
      }
    });
    const result = parsePushEvent(headers, body);
    expect(result).toEqual({
      ref: 'refs/heads/main',
      after: 'abcdef',
      installationId: 123,
      owner: 'owner',
      repo: 'repo',
      fullName: 'owner/repo',
      githubRepoId: 456,
      installationOrgId: 789
    });
  });

  it('returns null for push to feature branch', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'push' });
    const body = JSON.stringify({
      ref: 'refs/heads/feature',
      deleted: false
    });
    const result = parsePushEvent(headers, body);
    expect(result).toBeNull();
  });

  it('returns null for tag push', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'push' });
    const body = JSON.stringify({
      ref: 'refs/tags/v1.0',
      deleted: false
    });
    const result = parsePushEvent(headers, body);
    expect(result).toBeNull();
  });

  it('returns null for delete push', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'push' });
    const body = JSON.stringify({
      ref: 'refs/heads/main',
      deleted: true
    });
    const result = parsePushEvent(headers, body);
    expect(result).toBeNull();
  });

  it('returns null for non-push event', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'pull_request' });
    const body = JSON.stringify({ ref: 'refs/heads/main' });
    const result = parsePushEvent(headers, body);
    expect(result).toBeNull();
  });
});

describe('parseInstallationRepositoriesEvent', () => {
  it('should return null for incorrect event type', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'push' });
    expect(parseInstallationRepositoriesEvent(headers, '{}')).toBeNull();
  });

  it('should parse valid installation_repositories event', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'installation_repositories' });
    const body = JSON.stringify({
      action: 'added',
      installation: { id: 123 },
      repositories_added: [{ id: 1, name: 'repo', full_name: 'owner/repo', owner: { login: 'owner', id: 2 } }],
      repositories_removed: []
    });

    const result = parseInstallationRepositoriesEvent(headers, body);
    expect(result).not.toBeNull();
    expect(result?.action).toBe('added');
    expect(result?.repositories_added.length).toBe(1);
  });
});

describe('parseInstallationEvent', () => {
  it('should return null for incorrect event type', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'push' });
    expect(parseInstallationEvent(headers, '{}')).toBeNull();
  });

  it('should parse valid installation event', () => {
    const headers = new Headers({ 'X-GitHub-Event': 'installation' });
    const body = JSON.stringify({
      action: 'created',
      installation: { id: 123 },
      repositories: [{ id: 1, name: 'repo', full_name: 'owner/repo', owner: { login: 'owner', id: 2 } }]
    });

    const result = parseInstallationEvent(headers, body);
    expect(result).not.toBeNull();
    expect(result?.action).toBe('created');
    expect(result?.repositories?.length).toBe(1);
  });
});
