// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { vi } from 'vitest';
import { render } from '@testing-library/svelte';
import ServiceNode from './ServiceNode.svelte';
import { describe, it, expect } from 'vitest';

vi.mock('@xyflow/svelte', () => ({
	Handle: function() { return { render: () => '' } },
	Position: { Top: 'top', Bottom: 'bottom', Left: 'left', Right: 'right' },
	useSvelteFlow: () => ({ zoomIn: vi.fn(), zoomOut: vi.fn() })
}));

vi.mock('lucide-svelte/icons/app-window', () => ({ default: function() { return { render: () => '' } } }));
vi.mock('lucide-svelte/icons/database', () => ({ default: function() { return { render: () => '' } } }));
vi.mock('lucide-svelte/icons/smartphone', () => ({ default: function() { return { render: () => '' } } }));
vi.mock('lucide-svelte/icons/server', () => ({ default: function() { return { render: () => '' } } }));

describe('ServiceNode Component', () => {
	it('uses node_color from metadata for background color if present', () => {
		const { container } = render(ServiceNode, {
			props: {
				data: {
					metadata: {
						node_color: '#ff00ff'
					}
				}
			}
		});

		const card = container.querySelector('.service-node-card') as HTMLElement;
		expect(card).not.toBeNull();
		expect(card.style.backgroundColor).toBe('rgb(255, 0, 255)'); // #ff00ff
	});

	it('defaults to standard color when node_color is absent', () => {
		const { container } = render(ServiceNode, {
			props: {
				data: {
					metadata: {}
				}
			}
		});

		const card = container.querySelector('.service-node-card') as HTMLElement;
		expect(card).not.toBeNull();
		// Default is #1E222C which is rgb(30, 34, 44)
		expect(card.style.backgroundColor).toBe('rgb(30, 34, 44)');
	});
});
