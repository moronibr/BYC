import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Dashboard } from '../Dashboard';
import { useApi } from '../../hooks/useApi';

// Mock the useApi hook
jest.mock('../../hooks/useApi');

describe('Dashboard', () => {
  const mockApi = {
    getNodeStatus: jest.fn(),
    getNetworkStats: jest.fn(),
    getRecentTransactions: jest.fn(),
    getSystemMetrics: jest.fn(),
  };

  beforeEach(() => {
    (useApi as jest.Mock).mockReturnValue(mockApi);
    
    // Setup mock responses
    mockApi.getNodeStatus.mockResolvedValue({
      status: 'online',
      lastBlock: 1000,
      lastBlockTime: Date.now(),
      syncProgress: 100,
    });

    mockApi.getNetworkStats.mockResolvedValue({
      totalNodes: 100,
      activeNodes: 50,
      avgBlockTime: 10,
      hashRate: 1000,
    });

    mockApi.getRecentTransactions.mockResolvedValue([
      {
        hash: '0x123',
        from: '0xabc',
        to: '0xdef',
        amount: 100,
        status: 'confirmed',
        timestamp: Date.now(),
      },
    ]);

    mockApi.getSystemMetrics.mockResolvedValue({
      history: [
        {
          time: new Date().toISOString(),
          cpu: 50,
          memory: 60,
          network: 70,
        },
      ],
    });
  });

  it('renders loading state initially', () => {
    render(<Dashboard />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('renders dashboard data after loading', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
      expect(screen.getByText('Node Status')).toBeInTheDocument();
      expect(screen.getByText('Network Statistics')).toBeInTheDocument();
      expect(screen.getByText('System Metrics')).toBeInTheDocument();
      expect(screen.getByText('Recent Transactions')).toBeInTheDocument();
    });
  });

  it('displays node status information', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('Version:')).toBeInTheDocument();
      expect(screen.getByText('Network:')).toBeInTheDocument();
      expect(screen.getByText('Block Height:')).toBeInTheDocument();
      expect(screen.getByText('Peers:')).toBeInTheDocument();
      expect(screen.getByText('Uptime:')).toBeInTheDocument();
    });
  });

  it('displays network statistics', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('Total Nodes:')).toBeInTheDocument();
      expect(screen.getByText('Active Nodes:')).toBeInTheDocument();
      expect(screen.getByText('Average Block Time:')).toBeInTheDocument();
      expect(screen.getByText('Network Hash Rate:')).toBeInTheDocument();
    });
  });

  it('displays recent transactions', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('Hash: 0x123')).toBeInTheDocument();
      expect(screen.getByText('From: 0xabc')).toBeInTheDocument();
      expect(screen.getByText('To: 0xdef')).toBeInTheDocument();
      expect(screen.getByText('Amount: 100 BMC')).toBeInTheDocument();
      expect(screen.getByText('Status: confirmed')).toBeInTheDocument();
    });
  });

  it('handles API errors gracefully', async () => {
    mockApi.getNodeStatus.mockRejectedValue(new Error('API Error'));
    
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('Failed to fetch dashboard data')).toBeInTheDocument();
    });
  });

  it('refreshes data at specified interval', async () => {
    jest.useFakeTimers();
    render(<Dashboard refreshInterval={5000} />);

    await waitFor(() => {
      expect(mockApi.getNodeStatus).toHaveBeenCalledTimes(1);
    });

    jest.advanceTimersByTime(5000);

    await waitFor(() => {
      expect(mockApi.getNodeStatus).toHaveBeenCalledTimes(2);
    });

    jest.useRealTimers();
  });
}); 