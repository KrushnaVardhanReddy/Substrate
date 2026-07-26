import { describe, it, expect, vi } from 'vitest';
import { load } from './+page';

describe('Matrix Page Load Function', () => {
	it('returns fetched matrix data on success', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => ({ providers: ['P'], consumers: ['C'], grid: [] })
		});

		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		} as any);

		expect(result.org).toBe('myorg');
		expect(result.matrixData.providers).toHaveLength(1);
	});

	it('returns fallback mock data on network error', async () => {
		const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));

		const result: any = await load({
			fetch: mockFetch,
			params: { org: 'myorg' }
		} as any);

		expect(result.org).toBe('myorg');
		expect(result.matrixData.providers.length).toBe(0);
	});
});
