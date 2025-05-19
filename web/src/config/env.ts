declare global {
  namespace NodeJS {
    interface ProcessEnv {
      REACT_APP_API_URL?: string;
      REACT_APP_API_TIMEOUT?: string;
      REACT_APP_ENABLE_ANALYTICS?: string;
      REACT_APP_ENABLE_DEBUG_MODE?: string;
      REACT_APP_DEFAULT_NETWORK?: string;
      REACT_APP_NETWORK_TIMEOUT?: string;
      REACT_APP_THEME?: string;
      REACT_APP_REFRESH_INTERVAL?: string;
    }
  }
}

interface EnvConfig {
  apiUrl: string;
  apiTimeout: number;
  enableAnalytics: boolean;
  enableDebugMode: boolean;
  defaultNetwork: string;
  networkTimeout: number;
  theme: string;
  refreshInterval: number;
}

const env: EnvConfig = {
  apiUrl: process.env.REACT_APP_API_URL || 'http://localhost:8545',
  apiTimeout: parseInt(process.env.REACT_APP_API_TIMEOUT || '30000', 10),
  enableAnalytics: process.env.REACT_APP_ENABLE_ANALYTICS === 'true',
  enableDebugMode: process.env.REACT_APP_ENABLE_DEBUG_MODE === 'true',
  defaultNetwork: process.env.REACT_APP_DEFAULT_NETWORK || 'testnet',
  networkTimeout: parseInt(process.env.REACT_APP_NETWORK_TIMEOUT || '5000', 10),
  theme: process.env.REACT_APP_THEME || 'light',
  refreshInterval: parseInt(process.env.REACT_APP_REFRESH_INTERVAL || '30000', 10),
};

export default env; 