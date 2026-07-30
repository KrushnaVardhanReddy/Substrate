import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, waitFor } from '@testing-library/svelte';
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
	let mockFetch: any;

	beforeEach(() => {
		vi.spyOn(console, 'error').mockImplementation(() => {});

		mockFetch = vi.fn().mockResolvedValue({
			json: async () => []
		});
		global.fetch = mockFetch;
	});

	it('renders elements-api component with dummy data by default (Schema tab)', () => {
		const { container, getByText } = render(Page, {
			props: { data: { repos: [], org: 'myorg', repo: 'myrepo', yamlString: 'dummy-yaml', apiBaseUrl: 'https://api.substrate.com', dashboardUrl: 'https://app.substrate.com', repoData: null } }
		});

		expect(getByText('myrepo')).toBeInTheDocument();
		expect(getByText('API Documentation and Service Details')).toBeInTheDocument();

		const elementsApi = container.querySelector('elements-api');
		expect(elementsApi).toBeTruthy();
		expect(elementsApi?.getAttribute('apiDescriptionDocument')).toBe('dummy-yaml');
		expect(elementsApi?.getAttribute('router')).toBe('hash');
	});

	it('renders the badge snippet correctly', () => {
		const { getAllByTestId } = render(Page, {
			props: { data: { repos: [], org: 'myorg', repo: 'myrepo', yamlString: 'dummy-yaml', apiBaseUrl: 'https://api.substrate.com', dashboardUrl: 'https://app.substrate.com', repoData: null } }
		});

		const badgeCodes = getAllByTestId('badge-snippet');
		expect(badgeCodes[0].textContent).toBe('[![Contract Score](https://api.substrate.com/api/badges/myorg/myrepo)](https://app.substrate.com/org/myorg/catalog/myrepo)');
	});

	it('switches to Guides tab and fetches guides', async () => {
		mockFetch.mockResolvedValueOnce({
			json: async () => [{ title: 'My Guide', file_path: 'my-guide.md', content: '# Hello' }]
		});

		const { getAllByText, getByText, getByRole, container } = render(Page, {
			props: { data: { repos: [], org: 'myorg', repo: 'myrepo', yamlString: 'dummy-yaml', apiBaseUrl: 'https://api.substrate.com', dashboardUrl: 'https://app.substrate.com', repoData: null } }
		});

		await new Promise(r => setTimeout(r, 0));

		const guidesTab = getAllByText('Guides')[0];
		expect(guidesTab).not.toBeNull();
		if (guidesTab) {
			await fireEvent.click(guidesTab);
		}

		await waitFor(() => {
			expect(mockFetch).toHaveBeenCalledWith('/api/v1/docs/myorg/myrepo');
			expect(getByText('My Guide')).toBeInTheDocument();
		});

		mockFetch.mockResolvedValueOnce({
			json: async () => ({ title: 'My Guide', file_path: 'my-guide.md', content: '# Hello Markdown' })
		});

		await fireEvent.click(getByText('My Guide'));

		await waitFor(() => {
			expect(mockFetch).toHaveBeenCalledWith('/api/v1/docs/myorg/myrepo/my-guide.md');
			expect(getByText('Hello Markdown')).toBeInTheDocument();
		});
	});

	it('renders business context correctly when metadata is present', async () => {
		const repoDataWithMetadata = {
			metadata: {
				owner: 'platform-team',
				slack_channel: '#platform-alerts',
				pagerduty: 'PD-1234',
				pm: 'Jane Doe',
				sla_tier: 'Tier 1'
			}
		};

		const { getByText, queryByText, container } = render(Page, {
			props: { data: { repos: [], org: 'myorg', repo: 'myrepo', yamlString: 'dummy-yaml', apiBaseUrl: 'https://api.substrate.com', dashboardUrl: 'https://app.substrate.com', repoData: repoDataWithMetadata } }
		});

		const businessTab = Array.from(container.querySelectorAll('button')).find(el => el.textContent?.includes('Business Context')) as HTMLElement;
		await fireEvent.click(businessTab);

		await new Promise(r => setTimeout(r, 0));

		expect(getByText('platform-team')).toBeInTheDocument();
		expect(getByText('#platform-alerts')).toBeInTheDocument();
		expect(getByText('PD-1234')).toBeInTheDocument();
		expect(getByText('Jane Doe')).toBeInTheDocument();
		expect(getByText('Tier 1')).toBeInTheDocument();
	});

	it('handles missing metadata without crashing', async () => {
		const { getByText, container } = render(Page, {
			props: { data: { repos: [], org: 'myorg', repo: 'myrepo', yamlString: 'dummy-yaml', apiBaseUrl: 'https://api.substrate.com', dashboardUrl: 'https://app.substrate.com', repoData: null } }
		});

		const businessTab = Array.from(container.querySelectorAll('button')).find(el => el.textContent?.includes('Business Context')) as HTMLElement;
		await fireEvent.click(businessTab);

		await new Promise(r => setTimeout(r, 0));

		expect(getByText('No business metadata found for this repository.')).toBeInTheDocument();
	});
});
