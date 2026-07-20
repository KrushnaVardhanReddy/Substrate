import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/svelte';
import { load } from './+page';
import Page from './+page.svelte';

describe('Detail Page Load Function', () => {
	it('returns org, repo, and yaml string', async () => {
		// @ts-ignore
		const result: any = await load({
			params: { org: 'myorg', repo: 'myrepo' }
		});

		expect(result.org).toBe('myorg');
		expect(result.repo).toBe('myrepo');
		expect(result.yamlString).toContain('openapi: 3.0.0');
		expect(result.yamlString).toContain('title: myrepo API');
		expect(result.apiBaseUrl).toBe('https://api.substrate.com');
		expect(result.dashboardUrl).toBe('https://app.substrate.com');
	});
});

describe('Detail Page Component', () => {
	beforeEach(() => {
		// Elements web component is registered globally by script tag,
		// Since we just render the tag, we mock out custom element registration warnings in testing
		vi.spyOn(console, 'error').mockImplementation(() => {});
	});

	it('renders elements-api component with dummy data', () => {
		const { container, getByText } = render(Page, {
			data: { org: 'myorg', repo: 'myrepo', yamlString: 'dummy-yaml', repos: [], apiBaseUrl: 'https://api.substrate.com', dashboardUrl: 'https://app.substrate.com' }
		});

		// Check basic title presence
		expect(getByText('myrepo')).toBeInTheDocument();
		expect(getByText('API Documentation and Service Details')).toBeInTheDocument();

		// Check that the custom element is rendered with the right props
		const elementsApi = container.querySelector('elements-api');
		expect(elementsApi).toBeTruthy();
		expect(elementsApi?.getAttribute('apiDescriptionDocument')).toBe('dummy-yaml');
		expect(elementsApi?.getAttribute('router')).toBe('hash');
	});

	it('renders the badge snippet correctly', () => {
		const { getAllByTestId } = render(Page, {
			data: { org: 'myorg', repo: 'myrepo', yamlString: 'dummy-yaml', repos: [], apiBaseUrl: 'https://api.substrate.com', dashboardUrl: 'https://app.substrate.com' }
		});

		const badgeCodes = getAllByTestId('badge-snippet');
		expect(badgeCodes[0].textContent).toBe('[![Contract Score](https://api.substrate.com/api/badges/myorg/myrepo)](https://app.substrate.com/org/myorg/catalog/myrepo)');
	});
});
