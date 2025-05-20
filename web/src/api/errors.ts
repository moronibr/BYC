export class APIError extends Error {
  status?: number;

  constructor(message: string, status?: number) {
    super(message);
    this.name = 'APIError';
    this.status = status;
  }
}

export class NetworkError extends APIError {
  constructor(message: string) {
    super(message, 0);
    this.name = 'NetworkError';
  }
}

export class ValidationError extends APIError {
  details?: Record<string, any>;

  constructor(message: string, details?: Record<string, any>) {
    super(message, 400);
    this.name = 'ValidationError';
    this.details = details;
  }
}

export class TimeoutError extends APIError {
  constructor(message: string) {
    super(message, 408);
    this.name = 'TimeoutError';
  }
} 