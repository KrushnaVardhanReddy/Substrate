import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { SubstrateHealthCard } from './SubstrateHealthCard';

describe('SubstrateHealthCard', () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it('renders loading state initially', () => {
    (global.fetch as jest.Mock).mockReturnValue(new Promise(() => {}));
    render(<SubstrateHealthCard entityId="test-entity" />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('renders health score on successful fetch', async () => {
    const mockData = { score: 95, riskLevel: 'Low' };
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockData,
    });

    render(<SubstrateHealthCard entityId="test-entity" />);

    await waitFor(() => {
      expect(screen.getByText('API Breaking Change Risk')).toBeInTheDocument();
    });

    expect(screen.getByText('95')).toBeInTheDocument();
    expect(screen.getByText('Low Risk')).toBeInTheDocument();
  });

  it('renders error state on fetch failure', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
    });

    render(<SubstrateHealthCard entityId="test-entity" />);

    await waitFor(() => {
      expect(screen.getByText('Failed to fetch health score')).toBeInTheDocument();
    });
  });
});
