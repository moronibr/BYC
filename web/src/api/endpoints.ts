// Node endpoints
export const NODE = {
  INFO: '/api/v1/node/info',
  STATUS: '/api/v1/node/status',
  VERSION: '/api/v1/node/version',
  CONFIG: '/api/v1/node/config',
};

// Wallet endpoints
export const WALLET = {
  CREATE: '/api/v1/wallet/create',
  LIST: '/api/v1/wallet/list',
  INFO: (address: string) => `/api/v1/wallet/${address}`,
  BALANCE: (address: string) => `/api/v1/wallet/${address}/balance`,
  TRANSACTIONS: (address: string) => `/api/v1/wallet/${address}/transactions`,
  SEND: '/api/v1/wallet/send',
  BACKUP: (address: string) => `/api/v1/wallet/${address}/backup`,
  RESTORE: '/api/v1/wallet/restore',
};

// Transaction endpoints
export const TRANSACTION = {
  INFO: (hash: string) => `/api/v1/tx/${hash}`,
  STATUS: (hash: string) => `/api/v1/tx/${hash}/status`,
  BROADCAST: '/api/v1/tx/broadcast',
  PENDING: '/api/v1/tx/pending',
};

// Network endpoints
export const NETWORK = {
  PEERS: '/api/v1/network/peers',
  CONNECT: '/api/v1/network/connect',
  DISCONNECT: '/api/v1/network/disconnect',
  STATUS: '/api/v1/network/status',
  SYNC: '/api/v1/network/sync',
};

// Blockchain endpoints
export const BLOCKCHAIN = {
  BLOCK: (height: number) => `/api/v1/block/${height}`,
  BLOCK_BY_HASH: (hash: string) => `/api/v1/block/hash/${hash}`,
  LATEST: '/api/v1/block/latest',
  HEIGHT: '/api/v1/block/height',
  DIFFICULTY: '/api/v1/block/difficulty',
};

// Mining endpoints
export const MINING = {
  STATUS: '/api/v1/mining/status',
  START: '/api/v1/mining/start',
  STOP: '/api/v1/mining/stop',
  REWARD: '/api/v1/mining/reward',
  DIFFICULTY: '/api/v1/mining/difficulty',
};

// System endpoints
export const SYSTEM = {
  LOGS: '/api/v1/system/logs',
  METRICS: '/api/v1/system/metrics',
  CONFIG: '/api/v1/system/config',
  RESTART: '/api/v1/system/restart',
  SHUTDOWN: '/api/v1/system/shutdown',
}; 