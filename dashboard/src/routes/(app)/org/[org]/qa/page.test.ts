// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, cleanup, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte';
import QaPage from './[...repo]/+page.svelte';

vi.mock('$app/stores', () => ({
	page: {
		subscribe: (fn: any) => {
			fn({ params: { org: 'test-org', repo: 'test-repo' } });
			return () => {};
		}
	}
}));

describe('QA Page', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		global.fetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => ([
				{
					Severity: 'High',
					Method: 'GET',
					Path: '/api/v1/test',
					Issue: 'Missing parameter',
					Payload: { param: 'test' }
				}
			])
		});
	});

	afterEach(() => {
		cleanup();
	});

	it('renders correctly with mock data', async () => {
		render(QaPage);

		expect(screen.getByText('QA Dashboard: test-repo')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Export Postman Collection/i })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Export OpenAPI Spec/i })).toBeInTheDocument();

		await waitFor(async () => {
			expect(screen.getByText('High')).toBeInTheDocument();
			expect(screen.getByText((content) => content.includes('GET /api/v1/test'))).toBeInTheDocument();
			expect(screen.getByText((content) => content.includes('Missing parameter'))).toBeInTheDocument();
		});
	});
});
