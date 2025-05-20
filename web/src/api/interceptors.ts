import { AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { API_CONFIG } from './config';
import { APIError, NetworkError, ValidationError, TimeoutError } from './errors';

export const requestInterceptor = (config: AxiosRequestConfig): AxiosRequestConfig => {
  const headers = {
    ...API_CONFIG.headers,
    ...config.headers,
  };

  return {
    ...config,
    headers,
  };
};

export const responseInterceptor = (response: AxiosResponse): any => {
  return response.data;
};

export const errorInterceptor = (error: AxiosError): never => {
  if (error.code === 'ECONNABORTED') {
    if (error.message.includes('timeout')) {
      throw new TimeoutError('Request timeout');
    }
    throw new NetworkError('Network error');
  }

  if (error.response) {
    const { status, data } = error.response;

    if (status === 400) {
      throw new ValidationError(data.message, data.details);
    }

    throw new APIError(data.message || 'API error', status);
  }

  throw new APIError(error.message || 'Unknown error');
}; 