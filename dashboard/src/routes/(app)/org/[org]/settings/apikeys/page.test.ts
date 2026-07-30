import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import ApiKeysPage from '../../apikeys/+page.svelte';

vi.mock('$app/stores', () => ({
	page: {
		subscribe: (fn: any) => {
			fn({ params: { org: 'test-org' } });
			return () => {};
		}
	}
}));

describe('API Keys Page', () => {
	beforeEach(() => {
		vi.resetAllMocks();
	});

	afterEach(() => {
		cleanup();
	});

	it('renders correctly with no keys', () => {
		render(ApiKeysPage, {
			props: {
				data: { repos: [], keys: [] }
			}
		});

		expect(screen.getByText('API Keys')).toBeInTheDocument();
		expect(screen.getByText('Manage API keys for test-org')).toBeInTheDocument();
		expect(screen.getByText('No API keys generated yet.')).toBeInTheDocument();

		const btn = screen.getByRole('button', { name: /Generate New Key/i });
		expect(btn).toBeInTheDocument();
	});

	it('renders keys table correctly', () => {
		const mockKeys = [
			{
				id: '1',
				name: 'Test Key 1',
				prefix: 'test1_',
				created_at: '2023-01-01T00:00:00Z'
			},
			{
				id: '2',
				name: 'Prod Key',
				prefix: 'prod_',
				created_at: '2023-01-02T00:00:00Z'
			}
		];

		render(ApiKeysPage, {
			props: {
				data: { repos: [], keys: mockKeys }
			}
		});

		expect(screen.getByText('Test Key 1')).toBeInTheDocument();
		expect(screen.getByText('test1_')).toBeInTheDocument();
		expect(screen.getByText('Prod Key')).toBeInTheDocument();
		expect(screen.getByText('prod_')).toBeInTheDocument();

		const revokeButtons = screen.getAllByRole('button', { name: /Revoke/i });
		expect(revokeButtons).toHaveLength(2);
	});

	it('opens modal when clicking Create Key (Generate New Key)', async () => {
		render(ApiKeysPage, {
			props: {
				data: { repos: [], keys: [] }
			}
		});

		const generateBtn = screen.getByRole('button', { name: /Generate New Key/i });
		await fireEvent.click(generateBtn);

		expect(screen.getByText('Generate New API Key')).toBeInTheDocument();
		expect(screen.getByLabelText('Key Name')).toBeInTheDocument();
		const buttons = screen.getAllByRole('button', { name: /Generate/i });
		expect(buttons.length).toBeGreaterThan(0);
	});
});
