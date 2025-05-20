import apiClient from '../client';
import {
  nodeService,
  walletService,
  transactionService,
  networkService,
  blockchainService,
  miningService,
  systemService,
} from '../service';

// Mock the API client
jest.mock('../client');

describe('API Services', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Node Service', () => {
    it('should get node info', async () => {
      const mockResponse = { version: '1.0.0', network: 'testnet' };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await nodeService.getInfo();
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/node/info');
    });

    it('should get node status', async () => {
      const mockResponse = { status: 'online', lastBlock: 1000 };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await nodeService.getStatus();
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/node/status');
    });
  });

  describe('Wallet Service', () => {
    it('should create wallet', async () => {
      const mockResponse = { address: '0x123', balance: 0 };
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      const result = await walletService.create({ password: 'test' });
      expect(result).toEqual(mockResponse);
      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/wallet/create', { password: 'test' });
    });

    it('should get wallet info', async () => {
      const mockResponse = { address: '0x123', balance: 1000 };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await walletService.getInfo('0x123');
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/wallet/0x123');
    });

    it('should send transaction', async () => {
      const mockResponse = { hash: '0xabc', status: 'pending' };
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      const tx = { to: '0x456', amount: 100 };
      const result = await walletService.send(tx);
      expect(result).toEqual(mockResponse);
      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/wallet/send', tx);
    });
  });

  describe('Transaction Service', () => {
    it('should get transaction info', async () => {
      const mockResponse = { hash: '0xabc', status: 'confirmed' };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await transactionService.getInfo('0xabc');
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/tx/0xabc');
    });

    it('should broadcast transaction', async () => {
      const mockResponse = { hash: '0xabc', status: 'pending' };
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      const tx = { to: '0x456', amount: 100 };
      const result = await transactionService.broadcast(tx);
      expect(result).toEqual(mockResponse);
      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/tx/broadcast', tx);
    });
  });

  describe('Network Service', () => {
    it('should get peers', async () => {
      const mockResponse = [{ address: '127.0.0.1:8545' }];
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await networkService.getPeers();
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/network/peers');
    });

    it('should connect to peer', async () => {
      const mockResponse = { address: '127.0.0.1:8545', status: 'connected' };
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      const result = await networkService.connect({ address: '127.0.0.1:8545' });
      expect(result).toEqual(mockResponse);
      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/network/connect', { address: '127.0.0.1:8545' });
    });
  });

  describe('Blockchain Service', () => {
    it('should get block', async () => {
      const mockResponse = { height: 1000, hash: '0xabc' };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await blockchainService.getBlock(1000);
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/block/1000');
    });

    it('should get latest block', async () => {
      const mockResponse = { height: 1000, hash: '0xabc' };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await blockchainService.getLatest();
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/block/latest');
    });
  });

  describe('Mining Service', () => {
    it('should get mining status', async () => {
      const mockResponse = { active: true, hashRate: 1000 };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await miningService.getStatus();
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/mining/status');
    });

    it('should start mining', async () => {
      const mockResponse = { status: 'started' };
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      const result = await miningService.start();
      expect(result).toEqual(mockResponse);
      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/mining/start');
    });
  });

  describe('System Service', () => {
    it('should get system metrics', async () => {
      const mockResponse = { cpu: 50, memory: 60 };
      (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

      const result = await systemService.getMetrics();
      expect(result).toEqual(mockResponse);
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/system/metrics');
    });

    it('should restart system', async () => {
      const mockResponse = { status: 'restarting' };
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      const result = await systemService.restart();
      expect(result).toEqual(mockResponse);
      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/system/restart');
    });
  });
}); 