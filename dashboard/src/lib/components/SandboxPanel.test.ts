// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import SandboxPanel from './SandboxPanel.svelte';

global.fetch = vi.fn();

describe('SandboxPanel Component', () => {
	it('should expand and fetch token on click', async () => {
		(global.fetch as any).mockResolvedValueOnce({
			ok: true,
			json: async () => ({ token: 'mocked-ephemeral-token' })
		});

		render(SandboxPanel, { props: { org: 'test-org', repo: 'test-repo', apiBaseUrl: 'http://localhost:8090' } });

		const toggleBtn = screen.getByText('▶ Try It Out (Interactive API Sandbox)');
		await fireEvent.click(toggleBtn);

		expect(global.fetch).toHaveBeenCalledWith('http://localhost:8090/api/v1/sandbox/token', expect.any(Object));

		expect(screen.getByText('Request Builder')).toBeTruthy();
	});
});
