import { render } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeAll } from 'vitest';
import Page from './+page.svelte';

// Mock $app/stores
vi.mock('$app/stores', () => ({
	page: {
		subscribe: vi.fn((fn) => {
			fn({ params: { org: 'test-org' } });
			return () => {};
		})
	}
}));

// Mock Svelte Flow and html-to-image
vi.mock('@xyflow/svelte', () => {
    const DummyComponent = function(anchor: any, props: any) { return {}; };
	DummyComponent.render = () => ({ html: '' });
	return {
		SvelteFlow: DummyComponent,
		MiniMap: DummyComponent,
		Controls: DummyComponent,
		Background: DummyComponent,
		BackgroundVariant: { Dots: 'dots' },
		useSvelteFlow: vi.fn()
	};
});

describe('Graph Page UI', () => {
	it('renders Export PNG button', () => {
		const { getByText } = render(Page, { props: { data: { graphData: [] } } });
		expect(getByText('Export PNG')).toBeInTheDocument();
	});
});
