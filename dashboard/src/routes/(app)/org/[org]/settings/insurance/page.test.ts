// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import InsurancePage from './+page.svelte';

vi.mock('$app/stores', () => ({
  page: {
    subscribe: (fn: any) => {
      fn({ params: { org: 'test-org-id' } });
      return () => {};
    }
  }
}));

describe('Insurance Settings Page', () => {
  it('renders loading state initially', () => {
    global.fetch = vi.fn().mockImplementation(() => new Promise(() => {}));
    render(InsurancePage);
    expect(screen.getByText('Loading...')).toBeDefined();
  });

  it('renders policy and claims when fetch is successful', async () => {
    global.fetch = vi.fn().mockImplementation((url) => {
      if (url.includes('/policy')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ id: 'policy-1', policy_limit_cents: 500000 })
        });
      }
      if (url.includes('/claims')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([
            { id: 'claim-1', github_pr_url: 'https://github.com/a/b/pull/1', incident_date: '2023-01-01T00:00:00Z', amount_cents: 1000, status: 'PENDING' }
          ])
        });
      }
      return Promise.reject(new Error('Unknown url'));
    });

    render(InsurancePage);

    await waitFor(() => {
      expect(screen.getByText(/Policy ID:/)).toBeDefined();
    }, { timeout: 3000 });

    expect(screen.getByText('$5000.00')).toBeDefined();
    expect(screen.getByText('PENDING')).toBeDefined();
  });

  it('files a claim successfully', async () => {
    global.fetch = vi.fn().mockImplementation((url, options) => {
      if (url.includes('/policy')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ id: 'policy-1', policy_limit_cents: 500000 })
        });
      }
      if (url.includes('/claims') && !options) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve([])
        });
      }
      if (url.includes('/claims') && options?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            id: 'new-claim',
            github_pr_url: 'https://github.com/a/b/pull/2',
            incident_date: '2023-01-02T00:00:00Z',
            amount_cents: 5000,
            status: 'PENDING'
          })
        });
      }
      return Promise.reject(new Error('Unknown url'));
    });

    render(InsurancePage);

    await waitFor(() => {
      expect(screen.getByText('File Claim')).toBeDefined();
    }, { timeout: 3000 });

    const prInput = screen.getByLabelText('GitHub PR URL');
    const dateInput = screen.getByLabelText('Incident Date');
    const amountInput = screen.getByLabelText('Amount ($)');
    const fileButton = screen.getByText('File Claim');

    await fireEvent.input(prInput, { target: { value: 'https://github.com/a/b/pull/2' } });
    await fireEvent.input(dateInput, { target: { value: '2023-01-02' } });
    await fireEvent.input(amountInput, { target: { value: '50' } });

    await fireEvent.click(fileButton);

    await waitFor(() => {
      expect(screen.getByText('$50.00')).toBeDefined();
    }, { timeout: 3000 });
  });
});
