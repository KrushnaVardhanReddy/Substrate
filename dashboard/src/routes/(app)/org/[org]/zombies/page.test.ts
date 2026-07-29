import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, cleanup } from '@testing-library/svelte';
import { tick } from 'svelte';
import ZombiesPage from './+page.svelte';

vi.mock('$app/stores', () => ({
	page: {
		subscribe: (fn: any) => {
			fn({ params: { org: 'test-org' } });
			return () => {};
		}
	}
}));

describe('ZombiesPage', () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

    afterEach(() => {
        cleanup();
    });

	it('renders FinOps banner with total monthly savings', async () => {
		global.fetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => ({
				zombies: [
					{ path: '/api/v1/old', method: 'GET', request_count: 0 }
				],
				total_monthly_savings: 500,
				savings_per_endpoint: 500
			})
		});

		render(ZombiesPage);

        await tick();
        await new Promise(r => setTimeout(r, 100)); // allow fetch to resolve
        await tick();

		await waitFor(() => {
			expect(screen.getByText('Total Potential Monthly Savings')).toBeInTheDocument();
			expect(screen.getByText('$500')).toBeInTheDocument();
		}, { timeout: 3000 });
	});

	it('displays potential savings per row', async () => {
		global.fetch = vi.fn().mockResolvedValue({
			ok: true,
			json: async () => ({
				zombies: [
					{ path: '/api/v1/old', method: 'GET', request_count: 0 }
				],
				total_monthly_savings: 500,
				savings_per_endpoint: 500
			})
		});

		render(ZombiesPage);

        await tick();
        await new Promise(r => setTimeout(r, 100));
        await tick();

		await waitFor(() => {
			expect(screen.getByText('Potential Savings: $500/mo')).toBeInTheDocument();
		}, { timeout: 3000 });
	});
});
