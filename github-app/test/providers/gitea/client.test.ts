import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GiteaProvider } from '../../../src/providers/gitea/client.js';

describe('GiteaProvider', () => {
  const mockApiUrl = 'https://gitea.example.com';
  const mockToken = 'test-token';
  let provider: GiteaProvider;

  beforeEach(() => {
    provider = new GiteaProvider(mockApiUrl, mockToken);
    globalThis.fetch = vi.fn();
  });

  it('fetchFileContent should return decoded content', async () => {
    const mockContent = 'hello world';
    const mockBase64 = btoa(mockContent);

    vi.mocked(globalThis.fetch).mockResolvedValueOnce({
      status: 200,
      ok: true,
      json: async () => ({ type: 'file', content: mockBase64 })
    } as any);

    const result = await provider.fetchFileContent('owner', 'repo', 'path/file.txt', 'main');
    expect(result).toBe(mockContent);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      'https://gitea.example.com/api/v1/repos/owner/repo/contents/path/file.txt?ref=main',
      expect.objectContaining({
        headers: {
          'Authorization': 'token test-token',
          'Accept': 'application/json'
        }
      })
    );
  });

  it('fetchFileContent should return null for 404', async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce({
      status: 404,
      ok: false,
      json: async () => ({})
    } as any);

    const result = await provider.fetchFileContent('owner', 'repo', 'path/file.txt', 'main');
    expect(result).toBeNull();
  });

  it('postPRComment should patch existing bot comment', async () => {
    vi.mocked(globalThis.fetch)
      .mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => [{ id: 123, body: '## Substrate existing' }]
      } as any)
      .mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => ({})
      } as any);

    await provider.postPRComment('owner', 'repo', 1, 'new body');

    expect(globalThis.fetch).toHaveBeenCalledTimes(2);
    expect(globalThis.fetch).toHaveBeenLastCalledWith(
      'https://gitea.example.com/api/v1/repos/owner/repo/issues/comments/123',
      expect.objectContaining({
        method: 'PATCH',
        body: JSON.stringify({ body: 'new body' })
      })
    );
  });

  it('postPRComment should post new comment if none exists', async () => {
    vi.mocked(globalThis.fetch)
      .mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => []
      } as any)
      .mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => ({})
      } as any);

    await provider.postPRComment('owner', 'repo', 1, 'new body');

    expect(globalThis.fetch).toHaveBeenCalledTimes(2);
    expect(globalThis.fetch).toHaveBeenLastCalledWith(
      'https://gitea.example.com/api/v1/repos/owner/repo/issues/1/comments',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ body: 'new body' })
      })
    );
  });

  it('setCommitStatus should POST to status endpoint', async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce({
      status: 200,
      ok: true,
      json: async () => ({})
    } as any);

    await provider.setCommitStatus('owner', 'repo', 'sha123', 'success', 'desc');

    expect(globalThis.fetch).toHaveBeenCalledWith(
      'https://gitea.example.com/api/v1/repos/owner/repo/statuses/sha123',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          state: 'success',
          description: 'desc',
          context: 'substrate/breaking-changes'
        })
      })
    );
  });

  it('fetchPRFiles should return filenames', async () => {
    vi.mocked(globalThis.fetch).mockResolvedValueOnce({
      status: 200,
      ok: true,
      json: async () => [{ filename: 'file1.ts' }, { filename: 'file2.ts' }]
    } as any);

    const result = await provider.fetchPRFiles('owner', 'repo', 1);
    expect(result).toEqual(['file1.ts', 'file2.ts']);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      'https://gitea.example.com/api/v1/repos/owner/repo/pulls/1/files?limit=100',
      expect.any(Object)
    );
  });
});
