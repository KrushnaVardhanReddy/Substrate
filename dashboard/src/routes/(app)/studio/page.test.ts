// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, afterEach } from 'vitest';
import { render, cleanup, fireEvent } from '@testing-library/svelte';
import StudioPage from './+page.svelte';

describe('Visual API Design Studio', () => {
	afterEach(() => {
		cleanup();
	});

	it('renders initial state correctly with default YAML', () => {
		const { getByText, getByRole, getAllByText } = render(StudioPage);

		expect(getByText('Visual API Design Studio')).toBeInTheDocument();
		expect(getByText('OpenAPI YAML')).toBeInTheDocument();
		expect(getByText('Visual Designer')).toBeInTheDocument();

		// Default YAML has /users GET endpoint
		expect(getByText('/users')).toBeInTheDocument();
		const getBadges = getAllByText('GET');
		expect(getBadges.length).toBeGreaterThan(0);
	});

	it('updates visual representation when YAML is edited', async () => {
		const { getAllByRole, getByText, getAllByText, queryByText } = render(StudioPage);

		const textareas = getAllByRole('textbox');
		// The textarea we want is the actual textarea element, not the input types
		const textarea = textareas.find(t => t.tagName === 'TEXTAREA') || textareas[0];
		expect(textarea).toBeInTheDocument();

		const newYaml = `openapi: 3.0.0
info:
  title: Sample API
  version: 1.0.0
paths:
  /products:
    post:
      summary: Creates a product
`;

		await fireEvent.input(textarea, { target: { value: newYaml } });

		// Old endpoint should be gone, new one should be present
		expect(queryByText('/users')).not.toBeInTheDocument();
		expect(getByText('/products')).toBeInTheDocument();
		expect(getAllByText('POST').length).toBeGreaterThan(0);
		expect(getByText('Creates a product')).toBeInTheDocument();
	});

	it('shows error message on invalid YAML', async () => {
		const { getAllByRole, getByText } = render(StudioPage);

		const textareas = getAllByRole('textbox');
		const textarea = textareas.find(t => t.tagName === 'TEXTAREA') || textareas[0];

		// Intentionally malformed YAML (missing quotes, bad syntax)
		const badYaml = `openapi: 3.0.0
info:
  title: Sample: [unclosed bracket
`;

		await fireEvent.input(textarea, { target: { value: badYaml } });

		// Should show an error message. Error message content depends on js-yaml,
		// but it should render something in a danger color.
		await new Promise(resolve => setTimeout(resolve, 100)); // Wait for tick

		// The component displays `parseError` which comes from `e.message` of js-yaml.
		const errorText = document.body.textContent;
		expect(errorText).toContain('bad indentation of a mapping entry');
	});

	it('adds a new endpoint via form and updates YAML', async () => {
		const { getByRole, getAllByRole, getByPlaceholderText, getByText, getAllByText } = render(StudioPage);

		const methodSelect = getByRole('combobox');
		const pathInput = getByPlaceholderText('/path');
		const summaryInput = getByPlaceholderText('Summary');
		const addButton = getByRole('button', { name: 'Add' });

		await fireEvent.change(methodSelect, { target: { value: 'POST' } });
		await fireEvent.input(pathInput, { target: { value: '/new-test-route' } });
		await fireEvent.input(summaryInput, { target: { value: 'Test Summary' } });

		await fireEvent.click(addButton);

		// Visual list should have it
		expect(getByText('/new-test-route')).toBeInTheDocument();
		expect(getByText('Test Summary')).toBeInTheDocument();

		// Textarea should contain the new path
		const textareas = getAllByRole('textbox');
		const textarea = (textareas.find(t => t.tagName === 'TEXTAREA') || textareas[0]) as HTMLTextAreaElement;
		expect(textarea.value).toContain('/new-test-route');
		expect(textarea.value).toContain('Test Summary');
		expect(textarea.value).toContain('post:');
	});
});
