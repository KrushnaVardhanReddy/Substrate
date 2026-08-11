// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

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

	it('returns fetched repositories and returns them', async () => {
		const mockRepos = [{ id: '1', name: 'repo-1', full_name: 'org/repo-1' }];

		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => mockRepos
		});

		// @ts-ignore
		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		});

		expect(mockFetch).toHaveBeenCalledWith('http://test-api/api/v1/repos/myorg', {
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
		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		});

		expect(result).toEqual({ repos: [] });
	});

	it('returns empty array on network error', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));

		// @ts-ignore
		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		}) as any;

		expect(result).toEqual({ repos: [] });
	});
});
