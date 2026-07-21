import { render, screen, fireEvent, waitFor } from '@testing-library/svelte/svelte5';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartnersPage from './+page.svelte';

// Mock the API URL and page store
vi.mock('$env/dynamic/public', () => ({
	env: { PUBLIC_API_URL: 'http://localhost:8090' }
}));

vi.mock('$app/stores', () => ({
	page: {
		subscribe: (fn: any) => {
			fn({ params: { org: 'test-org' } });
			return () => {};
		}
	}
}));

describe('Partners Admin Page', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		global.fetch = vi.fn();
	});

	it('renders empty state correctly', async () => {
		(global.fetch as any).mockResolvedValueOnce({
			ok: true,
			json: async () => []
		});

		render(PartnersPage);

		expect(await screen.findByText('No partners registered yet.')).toBeInTheDocument();
	});

	it('renders partners list', async () => {
		(global.fetch as any).mockResolvedValueOnce({
			ok: true,
			json: async () => [
				{
					id: '1',
					vendor_name: 'Test Vendor',
					status: 'pending',
					webhook_url: 'http://example.com/webhook'
				},
				{
					id: '2',
					vendor_name: 'Certified Vendor',
					status: 'certified',
					webhook_url: 'http://example.com/webhook2'
				}
			]
		});

		render(PartnersPage);

		expect(await screen.findByText('Test Vendor')).toBeInTheDocument();
		expect(await screen.findByText('Pending')).toBeInTheDocument();

		expect(await screen.findByText('Certified Vendor')).toBeInTheDocument();
		expect(await screen.findByText('Certified')).toBeInTheDocument();
	});

	it('triggers handshake successfully', async () => {
		// First call: initial load
		(global.fetch as any).mockResolvedValueOnce({
			ok: true,
			json: async () => [
				{
					id: '1',
					vendor_name: 'Test Vendor',
					status: 'pending',
					webhook_url: 'http://example.com/webhook'
				}
			]
		});

		render(PartnersPage);

		const triggerBtn = await screen.findByText('Trigger Handshake');
		expect(triggerBtn).toBeInTheDocument();

		// Second call: verify endpoint
		(global.fetch as any).mockResolvedValueOnce({
			ok: true,
			json: async () => ({})
		});

		// Third call: reload list after verification
		(global.fetch as any).mockResolvedValueOnce({
			ok: true,
			json: async () => [
				{
					id: '1',
					vendor_name: 'Test Vendor',
					status: 'certified',
					webhook_url: 'http://example.com/webhook'
				}
			]
		});

		await fireEvent.click(triggerBtn);

		// Assert it called the correct endpoint
		expect(global.fetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/v1/org/substrate/partners/1/verify'),
			expect.objectContaining({ method: 'POST' })
		);

		// State should update to certified
		expect(await screen.findByText('Certified')).toBeInTheDocument();
	});
});