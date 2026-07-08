import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fetchFileContent, postPRComment, setCommitStatus } from '../src/github-client';

describe('github client', () => {
  let fetchMock: any;

  beforeEach(() => {
    fetchMock = vi.fn();
    globalThis.fetch = fetchMock;
    vi.resetAllMocks();
  });

  describe('fetchFileContent', () => {
    it('returns decoded string for 200 response with base64 content', async () => {
      fetchMock.mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => ({ type: 'file', content: 'SGVsbG8gV29ybGQ=' }) // "Hello World"
      });

      const result = await fetchFileContent('token', 'owner', 'repo', 'path', 'ref');
      expect(result).toBe('Hello World');
    });

    it('returns null for 404 response', async () => {
      fetchMock.mockResolvedValueOnce({
        status: 404,
        ok: false
      });

      const result = await fetchFileContent('token', 'owner', 'repo', 'path', 'ref');
      expect(result).toBeNull();
    });
  });

  describe('postPRComment', () => {
    it('calls POST when no existing bot comment found', async () => {
      // Mock GET comments
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => [
          { user: { type: 'User' }, body: 'human comment' }
        ]
      });
      // Mock POST comment
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201
      });

      await postPRComment('token', 'owner', 'repo', 1, 'new comment');

      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(fetchMock.mock.calls[1][1].method).toBe('POST');
    });

    it('calls PATCH when existing bot comment found', async () => {
      // Mock GET comments
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: async () => [
          {
            user: { type: 'Bot' },
            body: '## Substrate header',
            url: 'https://api.github.com/comments/123'
          }
        ]
      });
      // Mock PATCH comment
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200
      });

      await postPRComment('token', 'owner', 'repo', 1, 'updated comment');

      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(fetchMock.mock.calls[1][1].method).toBe('PATCH');
      expect(fetchMock.mock.calls[1][0]).toBe('https://api.github.com/comments/123');
    });
  });

  describe('setCommitStatus', () => {
    it('sends correct JSON body with context', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201
      });

      await setCommitStatus('token', 'owner', 'repo', 'sha123', 'success', 'desc');

      expect(fetchMock).toHaveBeenCalledTimes(1);
      const call = fetchMock.mock.calls[0];
      expect(call[1].method).toBe('POST');
      const body = JSON.parse(call[1].body);
      expect(body).toEqual({
        state: 'success',
        description: 'desc',
        context: 'substrate/breaking-changes'
      });
    });
  });
});
