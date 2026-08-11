// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi } from 'vitest';
import { load } from './+page';

describe('Graph Page Load Function', () => {
	it('returns fetched graph data on success', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => [{ provider: 'A', consumer: 'B', status: 'SAFE' }]
		});

		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		} as any);

		expect(result.org).toBe('myorg');
		expect(result.graphData).toHaveLength(1);
		expect(result.graphData[0].provider).toBe('A');
	});

	it('returns fallback mock data on network error', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));

		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		} as any);

		expect(result.org).toBe('myorg');
		expect(result.graphData.length).toBe(0);

	});
});
