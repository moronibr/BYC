// Node types
export interface NodeInfo {
  version: string;
  network: string;
  uptime: number;
  peers: number;
  blocks: number;
  difficulty: number;
}

export interface NodeStatus {
  status: 'online' | 'offline' | 'syncing';
  lastBlock: number;
  lastBlockTime: number;
  syncProgress: number;
}

// Wallet types
export interface Wallet {
  address: string;
  balance: number;
  nonce: number;
  createdAt: number;
  lastUsed: number;
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  amount: number;
  fee: number;
  timestamp: number;
  status: 'pending' | 'confirmed' | 'failed';
  blockHeight?: number;
}

// Network types
export interface Peer {
  address: string;
  version: string;
  lastSeen: number;
  height: number;
  latency: number;
}

export interface NetworkStatus {
  peers: number;
  height: number;
  difficulty: number;
  hashRate: number;
  mempoolSize: number;
}

// Blockchain types
export interface Block {
  height: number;
  hash: string;
  previousHash: string;
  timestamp: number;
  transactions: number;
  size: number;
  difficulty: number;
  miner: string;
  reward: number;
}

// Mining types
export interface MiningStatus {
  active: boolean;
  hashRate: number;
  difficulty: number;
  reward: number;
  lastBlock: number;
}

// System types
export interface SystemMetrics {
  cpu: {
    usage: number;
    cores: number;
  };
  memory: {
    total: number;
    used: number;
    free: number;
  };
  disk: {
    total: number;
    used: number;
    free: number;
  };
  network: {
    bytesIn: number;
    bytesOut: number;
    connections: number;
  };
}

export interface SystemConfig {
  network: string;
  rpcPort: number;
  p2pPort: number;
  dataDir: string;
  logLevel: string;
  maxPeers: number;
  maxConnections: number;
  syncMode: string;
  miningEnabled: boolean;
}

// Error types
export interface ApiError {
  code: number;
  message: string;
  details?: any;
}

// Response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: ApiError;
} 