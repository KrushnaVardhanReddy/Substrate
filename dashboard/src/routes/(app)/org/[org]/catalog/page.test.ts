import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/svelte';
import Page from './+page.svelte';

describe('Catalog List Page', () => {
	it('renders empty state when no data', () => {
		const { getByText } = render(Page, {
			data: {
				org: 'test-org',
				catalogData: []
			}
		});

		expect(getByText('No services found in this organization.')).toBeInTheDocument();
	});

	it('renders catalog data correctly', () => {
		const mockData = {
			org: 'demo-org',
			catalogData: [
				{
					name: 'core-service',
					protocol: 'gRPC',
					owner: 'platform',
					status: 'SAFE'
				},
				{
					name: 'mobile-ios',
					protocol: 'GraphQL',
					owner: 'mobile',
					status: 'BREAKING'
				}
			]
		};

		const { getByText } = render(Page, {
			data: mockData
		});

		expect(getByText('core-service')).toBeInTheDocument();
		expect(getByText('gRPC')).toBeInTheDocument();
		expect(getByText('platform')).toBeInTheDocument();

		expect(getByText('mobile-ios')).toBeInTheDocument();
		expect(getByText('GraphQL')).toBeInTheDocument();
		expect(getByText('mobile')).toBeInTheDocument();

		const breakingElements = document.querySelectorAll('.status-indicator.breaking');
		expect(breakingElements.length).toBeGreaterThan(0);
		expect(breakingElements[0]?.textContent?.trim()).toBe('BREAKING');
	});
});
