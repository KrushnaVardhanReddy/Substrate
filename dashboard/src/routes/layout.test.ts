import { describe, it, expect, vi } from 'vitest';
import { load } from './+layout';

describe('Layout Load Function', () => {
	it('fetches repositories and returns them', async () => {
		const mockRepos = [{ id: '1', name: 'repo-1', full_name: 'org/repo-1' }];

		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => mockRepos
		});

		// @ts-ignore - Partial LayoutLoadEvent mock
		const result: any = await load({ fetch: mockFetch });

		expect(mockFetch).toHaveBeenCalledWith('http://localhost:8090/api/v1/repos/myorg', {
			headers: {
				Authorization: 'Bearer test-token'
			}
		});

		expect(result).toEqual({ repos: mockRepos });
	});

	it('returns empty array on failed fetch', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			statusText: 'Not Found'
		});

		// @ts-ignore
		const result: any = await load({ fetch: mockFetch });

		expect(result).toEqual({ repos: [] });
	});

	it('returns empty array on network error', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));

		// @ts-ignore
		const result: any = await load({ fetch: mockFetch });

		expect(result).toEqual({ repos: [] });
	});
});
