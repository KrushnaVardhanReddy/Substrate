// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi } from 'vitest';
import { load } from './+page';
import { error } from '@sveltejs/kit';

describe('Preview Route Load Function', () => {
	it('should return diff data successfully on 200', async () => {
		const mockReport = {
			base_schema: 'old schema',
			head_schema: 'new schema'
		};
		const mockFetch = vi.fn().mockResolvedValue({
			ok: true,
			status: 200,
			json: async () => mockReport
		});

		const result = await load({
			params: { token: 'mock-token' },
			fetch: mockFetch
		} as any);

		expect(result).toEqual({
			token: 'mock-token',
			oldText: 'old schema',
			newText: 'new schema',
			report: mockReport,
			error: null
		});
	});

	it('should return expired error on 410', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			status: 410
		});

		const result = await load({
			params: { token: 'mock-token' },
			fetch: mockFetch
		} as any);

		expect(result).toEqual({ error: 'This preview link has expired.' });
	});

	it('should return not found error on 404', async () => {
		const mockFetch = vi.fn().mockResolvedValue({
			ok: false,
			status: 404
		});

		const result = await load({
			params: { token: 'mock-token' },
			fetch: mockFetch
		} as any);

		expect(result).toEqual({ error: 'Preview not found or invalid token.' });
	});
});
