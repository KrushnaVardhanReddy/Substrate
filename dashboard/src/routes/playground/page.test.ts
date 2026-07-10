import { describe, it, expect, afterEach, vi } from 'vitest';
import { render, cleanup, fireEvent, act } from '@testing-library/svelte';
import PlaygroundPage from './+page.svelte';

describe('AI Playground Page', () => {
	afterEach(() => {
		cleanup();
		vi.useRealTimers();
	});

	it('renders initial state correctly', () => {
		const { getByText, getByRole, getAllByRole } = render(PlaygroundPage);

		expect(getByText('AI Schema Validator Playground')).toBeInTheDocument();
		expect(getByRole('button', { name: /Analyze with Substrate AI/i })).toBeInTheDocument();

		const textareas = getAllByRole('textbox');
		expect(textareas).toHaveLength(2);
		expect(textareas[0]).toHaveValue(`type User {\n  id: ID!\n  user_id: String!\n  name: String\n  email: String\n}`);
		expect(textareas[1]).toHaveValue(`type User {\n  id: ID!\n  name: String\n  email: String\n}`);
	});

	it('shows loading state when clicking analyze', async () => {
		const { getByRole, getByText } = render(PlaygroundPage);
		const analyzeBtn = getByRole('button', { name: /Analyze with Substrate AI/i });

		await fireEvent.click(analyzeBtn);

		expect(analyzeBtn).toBeDisabled();
		expect(getByText(/AI is analyzing cross-repo impact.../i)).toBeInTheDocument();
	});

	it('shows analysis result after delay and applies fix', async () => {
		vi.useFakeTimers();

		const { getByRole, getByText, getAllByRole } = render(PlaygroundPage);
		const analyzeBtn = getByRole('button', { name: /Analyze with Substrate AI/i });

		await fireEvent.click(analyzeBtn);

		// Advance time to trigger setTimeout callback
		await act(() => {
			vi.advanceTimersByTime(1500);
		});

		// Check analysis panel is shown
		expect(getByText('🚨 Breaking Change Detected')).toBeInTheDocument();
		expect(getByText(/Contextual Impact/)).toBeInTheDocument();

		// Check apply fix functionality
		const applyFixBtn = getByRole('button', { name: 'Apply Fix' });
		expect(applyFixBtn).toBeInTheDocument();

		await fireEvent.click(applyFixBtn);

		// Verify proposed schema textarea was updated
		const textareas = getAllByRole('textbox');
		expect(textareas[1]).toHaveValue(`type User {\n  id: ID!\n  user_id: String! @deprecated(reason: "Use id instead")\n  name: String\n  email: String\n}`);
	});
});
