import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Wallet } from '../Wallet';
import { useApi } from '../../hooks/useApi';

// Mock the useApi hook
jest.mock('../../hooks/useApi');

describe('Wallet', () => {
  const mockAddress = '0x123';
  const mockApi = {
    getWallet: jest.fn(),
    sendTransaction: jest.fn(),
  };

  beforeEach(() => {
    (useApi as jest.Mock).mockReturnValue(mockApi);
    
    // Setup mock responses
    mockApi.getWallet.mockResolvedValue({
      address: mockAddress,
      balance: 1000,
      transactions: [
        {
          hash: '0xabc',
          from: mockAddress,
          to: '0xdef',
          amount: 100,
          status: 'confirmed',
          timestamp: Date.now(),
        },
      ],
    });

    mockApi.sendTransaction.mockResolvedValue({
      hash: '0xnew',
      status: 'pending',
    });
  });

  it('renders loading state initially', () => {
    render(<Wallet address={mockAddress} />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('renders wallet information after loading', async () => {
    render(<Wallet address={mockAddress} />);

    await waitFor(() => {
      expect(screen.getByText('Wallet')).toBeInTheDocument();
      expect(screen.getByText('Wallet Information')).toBeInTheDocument();
      expect(screen.getByText('Recent Transactions')).toBeInTheDocument();
    });
  });

  it('displays wallet balance and address', async () => {
    render(<Wallet address={mockAddress} />);

    await waitFor(() => {
      expect(screen.getByText(`Address: ${mockAddress}`)).toBeInTheDocument();
      expect(screen.getByText('Balance: 1000 BMC')).toBeInTheDocument();
    });
  });

  it('displays transaction history', async () => {
    render(<Wallet address={mockAddress} />);

    await waitFor(() => {
      expect(screen.getByText('100 BMC to 0xdef')).toBeInTheDocument();
      expect(screen.getByText('Status: confirmed')).toBeInTheDocument();
    });
  });

  it('opens send dialog when clicking send button', async () => {
    render(<Wallet address={mockAddress} />);

    await waitFor(() => {
      expect(screen.getByText('Send BMC')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Send BMC'));

    expect(screen.getByText('Send BMC')).toBeInTheDocument();
    expect(screen.getByLabelText('Recipient Address')).toBeInTheDocument();
    expect(screen.getByLabelText('Amount')).toBeInTheDocument();
  });

  it('sends transaction when form is submitted', async () => {
    render(<Wallet address={mockAddress} />);

    await waitFor(() => {
      expect(screen.getByText('Send BMC')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Send BMC'));

    const recipientInput = screen.getByLabelText('Recipient Address');
    const amountInput = screen.getByLabelText('Amount');
    const sendButton = screen.getByText('Send');

    fireEvent.change(recipientInput, { target: { value: '0xdef' } });
    fireEvent.change(amountInput, { target: { value: '50' } });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(mockApi.sendTransaction).toHaveBeenCalledWith({
        to: '0xdef',
        amount: 50,
      });
    });
  });

  it('handles API errors gracefully', async () => {
    mockApi.getWallet.mockRejectedValue(new Error('API Error'));
    
    render(<Wallet address={mockAddress} />);

    await waitFor(() => {
      expect(screen.getByText('Failed to fetch wallet data')).toBeInTheDocument();
    });
  });

  it('validates transaction form inputs', async () => {
    render(<Wallet address={mockAddress} />);

    await waitFor(() => {
      expect(screen.getByText('Send BMC')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Send BMC'));

    const sendButton = screen.getByText('Send');
    expect(sendButton).toBeDisabled();

    const recipientInput = screen.getByLabelText('Recipient Address');
    const amountInput = screen.getByLabelText('Amount');

    fireEvent.change(recipientInput, { target: { value: '0xdef' } });
    expect(sendButton).toBeDisabled();

    fireEvent.change(amountInput, { target: { value: '50' } });
    expect(sendButton).not.toBeDisabled();
  });
}); 