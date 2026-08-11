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

describe('Catalog Page Load Function', () => {
	beforeEach(() => {
		vi.resetAllMocks();
	});

	it('returns fetched nodes when /api/v1/nodes succeeds', async () => {
		const mockNodes = [{ id: '1', name: 'repo-1', owner: 'team-a', protocol: 'REST', status: 'SAFE', upstreamCount: 1, downstreamCount: 1 }];

		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => mockNodes
		});

		// @ts-ignore
		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		});

		// Note: isBrowser evaluates to true in Vitest jsdom environment by default,
		// so baseUrl will be empty string rather than PUBLIC_API_URL.
		expect(mockFetch).toHaveBeenCalledWith('/api/v1/nodes', {
			headers: {
				Authorization: `Bearer test-token`
			}
		});

		expect(result.org).toBe('myorg');
		expect(result.nodes).toEqual(mockNodes);
	});

	it('falls back to mock data when /api/v1/nodes fails', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			statusText: 'Not Found'
		});

		// @ts-ignore
		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		});

		expect(mockFetch).toHaveBeenCalled();
		expect(result.org).toBe('myorg');
		expect(result.nodes).toBeDefined();
		expect(result.nodes.length).toBeGreaterThan(0);
		expect(result.nodes[0].name).toBe('core-service');
	});

	it('falls back to mock data on network error', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));

		// @ts-ignore
		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		});

		expect(mockFetch).toHaveBeenCalled();
		expect(result.org).toBe('myorg');
		expect(result.nodes).toBeDefined();
		expect(result.nodes.length).toBeGreaterThan(0);
		expect(result.nodes[0].name).toBe('core-service');
	});
});
