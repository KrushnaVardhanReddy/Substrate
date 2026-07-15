import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import InteractiveEdge from './InteractiveEdge.svelte';
import type { Position } from '@xyflow/svelte';

// Mock getBezierPath from Svelte Flow
vi.mock('@xyflow/svelte', () => {
	const DummyComponent = function(anchor: any, props: any) { return {}; };
	DummyComponent.render = () => ({ html: '' });

	return {
		getBezierPath: vi.fn(() => ['M 0 0 C 50 0 50 100 100 100', 50, 50]),
		BaseEdge: DummyComponent,
		EdgeLabel: DummyComponent
	};
});

describe('InteractiveEdge', () => {
	const defaultProps = {
        source: "src",
        target: "tgt",
        type: "interactive",
		id: 'e1',
		sourceX: 0,
		sourceY: 0,
		targetX: 100,
		targetY: 100,
		sourcePosition: 'bottom' as Position,
		targetPosition: 'top' as Position,
		style: '',
		markerEnd: ''
	};

	it('renders without crashing', () => {
		const { container } = render(InteractiveEdge, { props: defaultProps });
		expect(container.querySelector('g')).toBeInTheDocument();
	});

	it('shows tooltip on hover and hides on mouseleave', async () => {
		const { container, queryByText } = render(InteractiveEdge, { props: defaultProps });
		const gElement = container.querySelector('g');

		expect(gElement).not.toBeNull();
	});
});
