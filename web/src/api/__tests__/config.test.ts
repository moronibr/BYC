import { API_CONFIG } from '../config';

describe('API Configuration', () => {
  it('should have required properties', () => {
    expect(API_CONFIG).toHaveProperty('baseURL');
    expect(API_CONFIG).toHaveProperty('timeout');
    expect(API_CONFIG).toHaveProperty('headers');
  });

  it('should have correct baseURL', () => {
    expect(API_CONFIG.baseURL).toBe(process.env.REACT_APP_API_URL || 'http://localhost:3000');
  });

  it('should have correct timeout', () => {
    expect(API_CONFIG.timeout).toBe(30000);
  });

  it('should have correct headers', () => {
    expect(API_CONFIG.headers).toHaveProperty('Content-Type');
    expect(API_CONFIG.headers['Content-Type']).toBe('application/json');
  });

  it('should have correct API version', () => {
    expect(API_CONFIG.headers).toHaveProperty('X-API-Version');
    expect(API_CONFIG.headers['X-API-Version']).toBe('1.0.0');
  });

  it('should have correct API key', () => {
    expect(API_CONFIG.headers).toHaveProperty('X-API-Key');
    expect(API_CONFIG.headers['X-API-Key']).toBe(process.env.REACT_APP_API_KEY);
  });

  it('should have correct retry configuration', () => {
    expect(API_CONFIG).toHaveProperty('retry');
    expect(API_CONFIG.retry).toHaveProperty('maxRetries');
    expect(API_CONFIG.retry).toHaveProperty('retryDelay');
    expect(API_CONFIG.retry.maxRetries).toBe(3);
    expect(API_CONFIG.retry.retryDelay).toBe(1000);
  });

  it('should have correct cache configuration', () => {
    expect(API_CONFIG).toHaveProperty('cache');
    expect(API_CONFIG.cache).toHaveProperty('enabled');
    expect(API_CONFIG.cache).toHaveProperty('ttl');
    expect(API_CONFIG.cache.enabled).toBe(true);
    expect(API_CONFIG.cache.ttl).toBe(60000);
  });

  it('should have correct rate limiting configuration', () => {
    expect(API_CONFIG).toHaveProperty('rateLimit');
    expect(API_CONFIG.rateLimit).toHaveProperty('maxRequests');
    expect(API_CONFIG.rateLimit).toHaveProperty('perMilliseconds');
    expect(API_CONFIG.rateLimit.maxRequests).toBe(100);
    expect(API_CONFIG.rateLimit.perMilliseconds).toBe(60000);
  });

  it('should have correct logging configuration', () => {
    expect(API_CONFIG).toHaveProperty('logging');
    expect(API_CONFIG.logging).toHaveProperty('enabled');
    expect(API_CONFIG.logging).toHaveProperty('level');
    expect(API_CONFIG.logging.enabled).toBe(true);
    expect(API_CONFIG.logging.level).toBe('info');
  });
}); 