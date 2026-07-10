import { describe, it, expect, afterEach } from 'vitest';
import { render, cleanup } from '@testing-library/svelte';
import Sidebar from './Sidebar.svelte';

describe('Sidebar Component', () => {
	afterEach(() => {
		cleanup();
	});

	it('renders repository list dynamically', () => {
		const repos = [
			{ id: '1', name: 'repo-1', full_name: 'org/repo-1' },
			{ id: '2', name: 'repo-2', full_name: 'org/repo-2' }
		];

		const { getByText } = render(Sidebar, { props: { repos, org: 'myorg', pathname: '/org/myorg' } });

		expect(getByText('org/repo-1')).toBeInTheDocument();
		expect(getByText('org/repo-2')).toBeInTheDocument();
	});

	it('renders empty list if no repos', () => {
		const { getByText, queryByText } = render(Sidebar, { props: { repos: [], org: 'myorg', pathname: '/org/myorg' } });

		expect(getByText('Repositories')).toBeInTheDocument();
		expect(queryByText('org/repo-1')).not.toBeInTheDocument();
	});
});
