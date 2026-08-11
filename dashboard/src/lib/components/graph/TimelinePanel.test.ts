// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import TimelinePanel from './TimelinePanel.svelte';
import { tick } from 'svelte';

vi.mock('$app/stores', () => ({
	page: {
		subscribe: (fn: any) => {
			fn({ params: { org: 'test-org' } });
			return () => {};
		}
	}
}));

describe('TimelinePanel', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		global.fetch = vi.fn();
	});

	it('renders events correctly', async () => {
		const mockEvents = [
			{
				repo: 'frontend',
				event_type: 'deployment',
				description: 'v1.0.0 released',
				event_time: new Date().toISOString()
			},
			{
				repo: 'backend',
				event_type: 'incident',
				description: 'Database connection failed',
				event_time: new Date().toISOString()
			}
		];

		(global.fetch as any).mockResolvedValue({
			ok: true,
			json: async () => mockEvents
		});

		render(TimelinePanel, { props: { selectedEventRepo: null } });

		await waitFor(() => {
			expect(screen.getByText('frontend')).toBeInTheDocument();
			expect(screen.getByText('backend')).toBeInTheDocument();
			expect(screen.getByText('v1.0.0 released')).toBeInTheDocument();
			expect(screen.getByText('Database connection failed')).toBeInTheDocument();
		});
	});

	it('shows no events message when empty', async () => {
		(global.fetch as any).mockResolvedValue({
			ok: true,
			json: async () => []
		});

		render(TimelinePanel, { props: { selectedEventRepo: null } });

		await waitFor(() => {
			expect(screen.getByText('No events found in the selected timeframe.')).toBeInTheDocument();
		});
	});
});
