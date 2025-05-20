export interface NodeInfo {
  version: string;
  network: string;
  uptime: number;
  lastBlock: number;
  peers: number;
}

export interface NodeStatus {
  status: 'online' | 'offline' | 'syncing';
  lastBlock: number;
  peers: number;
  uptime: number;
}

export interface WalletInfo {
  address: string;
  balance: number;
  transactions: Transaction[];
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  amount: number;
  status: 'pending' | 'confirmed' | 'failed';
  timestamp: number;
}

export interface Peer {
  address: string;
  version: string;
  lastSeen: number;
}

export interface Block {
  height: number;
  hash: string;
  previousHash: string;
  timestamp: number;
  transactions: Transaction[];
}

export interface MiningStatus {
  active: boolean;
  hashRate: number;
  difficulty: number;
  lastBlock: number;
}

export interface SystemMetrics {
  cpu: number;
  memory: number;
  disk: number;
  network: {
    bytesIn: number;
    bytesOut: number;
  };
} 