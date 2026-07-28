import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Page from './+page.svelte';

// Mock the API URL
vi.mock('$env/dynamic/public', () => ({
	env: { PUBLIC_API_URL: 'http://localhost:8090' }
}));

describe('HITL Queue Page', () => {
	it('renders empty state when no items', () => {
		render(Page, { props: { data: { repos: [], org: 'test-org', items: [] } as any } });
		expect(screen.getByText('No pending items in the HITL queue.')).toBeDefined();
	});

	it('renders items and approve/reject buttons', () => {
		const mockItems = [
			{
				id: 1,
				tool_name: 'delete_governance_rule',
				arguments: { rule_id: 123 },
				status: 'pending'
			}
		];

		render(Page, { props: { data: { repos: [], org: 'test-org', items: mockItems } as any } });
		render(Page, { props: { data: { repos: [], org: 'test-org', items: mockItems } as any } });

		expect(screen.getByText('delete_governance_rule')).toBeDefined();
		expect(screen.getByText('Pending')).toBeDefined();
		expect(screen.getByText('Approve')).toBeDefined();
		expect(screen.getByText('Reject')).toBeDefined();
	});

	it('calls resolve API when clicking approve', async () => {
		const mockFetch = vi.fn().mockResolvedValue({ ok: true });
		global.fetch = mockFetch;

		const mockItems = [
			{
				id: 1,
				tool_name: 'delete_governance_rule',
				arguments: { rule_id: 123 },
				status: 'pending'
			}
		];

		render(Page, { props: { data: { repos: [], org: 'test-org', items: mockItems } as any } });

		const approveBtn = screen.getAllByText('Approve')[0];
		await fireEvent.click(approveBtn);

		expect(mockFetch).toHaveBeenCalledWith(
			'http://localhost:8090/api/v1/mcp/hitl-queue/1/resolve',
			expect.objectContaining({
				method: 'POST',
				body: JSON.stringify({ status: 'approved' })
			})
		);
	});
});
