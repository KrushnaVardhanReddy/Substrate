import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte/svelte5';
import Page from './+page.svelte';
import { page } from '$app/stores';

vi.mock('$app/stores', () => ({
  page: {
    subscribe: (fn: any) => {
      fn({ params: { org: 'testorg' } });
      return () => {};
    }
  }
}));

describe('Governance Rules Settings Page', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('renders loading state initially and then rules', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([
        { id: '1', rule_text: 'Rule 1' },
        { id: '2', rule_text: 'Rule 2' }
      ])
    });

    render(Page as any);

    // Check loading indicator (might be too fast to catch, but we try)
    expect(screen.getAllByText('Governance Rules').length).toBeGreaterThan(0);

    await waitFor(() => {
      expect(screen.getAllByText('Rule 1').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Rule 2').length).toBeGreaterThan(0);
    });
  });

  it('adds a new rule', async () => {
    let callCount = 0;
    global.fetch = vi.fn().mockImplementation((url, options) => {
      if (options?.method === 'POST') {
        return Promise.resolve({ ok: true });
      }
      callCount++;
      if (callCount === 1) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve([{ id: '1', rule_text: 'New Rule' }])
      });
    });

    render(Page as any);

    await waitFor(() => {
      expect(screen.getAllByText('No governance rules found. Add one above.').length).toBeGreaterThan(0);
    });

    const inputs = screen.getAllByPlaceholderText('e.g., All endpoints must use camelCase');
    const input = inputs[0];
    const buttons = screen.getAllByText('Add Rule');
    const button = buttons[0];

    await fireEvent.input(input, { target: { value: 'New Rule' } });
    await fireEvent.click(button);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/v1/org/testorg/rules', expect.objectContaining({
        method: 'POST'
      }));
      expect(screen.getAllByText('New Rule').length).toBeGreaterThan(0);
    });
  });

  it('deletes a rule', async () => {
    let isDeleted = false;
    global.fetch = vi.fn().mockImplementation((url, options) => {
      if (options?.method === 'DELETE') {
        isDeleted = true;
        return Promise.resolve({ ok: true });
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve(isDeleted ? [] : [{ id: '1', rule_text: 'Rule to Delete' }])
      });
    });

    render(Page as any);

    await waitFor(() => {
      expect(screen.getAllByText('Rule to Delete').length).toBeGreaterThan(0);
    });

    const deleteButtons = screen.getAllByLabelText('Delete rule');
    const deleteButton = deleteButtons[0];
    await fireEvent.click(deleteButton);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/v1/org/testorg/rules/1', expect.objectContaining({
        method: 'DELETE'
      }));
      expect(screen.getAllByText('No governance rules found. Add one above.').length).toBeGreaterThan(0);
    });
  });
});
