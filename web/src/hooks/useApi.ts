import { useState, useCallback } from 'react';
import {
  NodeStatus,
  NetworkStats,
  Transaction,
  SystemMetrics,
  Wallet,
  Peer,
  Block,
  Health
} from '../types';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8545';

export const useApi = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleRequest = useCallback(async <T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> => {
    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data as T;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'An error occurred';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Node Management
  const getNodeStatus = useCallback(() => 
    handleRequest<NodeStatus>('/api/v1/node/status'), [handleRequest]);

  const getNetworkStats = useCallback(() => 
    handleRequest<NetworkStats>('/api/v1/network/stats'), [handleRequest]);

  const getSystemMetrics = useCallback(() => 
    handleRequest<SystemMetrics>('/api/v1/system/metrics'), [handleRequest]);

  // Transaction Management
  const getRecentTransactions = useCallback(() => 
    handleRequest<Transaction[]>('/api/v1/tx/recent'), [handleRequest]);

  const sendTransaction = useCallback((tx: Partial<Transaction>) => 
    handleRequest<Transaction>('/api/v1/tx/send', {
      method: 'POST',
      body: JSON.stringify(tx),
    }), [handleRequest]);

  // Wallet Management
  const getWallet = useCallback((address: string) => 
    handleRequest<Wallet>(`/api/v1/wallet/${address}`), [handleRequest]);

  const createWallet = useCallback((password: string) => 
    handleRequest<Wallet>('/api/v1/wallet/create', {
      method: 'POST',
      body: JSON.stringify({ password }),
    }), [handleRequest]);

  // Network Management
  const getPeers = useCallback(() => 
    handleRequest<Peer[]>('/api/v1/network/peers'), [handleRequest]);

  const connectPeer = useCallback((address: string) => 
    handleRequest<Peer>('/api/v1/network/connect', {
      method: 'POST',
      body: JSON.stringify({ address }),
    }), [handleRequest]);

  // Blockchain Management
  const getBlock = useCallback((height: number) => 
    handleRequest<Block>(`/api/v1/block/${height}`), [handleRequest]);

  const getHealth = useCallback(() => 
    handleRequest<Health>('/api/v1/health'), [handleRequest]);

  return {
    loading,
    error,
    getNodeStatus,
    getNetworkStats,
    getSystemMetrics,
    getRecentTransactions,
    sendTransaction,
    getWallet,
    createWallet,
    getPeers,
    connectPeer,
    getBlock,
    getHealth,
  };
}; 