// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import TeamGroupNode from './TeamGroupNode.svelte';

describe('TeamGroupNode Component', () => {
	it('renders properly with given props', () => {
		const { container, getByText } = render(TeamGroupNode, {
			props: {
				data: { label: 'Platform Team' },
				width: 300,
				height: 200,
				id: 'team-platform-team',
				type: 'teamGroup',
				selected: false,
				zIndex: 1,
				isConnectable: false,
				positionAbsoluteX: 0,
				positionAbsoluteY: 0,
				dragging: false,
				deletable: false,
				selectable: false,
				draggable: false
			}
		});

		const label = getByText('Platform Team');
		expect(label).toBeTruthy();

		// Check width and height are properly set
		const nodeElement = container.querySelector('.team-group-node') as HTMLElement;
		expect(nodeElement).not.toBeNull();
		expect(nodeElement.style.width).toBe('300px');
		expect(nodeElement.style.height).toBe('200px');
	});
});
