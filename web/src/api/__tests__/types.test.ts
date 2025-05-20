import {
  NodeInfo,
  NodeStatus,
  WalletInfo,
  Transaction,
  Peer,
  Block,
  MiningStatus,
  SystemMetrics,
} from '../types';

describe('API Types', () => {
  describe('NodeInfo', () => {
    it('should have required properties', () => {
      const nodeInfo: NodeInfo = {
        version: '1.0.0',
        network: 'testnet',
        uptime: 3600,
        lastBlock: 1000,
        peers: 5,
      };

      expect(nodeInfo).toHaveProperty('version');
      expect(nodeInfo).toHaveProperty('network');
      expect(nodeInfo).toHaveProperty('uptime');
      expect(nodeInfo).toHaveProperty('lastBlock');
      expect(nodeInfo).toHaveProperty('peers');
    });
  });

  describe('NodeStatus', () => {
    it('should have required properties', () => {
      const nodeStatus: NodeStatus = {
        status: 'online',
        lastBlock: 1000,
        peers: 5,
        uptime: 3600,
      };

      expect(nodeStatus).toHaveProperty('status');
      expect(nodeStatus).toHaveProperty('lastBlock');
      expect(nodeStatus).toHaveProperty('peers');
      expect(nodeStatus).toHaveProperty('uptime');
    });
  });

  describe('WalletInfo', () => {
    it('should have required properties', () => {
      const walletInfo: WalletInfo = {
        address: '0x123',
        balance: 1000,
        transactions: [],
      };

      expect(walletInfo).toHaveProperty('address');
      expect(walletInfo).toHaveProperty('balance');
      expect(walletInfo).toHaveProperty('transactions');
    });
  });

  describe('Transaction', () => {
    it('should have required properties', () => {
      const transaction: Transaction = {
        hash: '0xabc',
        from: '0x123',
        to: '0x456',
        amount: 100,
        status: 'confirmed',
        timestamp: 1234567890,
      };

      expect(transaction).toHaveProperty('hash');
      expect(transaction).toHaveProperty('from');
      expect(transaction).toHaveProperty('to');
      expect(transaction).toHaveProperty('amount');
      expect(transaction).toHaveProperty('status');
      expect(transaction).toHaveProperty('timestamp');
    });
  });

  describe('Peer', () => {
    it('should have required properties', () => {
      const peer: Peer = {
        address: '127.0.0.1:8545',
        version: '1.0.0',
        lastSeen: 1234567890,
      };

      expect(peer).toHaveProperty('address');
      expect(peer).toHaveProperty('version');
      expect(peer).toHaveProperty('lastSeen');
    });
  });

  describe('Block', () => {
    it('should have required properties', () => {
      const block: Block = {
        height: 1000,
        hash: '0xabc',
        previousHash: '0xdef',
        timestamp: 1234567890,
        transactions: [],
      };

      expect(block).toHaveProperty('height');
      expect(block).toHaveProperty('hash');
      expect(block).toHaveProperty('previousHash');
      expect(block).toHaveProperty('timestamp');
      expect(block).toHaveProperty('transactions');
    });
  });

  describe('MiningStatus', () => {
    it('should have required properties', () => {
      const miningStatus: MiningStatus = {
        active: true,
        hashRate: 1000,
        difficulty: 1000000,
        lastBlock: 1000,
      };

      expect(miningStatus).toHaveProperty('active');
      expect(miningStatus).toHaveProperty('hashRate');
      expect(miningStatus).toHaveProperty('difficulty');
      expect(miningStatus).toHaveProperty('lastBlock');
    });
  });

  describe('SystemMetrics', () => {
    it('should have required properties', () => {
      const systemMetrics: SystemMetrics = {
        cpu: 50,
        memory: 60,
        disk: 70,
        network: {
          bytesIn: 1000,
          bytesOut: 2000,
        },
      };

      expect(systemMetrics).toHaveProperty('cpu');
      expect(systemMetrics).toHaveProperty('memory');
      expect(systemMetrics).toHaveProperty('disk');
      expect(systemMetrics).toHaveProperty('network');
      expect(systemMetrics.network).toHaveProperty('bytesIn');
      expect(systemMetrics.network).toHaveProperty('bytesOut');
    });
  });
}); 