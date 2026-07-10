import { describe, it, expect, vi, beforeEach } from 'vitest';
import { load } from './+page';

vi.mock('$env/dynamic/public', () => ({
	env: {
		PUBLIC_API_URL: 'http://test-api',
		PUBLIC_API_TOKEN: 'test-token'
	}
}));

describe('Org Page Load Function', () => {
	beforeEach(() => {
		vi.resetAllMocks();
	});

	it('returns fetched repos successfully', async () => {
		const mockRepos = [{ name: 'repo1', full_name: 'test/repo1' }];
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			json: vi.fn().mockResolvedValue(mockRepos)
		});

		const result = await load({ fetch: mockFetch as any, params: { org: 'test' } } as any);

		expect(mockFetch).toHaveBeenCalledWith('http://test-api/api/v1/repos/test', {
			headers: { Authorization: 'Bearer test-token' }
		});
		expect(result).toEqual({ repos: mockRepos });
	});

	it('returns empty array on failed fetch', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			statusText: 'Not Found'
		});

		const result = await load({ fetch: mockFetch as any, params: { org: 'test' } } as any);

		expect(result).toEqual({ repos: [] });
	});

	it('returns empty array on network error', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));

		const result = await load({ fetch: mockFetch as any, params: { org: 'test' } } as any);

		expect(result).toEqual({ repos: [] });
	});
});
