import { requestInterceptor, responseInterceptor, errorInterceptor } from '../interceptors';
import { API_CONFIG } from '../config';
import { APIError, NetworkError, ValidationError, TimeoutError } from '../errors';

describe('API Interceptors', () => {
  describe('Request Interceptor', () => {
    it('should add default headers', () => {
      const config = {
        headers: {},
      };

      const result = requestInterceptor(config);
      expect(result.headers).toHaveProperty('Content-Type');
      expect(result.headers['Content-Type']).toBe('application/json');
    });

    it('should add API version header', () => {
      const config = {
        headers: {},
      };

      const result = requestInterceptor(config);
      expect(result.headers).toHaveProperty('X-API-Version');
      expect(result.headers['X-API-Version']).toBe(API_CONFIG.headers['X-API-Version']);
    });

    it('should add API key header', () => {
      const config = {
        headers: {},
      };

      const result = requestInterceptor(config);
      expect(result.headers).toHaveProperty('X-API-Key');
      expect(result.headers['X-API-Key']).toBe(API_CONFIG.headers['X-API-Key']);
    });

    it('should preserve existing headers', () => {
      const config = {
        headers: {
          'Custom-Header': 'test',
        },
      };

      const result = requestInterceptor(config);
      expect(result.headers).toHaveProperty('Custom-Header');
      expect(result.headers['Custom-Header']).toBe('test');
    });
  });

  describe('Response Interceptor', () => {
    it('should return response data', () => {
      const response = {
        data: { test: 'data' },
      };

      const result = responseInterceptor(response);
      expect(result).toEqual({ test: 'data' });
    });

    it('should handle empty response', () => {
      const response = {
        data: null,
      };

      const result = responseInterceptor(response);
      expect(result).toBeNull();
    });
  });

  describe('Error Interceptor', () => {
    it('should handle network errors', () => {
      const error = {
        message: 'Network Error',
        code: 'ECONNABORTED',
      };

      expect(() => errorInterceptor(error)).toThrow(NetworkError);
    });

    it('should handle timeout errors', () => {
      const error = {
        message: 'timeout of 30000ms exceeded',
        code: 'ECONNABORTED',
      };

      expect(() => errorInterceptor(error)).toThrow(TimeoutError);
    });

    it('should handle validation errors', () => {
      const error = {
        response: {
          status: 400,
          data: {
            message: 'Validation error',
            details: { field: 'email' },
          },
        },
      };

      expect(() => errorInterceptor(error)).toThrow(ValidationError);
    });

    it('should handle API errors', () => {
      const error = {
        response: {
          status: 500,
          data: {
            message: 'Internal server error',
          },
        },
      };

      expect(() => errorInterceptor(error)).toThrow(APIError);
    });

    it('should handle unknown errors', () => {
      const error = {
        message: 'Unknown error',
      };

      expect(() => errorInterceptor(error)).toThrow(APIError);
    });
  });
}); 