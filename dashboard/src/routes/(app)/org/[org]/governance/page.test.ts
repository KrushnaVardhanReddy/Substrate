import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/svelte';
import Page from './+page.svelte';

vi.mock('$app/stores', () => ({
	page: {
		subscribe: vi.fn((fn) => {
			fn({ params: { org: 'test-org' } });
			return () => {};
		})
	}
}));

vi.mock('$env/dynamic/public', () => ({
	env: {
		PUBLIC_API_URL: 'http://test-api',
	}
}));

describe('Governance Rules Page', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		global.fetch = vi.fn();
	});

	afterEach(() => {
		cleanup();
	});

	it('renders correctly', () => {
		render(Page);
		expect(screen.getByText('Governance Rules')).toBeInTheDocument();
		expect(screen.getByPlaceholderText('e.g. All payment APIs must require authentication')).toBeInTheDocument();
		expect(screen.getByText('Generate Rule')).toBeInTheDocument();
	});

	it('shows error if prompt is empty', async () => {
		render(Page);
		const button = screen.getByText('Generate Rule');
		await fireEvent.click(button);

		expect(screen.getByText('Please enter a rule description.')).toBeInTheDocument();
		expect(global.fetch).not.toHaveBeenCalled();
	});

	it('calls API and displays generated CEL', async () => {
		const mockResponse = { cel: 'request.path.matches("^/api/")' };
		(global.fetch as any).mockResolvedValueOnce({
			ok: true,
			json: async () => mockResponse,
		});

		render(Page);

		const textarea = screen.getByPlaceholderText('e.g. All payment APIs must require authentication');
		await fireEvent.input(textarea, { target: { value: 'Require auth' } });

		const button = screen.getByText('Generate Rule');
		await fireEvent.click(button);

		expect(global.fetch).toHaveBeenCalledWith('http://test-api/api/governance/generate-cel', expect.objectContaining({
			method: 'POST',
			body: JSON.stringify({ prompt: 'Require auth' })
		}));

		await waitFor(() => {
			expect(screen.getByText('request.path.matches("^/api/")')).toBeInTheDocument();
		});
	});

	it('shows error message if API fails', async () => {
		(global.fetch as any).mockResolvedValueOnce({
			ok: false,
			json: async () => ({ error: 'invalid rule' }),
		});

		render(Page);

		const textarea = screen.getByPlaceholderText('e.g. All payment APIs must require authentication');
		await fireEvent.input(textarea, { target: { value: 'Bad rule' } });

		const button = screen.getByText('Generate Rule');
		await fireEvent.click(button);

		await waitFor(() => {
			expect(screen.getByText('invalid rule')).toBeInTheDocument();
		});
	});
});
