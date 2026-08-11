// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte';
import CatalogPage from './+page.svelte';

const mockData = {
	org: 'myorg',
	repos: [],
	nodes: [
		{
			id: "demo-org/core-service",
			name: "core-service",
			owner: "platform",
			protocol: "GraphQL",
			status: "SAFE",
			upstreamCount: 1,
			downstreamCount: 2
		},
		{
			id: "demo-org/gateway-service",
			name: "gateway-service",
			owner: "infrastructure",
			protocol: "OpenAPI",
			status: "SAFE",
			upstreamCount: 1,
			downstreamCount: 2
		}
	]
};

vi.mock('$app/navigation', () => {
	return {
		goto: vi.fn()
	};
});

describe('Catalog Page Component', () => {
	beforeEach(() => {
		vi.resetAllMocks();
	});

	it('renders catalog list correctly', () => {
		const { getByText } = render(CatalogPage, { data: mockData });

		expect(getByText('Service Catalog')).toBeInTheDocument();
		expect(getByText('core-service')).toBeInTheDocument();
		expect(getByText('gateway-service')).toBeInTheDocument();
		expect(getByText('platform')).toBeInTheDocument();
		expect(getByText('infrastructure')).toBeInTheDocument();
		expect(getByText('GraphQL')).toBeInTheDocument();
		expect(getByText('OpenAPI')).toBeInTheDocument();
	});

	it('filters the nodes correctly when searching', async () => {
		const { getAllByPlaceholderText, queryByText, getAllByText } = render(CatalogPage, { data: mockData });

		const searchInputs = getAllByPlaceholderText('Search services...');
		const searchInput = searchInputs[0];

		await fireEvent.input(searchInput, { target: { value: 'core' } });

		expect(getAllByText('core-service', { selector: 'td' }).length).toBeGreaterThan(0);
		// Note queryByText fails on multiple matches. If it's not rendered, queryAllByText returns []

		await fireEvent.input(searchInput, { target: { value: 'infrastructure' } });
		expect(getAllByText('gateway-service', { selector: 'td' }).length).toBeGreaterThan(0);
	});

	it('navigates to detail page when a row is clicked', async () => {
		const { getAllByText } = render(CatalogPage, { data: mockData });
		const { goto } = await import('$app/navigation');

		const rows = getAllByText('core-service', { selector: 'td' });
		const row = rows[0].closest('tr');
		await fireEvent.click(row!);

		expect(goto).toHaveBeenCalledWith('/org/myorg/catalog/core-service');
	});
});
