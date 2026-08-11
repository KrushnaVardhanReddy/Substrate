// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte';
import TopNav from './TopNav.svelte';
import * as navigation from '$app/navigation';

vi.mock('$app/navigation', () => ({
	goto: vi.fn()
}));

vi.mock('$env/dynamic/public', () => ({
	env: { PUBLIC_ORG_NAME: 'Test Org' }
}));

describe('TopNav Component', () => {
	it('renders user profile and sign out button', () => {
		const { getByText, getAllByLabelText } = render(TopNav);
		expect(getByText('Test Org')).toBeInTheDocument();
		expect(getByText('T')).toBeInTheDocument(); // Avatar initial
		expect(getAllByLabelText('Sign Out')[0]).toBeInTheDocument();
	});

	it('calls sign out logic when sign out button is clicked', async () => {
		const { getAllByLabelText } = render(TopNav);
		const signOutBtn = getAllByLabelText('Sign Out')[0];

		const removeItemSpy = vi.spyOn(Storage.prototype, 'removeItem');

		await fireEvent.click(signOutBtn);

		expect(removeItemSpy).toHaveBeenCalledWith('github_token');
		expect(navigation.goto).toHaveBeenCalledWith('/onboarding');
	});
});
