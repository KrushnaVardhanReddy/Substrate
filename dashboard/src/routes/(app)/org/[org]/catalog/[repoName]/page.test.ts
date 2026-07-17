import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import Page from './+page.svelte';

describe('Catalog Detail Page', () => {
	it('renders header and embed stoplight elements component', () => {
		const mockData = {
			org: 'demo-org',
			repoName: 'test-service',
			rawYaml: 'openapi: 3.0.0\ninfo:\n  title: test-service API'
		};

		const { getByText } = render(Page, {
			data: mockData
		});

		// Check the header is rendered correctly
		expect(getByText('test-service API')).toBeInTheDocument();
		expect(getByText('← Back to Catalog')).toBeInTheDocument();

		// Ensure the custom element is rendered in DOM
		const stoplightElement = document.querySelector('elements-api');
		expect(stoplightElement).not.toBeNull();
		expect(stoplightElement?.getAttribute('apiDescriptionDocument')).toBe(mockData.rawYaml);
		expect(stoplightElement?.getAttribute('router')).toBe('hash');
	});
});
