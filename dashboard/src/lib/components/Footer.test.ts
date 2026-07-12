import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, fireEvent, cleanup } from '@testing-library/svelte';
import Footer from './Footer.svelte';

describe('Footer', () => {
    afterEach(() => {
        cleanup();
    });

	it('renders the footer link', () => {
		const { getByText } = render(Footer);
		expect(getByText('Credits / OSS Licenses')).toBeInTheDocument();
	});

	it('opens the modal when clicked', async () => {
		const { getByText, queryByText } = render(Footer);

		// Modal content should not be present initially
		expect(queryByText('Substrate is made possible by the following open-source software:')).not.toBeInTheDocument();

		// Click the link
		await fireEvent.click(getByText('Credits / OSS Licenses'));

		// Modal content should now be present
		expect(getByText('Substrate is made possible by the following open-source software:')).toBeInTheDocument();
		expect(getByText('Close')).toBeInTheDocument();
	});

	it('closes the modal when Close is clicked', async () => {
		const { getByText, queryByText } = render(Footer);

		// Click the link to open
		await fireEvent.click(getByText('Credits / OSS Licenses'));
		expect(getByText('Close')).toBeInTheDocument();

		// Click Close
		await fireEvent.click(getByText('Close'));

		// Modal should be gone
		expect(queryByText('Close')).not.toBeInTheDocument();
	});
});
