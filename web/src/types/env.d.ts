declare namespace NodeJS {
  interface ProcessEnv {
    NODE_ENV: 'development' | 'production' | 'test';
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