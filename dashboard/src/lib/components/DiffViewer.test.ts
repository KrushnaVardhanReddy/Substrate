import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import DiffViewer from './DiffViewer.svelte';

describe('DiffViewer', () => {
    it('renders normal lines when before and after are identical', () => {
        const text = 'line1\nline2';
        const { container } = render(DiffViewer, { before: text, after: text });

        // Should have 2 rows
        const rows = container.querySelectorAll('.diff-row');
        expect(rows.length).toBe(2);

        // First row left content
        const firstLeft = rows[0].querySelector('.left-content.normal');
        expect(firstLeft?.textContent?.trim()).toBe('line1');
    });

    it('renders added lines correctly', () => {
        const { container } = render(DiffViewer, { before: 'line1', after: 'line1\nline2' });

        const rows = container.querySelectorAll('.diff-row');
        expect(rows.length).toBe(2); // One normal, one added

        // Second row should be an addition
        const secondRight = rows[1].querySelector('.right-content.add');
        expect(secondRight?.textContent?.trim()).toBe('line2');
        const secondLeft = rows[1].querySelector('.left-content.empty');
        expect(secondLeft).not.toBeNull();
    });

    it('renders deleted lines correctly', () => {
        const { container } = render(DiffViewer, { before: 'line1\nline2', after: 'line1' });

        const rows = container.querySelectorAll('.diff-row');
        expect(rows.length).toBe(2);

        // Second row should be a deletion
        const secondLeft = rows[1].querySelector('.left-content.delete');
        expect(secondLeft?.textContent?.trim()).toBe('line2');
        const secondRight = rows[1].querySelector('.right-content.empty');
        expect(secondRight).not.toBeNull();
    });
});
