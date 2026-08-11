// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/svelte/svelte5';
import Page from './+page.svelte';

vi.mock('$env/dynamic/public', () => ({
	env: { PUBLIC_API_URL: 'http://localhost:8090' }
}));

describe('Marketplace Page', () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it('renders loading state initially', () => {
		global.fetch = vi.fn(() => new Promise(() => {})) as any;
		render(Page);
		expect(screen.getByText('Loading plugins...')).toBeInTheDocument();
	});

	it('renders error state on fetch failure', async () => {
		global.fetch = vi.fn().mockRejectedValue(new Error('Network error'));
		render(Page);

		await waitFor(() => {
			expect(screen.getByText('Network error')).toBeInTheDocument();
		});
	});

	it('renders list of plugins', async () => {
		const mockPlugins = [
			{ name: 'plugin-a', description: 'desc a' },
			{ name: 'plugin-b', description: 'desc b' }
		];

		global.fetch = vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(mockPlugins)
		}) as any;

		render(Page);

		await waitFor(() => {
			expect(screen.getByText('plugin-a')).toBeInTheDocument();
			expect(screen.getByText('desc a')).toBeInTheDocument();
			expect(screen.getByText('plugin-b')).toBeInTheDocument();
		});
	});

	it('copies install command to clipboard', async () => {
		const mockPlugins = [{ name: 'plugin-a', description: 'desc a' }, { name: 'plugin-b', description: 'desc b' }];
		global.fetch = vi.fn().mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(mockPlugins)
		}) as any;

		const writeTextMock = vi.fn().mockResolvedValue(undefined);
		Object.assign(navigator, {
			clipboard: {
				writeText: writeTextMock
			}
		});

		render(Page);

		await waitFor(() => {
			expect(screen.getByText('plugin-a')).toBeInTheDocument();
		});

		const copyButtons = screen.getAllByText('Copy');
		await fireEvent.click(copyButtons[0]);

		expect(writeTextMock).toHaveBeenCalledWith('substrate plugin install plugin-a');
		expect(screen.getByText('Copied!')).toBeInTheDocument();
	});
});
