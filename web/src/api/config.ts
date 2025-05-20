export const API_CONFIG = {
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:3000',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
    'X-API-Version': '1.0.0',
    'X-API-Key': process.env.REACT_APP_API_KEY,
  },
  retry: {
    maxRetries: 3,
    retryDelay: 1000,
  },
  cache: {
    enabled: true,
    ttl: 60000, // 1 minute
  },
  rateLimit: {
    maxRequests: 100,
    perMilliseconds: 60000, // 1 minute
  },
  logging: {
    enabled: true,
    level: 'info',
  },
}; 