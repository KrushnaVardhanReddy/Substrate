import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { validateWebhookSignature, parsePREvent } from '../src/webhook';

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
