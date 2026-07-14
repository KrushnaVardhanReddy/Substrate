import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { SubstrateGraphCard } from './SubstrateGraphCard';

describe('SubstrateGraphCard', () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it('renders loading state initially', () => {
    (global.fetch as jest.Mock).mockReturnValue(new Promise(() => {}));
    render(<SubstrateGraphCard entityId="test-entity" />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('renders graph on successful fetch', async () => {
    const mockEdges = [
      { provider: 'service-a', consumer: 'service-b', status: 'active' }
    ];
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockEdges,
    });

    render(<SubstrateGraphCard entityId="test-entity" />);

    await waitFor(() => {
      expect(screen.getByText('Substrate Dependency Graph (test-entity)')).toBeInTheDocument();
    });

    expect(screen.getByTestId('react-flow-container')).toBeInTheDocument();
  });

  it('renders error state on fetch failure', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
    });

    render(<SubstrateGraphCard entityId="test-entity" />);

    await waitFor(() => {
      expect(screen.getByText('Failed to fetch dependency graph')).toBeInTheDocument();
    });
  });
});
