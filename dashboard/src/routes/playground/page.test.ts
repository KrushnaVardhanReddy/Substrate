import { describe, it, expect, afterEach, beforeEach, vi, type Mock } from 'vitest';
import { render, cleanup, fireEvent, act } from '@testing-library/svelte';
import PlaygroundPage from './+page.svelte';

describe('AI Playground Page', () => {
	beforeEach(() => {
		// Mock global fetch for AI endpoints
		global.fetch = vi.fn().mockImplementation((url: string | URL | Request) => {
			const urlStr = url.toString();
			if (urlStr.includes('/api/v1/ai/analyze')) {
				const stream = new ReadableStream({
					start(controller) {
						// Send finding event
						const findingEvent = `event: finding\ndata: {"type": "finding", "severity": "🚨 Breaking Change Detected", "content": "Contextual Impact: Breaking Change Detected."}\n\n`;
						controller.enqueue(new TextEncoder().encode(findingEvent));

						// Send fix event
						const fixEvent = `event: fix\ndata: {"type": "fix", "language": "graphql", "code": "type User {\\n  id: ID!\\n  user_id: String! @deprecated(reason: \\"Use id instead\\")\\n  name: String\\n  email: String\\n}"}\n\n`;
						controller.enqueue(new TextEncoder().encode(fixEvent));

						// Send done event
						const doneEvent = `event: done\ndata: {"type": "done"}\n\n`;
						controller.enqueue(new TextEncoder().encode(doneEvent));

						controller.close();
					}
				});
				return Promise.resolve({
					ok: true,
					body: stream
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
		});
	});

	afterEach(() => {
		cleanup();
		vi.useRealTimers();
		vi.restoreAllMocks();
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
		expect(getByText('🚨 Analysis Results')).toBeInTheDocument();
		expect(getByText(/Contextual Impact/)).toBeInTheDocument();

		// Check apply fix functionality
		const applyFixBtn = getByRole('button', { name: 'Apply Fix' });
		expect(applyFixBtn).toBeInTheDocument();

		await fireEvent.click(applyFixBtn);

		// Allow tick/act to process the update
		await act(() => {
			vi.advanceTimersByTime(100);
		});

		// Verify proposed schema textarea was updated
		const textareas = getAllByRole('textbox');
		expect(textareas[1]).toHaveValue(`type User {\n  id: ID!\n  user_id: String! @deprecated(reason: "Use id instead")\n  name: String\n  email: String\n}`);
	});
});
