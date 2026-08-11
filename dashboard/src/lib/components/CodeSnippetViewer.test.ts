// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { render, fireEvent, waitFor, cleanup } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import CodeSnippetViewer from './CodeSnippetViewer.svelte';

describe('CodeSnippetViewer Component', () => {
	beforeEach(() => {
		// Mock the clipboard API
		Object.assign(navigator, {
			clipboard: {
				writeText: vi.fn().mockImplementation(() => Promise.resolve())
			}
		});
	});

	afterEach(() => {
		vi.clearAllMocks();
		cleanup();
	});

	it('renders cURL snippet by default', () => {
		const { getByText, container } = render(CodeSnippetViewer, {
			props: {
				method: 'GET',
				path: 'https://api.example.com/items'
			}
		});

		expect(getByText('cURL')).toBeInTheDocument();
		expect(container.textContent).toContain('curl -X GET "https://api.example.com/items"');
	});

	it('switches tabs and displays the correct snippet', async () => {
		const { getByText, container } = render(CodeSnippetViewer, {
			props: {
				method: 'POST',
				path: 'https://api.example.com/items'
			}
		});

		const pythonTab = getByText('Python');
		await fireEvent.click(pythonTab);

		expect(container.textContent).toContain('requests.post(url)');

		const nodeTab = getByText('Node.js');
		await fireEvent.click(nodeTab);

		expect(container.textContent).toContain("fetch('https://api.example.com/items'");

		const goTab = getByText('Go');
		await fireEvent.click(goTab);

		expect(container.textContent).toContain('http.NewRequest("POST", url, nil)');
	});

	it('copies code to clipboard', async () => {
		const { getByText, getByRole } = render(CodeSnippetViewer, {
			props: {
				method: 'GET',
				path: 'https://api.example.com/items'
			}
		});

		const copyButton = getByRole('button', { name: 'Copy code' });
		await fireEvent.click(copyButton);

		expect(navigator.clipboard.writeText).toHaveBeenCalledWith('curl -X GET "https://api.example.com/items"');
		expect(getByText('Copied!')).toBeInTheDocument();
	});
});
