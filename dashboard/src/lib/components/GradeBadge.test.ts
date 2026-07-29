import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import GradeBadge from './GradeBadge.svelte';

describe('GradeBadge', () => {
	it('renders correctly with grade A', () => {
		const { container } = render(GradeBadge, { props: { grade: 'A', org: 'my-org', repo: 'my-repo' } });
		const badge = container.querySelector('.grade-badge');
		expect(badge).toBeTruthy();
		expect(badge?.textContent?.trim()).toBe('A');
		expect(badge?.className).toContain('bg-green-100');
		expect(badge?.getAttribute('href')).toBe('/org/my-org/scorecard/my-repo');
	});

	it('renders correctly with grade F', () => {
		const { container } = render(GradeBadge, { props: { grade: 'F', org: 'my-org', repo: 'my-repo' } });
		const badge = container.querySelector('.grade-badge');
		expect(badge?.textContent?.trim()).toBe('F');
		expect(badge?.className).toContain('bg-red-100');
	});
});
