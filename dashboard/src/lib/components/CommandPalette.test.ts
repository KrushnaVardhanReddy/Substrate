// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import CommandPalette from './CommandPalette.svelte';
import { page } from '$app/stores';

vi.mock('$app/stores', () => ({
	page: {
		subscribe: vi.fn((cb) => {
			cb({ params: { org: 'test-org' } });
			return () => {};
		})
	}
}));

describe('CommandPalette Component', () => {
	const mockRepos = [
		{ full_name: 'test-org/repo-1', name: 'repo-1' },
		{ full_name: 'test-org/repo-2', name: 'repo-2' }
	];

	beforeEach(() => {
		// Mock window.location
		const mockLocation = new URL('http://localhost');
		Object.defineProperty(window, 'location', {
			value: mockLocation,
			writable: true
		});
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('renders correctly but hidden initially', () => {
		const { container } = render(CommandPalette, { props: { repos: mockRepos } });
		expect(container.querySelector('.palette-overlay')).toBeNull();
	});

	it('opens on Cmd+K', async () => {
		const { container } = render(CommandPalette, { props: { repos: mockRepos } });

		await fireEvent.keyDown(window, { key: 'k', metaKey: true });
		expect(container.querySelector('.palette-overlay')).not.toBeNull();
	});

	it('filters nodes based on search query', async () => {
		const { container } = render(CommandPalette, { props: { repos: mockRepos } });
		await fireEvent.keyDown(window, { key: 'k', metaKey: true });

		const input = screen.getByPlaceholderText('Search for a microservice, team, or settings...');
		await fireEvent.input(input, { target: { value: 'repo-1' } });

		const items = container.querySelectorAll('.result-item');
		expect(items.length).toBe(1);
		expect(items[0].textContent).toContain('repo-1');
	});

	it('navigates with arrow keys', async () => {
		const { container } = render(CommandPalette, { props: { repos: mockRepos } });
		await fireEvent.keyDown(window, { key: 'k', metaKey: true });

		await fireEvent.keyDown(window, { key: 'ArrowDown' });

		const items = container.querySelectorAll('.result-item');
		expect(items[1].classList.contains('selected')).toBe(true);
	});
});
