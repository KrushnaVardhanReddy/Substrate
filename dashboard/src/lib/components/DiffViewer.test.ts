// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import DiffViewer from './DiffViewer.svelte';

describe('DiffViewer Component', () => {
	it('renders without crashing', () => {
		const { container } = render(DiffViewer, { before: 'a', after: 'a' });
		expect(container.querySelector('.diff-viewer')).toBeInTheDocument();
	});

	it('renders added and removed lines correctly', () => {
		const beforeStr = `name: A
version: 1.0.0`;
		const afterStr = `name: A
version: 1.1.0`;

		const { container } = render(DiffViewer, { before: beforeStr, after: afterStr });

		const removedLines = container.querySelectorAll('.diff-side.removed');
		const addedLines = container.querySelectorAll('.diff-side.added');
		const normalLines = container.querySelectorAll('.diff-side.normal');

		expect(removedLines.length).toBeGreaterThan(0);
		expect(addedLines.length).toBeGreaterThan(0);
		expect(normalLines.length).toBeGreaterThan(0);

		// Content validation
		expect(removedLines[0].textContent).toContain('version: 1.0.0');
		expect(addedLines[0].textContent).toContain('version: 1.1.0');
		expect(normalLines[0].textContent).toContain('name: A');
	});

	it('handles empty inputs gracefully', () => {
		const { container } = render(DiffViewer, { before: '', after: '' });
		const viewer = container.querySelector('.diff-viewer');
		expect(viewer).toBeInTheDocument();
	});
});
