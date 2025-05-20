import React from 'react';
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Network } from '../Network';
import { useApi } from '../../hooks/useApi';

// Mock the useApi hook
jest.mock('../../hooks/useApi');

describe('Network', () => {
  const mockApi = {
    getPeers: jest.fn(),
    getHealth: jest.fn(),
    connectPeer: jest.fn(),
  };

  beforeEach(() => {
    (useApi as jest.Mock).mockReturnValue(mockApi);
    
    // Setup mock responses
    mockApi.getPeers.mockResolvedValue([
      {
        address: '127.0.0.1:8545',
        version: '1.0.0',
        lastSeen: Date.now(),
        height: 1000,
        latency: 50,
      },
    ]);

    mockApi.getHealth.mockResolvedValue({
      status: 'healthy',
      issues: [],
      lastCheck: Date.now(),
    });

    mockApi.connectPeer.mockResolvedValue({
      address: '127.0.0.1:8546',
      version: '1.0.0',
      lastSeen: Date.now(),
      height: 1000,
      latency: 50,
    });
  });

  it('renders loading state initially', async () => {
    await act(async () => {
      render(<Network />);
    });
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
  });

  it('renders network information after loading', async () => {
    await act(async () => {
      render(<Network />);
    });

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Network' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { name: 'Network Health' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { name: 'Connect to Peer' })).toBeInTheDocument();
      expect(screen.getByRole('heading', { name: /Connected Peers/ })).toBeInTheDocument();
    });
  });

  it('displays network health status', async () => {
    await act(async () => {
      render(<Network />);
    });

    await waitFor(() => {
      expect(screen.getByText(/healthy/i)).toBeInTheDocument();
    });
  });

  it('displays connected peers', async () => {
    await act(async () => {
      render(<Network />);
    });

    await waitFor(() => {
      expect(screen.getByText('127.0.0.1:8545')).toBeInTheDocument();
      expect(screen.getByText(/Version: 1.0.0/)).toBeInTheDocument();
      expect(screen.getByText(/Latency: 50ms/)).toBeInTheDocument();
    });
  });

  it('connects to a new peer', async () => {
    await act(async () => {
      render(<Network />);
    });

    await waitFor(() => {
      expect(screen.getByLabelText('Peer Address')).toBeInTheDocument();
    });

    const addressInput = screen.getByLabelText('Peer Address');
    const connectButton = screen.getByRole('button', { name: 'Connect' });

    await act(async () => {
      fireEvent.change(addressInput, { target: { value: '127.0.0.1:8546' } });
      fireEvent.click(connectButton);
    });

    await waitFor(() => {
      expect(mockApi.connectPeer).toHaveBeenCalledWith('127.0.0.1:8546');
    });
  });

  it('handles API errors gracefully', async () => {
    mockApi.getPeers.mockRejectedValue(new Error('API Error'));
    
    await act(async () => {
      render(<Network />);
    });

    await waitFor(() => {
      expect(screen.getByText(/API Error/)).toBeInTheDocument();
    });
  });

  it('validates peer connection form', async () => {
    await act(async () => {
      render(<Network />);
    });

    await waitFor(() => {
      expect(screen.getByLabelText('Peer Address')).toBeInTheDocument();
    });

    const connectButton = screen.getByRole('button', { name: 'Connect' });
    expect(connectButton).toBeDisabled();

    await act(async () => {
      fireEvent.change(screen.getByLabelText('Peer Address'), { target: { value: '127.0.0.1:8546' } });
    });
    
    expect(connectButton).not.toBeDisabled();
  });

  it('displays network issues when health is degraded', async () => {
    mockApi.getHealth.mockResolvedValue({
      status: 'degraded',
      issues: ['High latency', 'Low peer count'],
      lastCheck: Date.now(),
    });

    await act(async () => {
      render(<Network />);
    });

    await waitFor(() => {
      expect(screen.getByText(/degraded/i)).toBeInTheDocument();
      expect(screen.getByText('High latency')).toBeInTheDocument();
      expect(screen.getByText('Low peer count')).toBeInTheDocument();
    });
  });
}); 