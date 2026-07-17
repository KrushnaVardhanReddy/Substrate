import { describe, it, expect } from 'vitest';
import { parsePREvent, parsePushEvent } from '../../../src/providers/gitea/webhook.js';

describe('Gitea Webhook Parser', () => {
  it('parsePREvent should return StandardPREvent for opened PR', () => {
    const headers = new Headers({ 'X-Gitea-Event': 'pull_request' });
    const payload = JSON.stringify({
      action: 'opened',
      repository: {
        owner: { username: 'test-owner' },
        name: 'test-repo',
        full_name: 'test-owner/test-repo'
      },
      pull_request: {
        number: 42,
        head: { sha: 'head-sha' },
        base: { ref: 'main' }
      }
    });

    const result = parsePREvent(headers, payload);
    expect(result).toEqual({
      owner: 'test-owner',
      repo: 'test-repo',
      fullName: 'test-owner/test-repo',
      prNumber: 42,
      headSha: 'head-sha',
      baseBranch: 'main',
      installationId: 0
    });
  });

  it('parsePREvent should return null for unsupported action', () => {
    const headers = new Headers({ 'X-Gitea-Event': 'pull_request' });
    const payload = JSON.stringify({ action: 'closed' });
    const result = parsePREvent(headers, payload);
    expect(result).toBeNull();
  });

  it('parsePushEvent should return StandardPushEvent for push to main', () => {
    const headers = new Headers({ 'X-Forgejo-Event': 'push' });
    const payload = JSON.stringify({
      ref: 'refs/heads/main',
      after: 'push-sha',
      repository: {
        id: 101,
        owner: { username: 'test-owner', id: 202 },
        name: 'test-repo',
        full_name: 'test-owner/test-repo'
      }
    });

    const result = parsePushEvent(headers, payload);
    expect(result).toEqual({
      ref: 'refs/heads/main',
      after: 'push-sha',
      installationId: 0,
      owner: 'test-owner',
      repo: 'test-repo',
      fullName: 'test-owner/test-repo',
      githubRepoId: 101,
      installationOrgId: 202
    });
  });

  it('parsePushEvent should return null if deleted is true', () => {
    const headers = new Headers({ 'X-Gitea-Event': 'push' });
    const payload = JSON.stringify({ deleted: true });
    const result = parsePushEvent(headers, payload);
    expect(result).toBeNull();
  });
});
