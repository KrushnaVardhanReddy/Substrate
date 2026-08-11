// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, afterEach } from 'vitest';
import { render, cleanup } from '@testing-library/svelte';
import Sidebar from './Sidebar.svelte';

describe('Sidebar Component', () => {
	afterEach(() => {
		cleanup();
	});

	it('renders org links properly', () => {
		const repos = [
			{ id: '1', name: 'repo-1', full_name: 'org/repo-1' },
			{ id: '2', name: 'repo-2', full_name: 'org/repo-2' }
		];

		const { getByText } = render(Sidebar, { props: { repos, org: 'myorg', pathname: '/org/myorg' } });

		expect(getByText('Repositories')).toBeInTheDocument();
	});

	it('renders empty list if no repos', () => {
		const { getByText, queryByText } = render(Sidebar, { props: { repos: [], org: 'myorg', pathname: '/org/myorg' } });

		expect(getByText('Repositories')).toBeInTheDocument();
		expect(queryByText('org/repo-1')).not.toBeInTheDocument();
	});

	it('renders AI Playground link', () => {
		const { getByText } = render(Sidebar, { props: { repos: [], org: 'myorg', pathname: '/org/myorg' } });

		expect(getByText('✨ AI Playground')).toBeInTheDocument();
	});
});
