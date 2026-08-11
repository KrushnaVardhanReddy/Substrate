// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { render, fireEvent, screen } from '@testing-library/svelte';
import { describe, it, expect, afterEach } from 'vitest';
import TimeTravelScrubber from './TimeTravelScrubber.svelte';

describe('TimeTravelScrubber', () => {
	afterEach(() => {
		document.body.innerHTML = '';
	});

	it('renders the component with default state', () => {
		render(TimeTravelScrubber);

		// Should show default date based on 75%
		const expectedDate = "Oct 24, 2026";
		expect(screen.getByText(expectedDate)).toBeInTheDocument();

		// Should show historical view text
		expect(screen.getByText('Historical View')).toBeInTheDocument();

		// Button should be in play state (aria-label="Play")
		const btn = screen.getByRole('button', { name: 'Play' });
		expect(btn).toBeInTheDocument();
	});

	it('toggles play/pause state when button is clicked', async () => {
		render(TimeTravelScrubber);

		const btn = screen.getByRole('button', { name: 'Play' });

		// Click to pause
		await fireEvent.click(btn);
		expect(screen.getByRole('button', { name: 'Pause' })).toBeInTheDocument();

		// Click to play again
		await fireEvent.click(screen.getByRole('button', { name: 'Pause' }));
		expect(screen.getByRole('button', { name: 'Play' })).toBeInTheDocument();
	});

	it('updates date when slider value changes', async () => {
		render(TimeTravelScrubber);

		const slider = screen.getByRole('slider') as HTMLInputElement;

		// Change to beginning (should be first date)
		await fireEvent.input(slider, { target: { value: '1' } });
		expect(screen.getByText('Jan 15, 2026')).toBeInTheDocument();

		// Change to middle
		await fireEvent.input(slider, { target: { value: '50' } });
		expect(screen.getByText('Jul 30, 2026')).toBeInTheDocument();

		// Change to end
		await fireEvent.input(slider, { target: { value: '100' } });
		expect(screen.getByText('Dec 01, 2026')).toBeInTheDocument();
	});
});
