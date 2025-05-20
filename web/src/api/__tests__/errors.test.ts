import { APIError, NetworkError, ValidationError, TimeoutError } from '../errors';

describe('API Errors', () => {
  describe('APIError', () => {
    it('should create a basic API error', () => {
      const error = new APIError('Test error');
      expect(error.message).toBe('Test error');
      expect(error.name).toBe('APIError');
      expect(error.status).toBeUndefined();
    });

    it('should create an API error with status', () => {
      const error = new APIError('Test error', 404);
      expect(error.message).toBe('Test error');
      expect(error.name).toBe('APIError');
      expect(error.status).toBe(404);
    });
  });

  describe('NetworkError', () => {
    it('should create a network error', () => {
      const error = new NetworkError('Network error');
      expect(error.message).toBe('Network error');
      expect(error.name).toBe('NetworkError');
      expect(error.status).toBe(0);
    });
  });

  describe('ValidationError', () => {
    it('should create a validation error', () => {
      const error = new ValidationError('Validation error');
      expect(error.message).toBe('Validation error');
      expect(error.name).toBe('ValidationError');
      expect(error.status).toBe(400);
    });

    it('should create a validation error with details', () => {
      const details = { field: 'email', message: 'Invalid email format' };
      const error = new ValidationError('Validation error', details);
      expect(error.message).toBe('Validation error');
      expect(error.name).toBe('ValidationError');
      expect(error.status).toBe(400);
      expect(error.details).toEqual(details);
    });
  });

  describe('TimeoutError', () => {
    it('should create a timeout error', () => {
      const error = new TimeoutError('Request timeout');
      expect(error.message).toBe('Request timeout');
      expect(error.name).toBe('TimeoutError');
      expect(error.status).toBe(408);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors correctly', () => {
      const error = new APIError('API error', 500);
      expect(error instanceof Error).toBe(true);
      expect(error instanceof APIError).toBe(true);
      expect(error.status).toBe(500);
    });

    it('should handle network errors correctly', () => {
      const error = new NetworkError('Network error');
      expect(error instanceof Error).toBe(true);
      expect(error instanceof NetworkError).toBe(true);
      expect(error.status).toBe(0);
    });

    it('should handle validation errors correctly', () => {
      const error = new ValidationError('Validation error');
      expect(error instanceof Error).toBe(true);
      expect(error instanceof ValidationError).toBe(true);
      expect(error.status).toBe(400);
    });

    it('should handle timeout errors correctly', () => {
      const error = new TimeoutError('Request timeout');
      expect(error instanceof Error).toBe(true);
      expect(error instanceof TimeoutError).toBe(true);
      expect(error.status).toBe(408);
    });
  });
}); 