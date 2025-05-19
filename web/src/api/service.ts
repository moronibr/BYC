import apiClient from './client';
import {
  NODE,
  WALLET,
  TRANSACTION,
  NETWORK,
  BLOCKCHAIN,
  MINING,
  SYSTEM,
} from './endpoints';

// Node service
export const nodeService = {
  getInfo: () => apiClient.get(NODE.INFO),
  getStatus: () => apiClient.get(NODE.STATUS),
  getVersion: () => apiClient.get(NODE.VERSION),
  getConfig: () => apiClient.get(NODE.CONFIG),
};

// Wallet service
export const walletService = {
  create: (data: any) => apiClient.post(WALLET.CREATE, data),
  list: () => apiClient.get(WALLET.LIST),
  getInfo: (address: string) => apiClient.get(WALLET.INFO(address)),
  getBalance: (address: string) => apiClient.get(WALLET.BALANCE(address)),
  getTransactions: (address: string) => apiClient.get(WALLET.TRANSACTIONS(address)),
  send: (data: any) => apiClient.post(WALLET.SEND, data),
  backup: (address: string) => apiClient.get(WALLET.BACKUP(address)),
  restore: (data: any) => apiClient.post(WALLET.RESTORE, data),
};

// Transaction service
export const transactionService = {
  getInfo: (hash: string) => apiClient.get(TRANSACTION.INFO(hash)),
  getStatus: (hash: string) => apiClient.get(TRANSACTION.STATUS(hash)),
  broadcast: (data: any) => apiClient.post(TRANSACTION.BROADCAST, data),
  getPending: () => apiClient.get(TRANSACTION.PENDING),
};

// Network service
export const networkService = {
  getPeers: () => apiClient.get(NETWORK.PEERS),
  connect: (data: any) => apiClient.post(NETWORK.CONNECT, data),
  disconnect: (data: any) => apiClient.post(NETWORK.DISCONNECT, data),
  getStatus: () => apiClient.get(NETWORK.STATUS),
  sync: () => apiClient.post(NETWORK.SYNC),
};

// Blockchain service
export const blockchainService = {
  getBlock: (height: number) => apiClient.get(BLOCKCHAIN.BLOCK(height)),
  getBlockByHash: (hash: string) => apiClient.get(BLOCKCHAIN.BLOCK_BY_HASH(hash)),
  getLatest: () => apiClient.get(BLOCKCHAIN.LATEST),
  getHeight: () => apiClient.get(BLOCKCHAIN.HEIGHT),
  getDifficulty: () => apiClient.get(BLOCKCHAIN.DIFFICULTY),
};

// Mining service
export const miningService = {
  getStatus: () => apiClient.get(MINING.STATUS),
  start: () => apiClient.post(MINING.START),
  stop: () => apiClient.post(MINING.STOP),
  getReward: () => apiClient.get(MINING.REWARD),
  getDifficulty: () => apiClient.get(MINING.DIFFICULTY),
};

// System service
export const systemService = {
  getLogs: () => apiClient.get(SYSTEM.LOGS),
  getMetrics: () => apiClient.get(SYSTEM.METRICS),
  getConfig: () => apiClient.get(SYSTEM.CONFIG),
  restart: () => apiClient.post(SYSTEM.RESTART),
  shutdown: () => apiClient.post(SYSTEM.SHUTDOWN),
}; 