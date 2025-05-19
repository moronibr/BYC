export interface NodeStatus {
  version: string;
  network: string;
  blockHeight: number;
  peers: number;
  uptime: string;
}

export interface NetworkStats {
  totalNodes: number;
  activeNodes: number;
  avgBlockTime: number;
  hashRate: number;
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  amount: number;
  status: 'pending' | 'confirmed' | 'failed';
  timestamp: string;
}

export interface SystemMetrics {
  history: {
    time: string;
    cpu: number;
    memory: number;
    network: number;
  }[];
}

export interface Wallet {
  address: string;
  balance: number;
  transactions: Transaction[];
}

export interface Peer {
  address: string;
  version: string;
  latency: number;
  lastSeen: string;
}

export interface Block {
  height: number;
  hash: string;
  timestamp: string;
  transactions: number;
  size: number;
}

export interface Health {
  status: 'healthy' | 'degraded' | 'unhealthy';
  issues: string[];
  lastCheck: string;
} 