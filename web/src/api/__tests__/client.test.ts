import axios from 'axios';
import apiClient from '../client';
import { API_CONFIG } from '../config';
import { requestInterceptor, responseInterceptor, errorInterceptor } from '../interceptors';

// Mock axios
jest.mock('axios');

describe('API Client', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should create axios instance with correct config', () => {
    expect(axios.create).toHaveBeenCalledWith({
      baseURL: API_CONFIG.baseURL,
      timeout: API_CONFIG.timeout,
      headers: API_CONFIG.headers,
    });
  });

  it('should add request interceptor', () => {
    expect(apiClient.interceptors.request.use).toHaveBeenCalledWith(requestInterceptor);
  });

  it('should add response interceptor', () => {
    expect(apiClient.interceptors.response.use).toHaveBeenCalledWith(responseInterceptor);
  });

  it('should add error interceptor', () => {
    expect(apiClient.interceptors.response.use).toHaveBeenCalledWith(
      expect.any(Function),
      errorInterceptor
    );
  });

  it('should make GET request', async () => {
    const mockResponse = { data: 'test' };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

    const response = await apiClient.get('/test');
    expect(response).toEqual(mockResponse);
    expect(apiClient.get).toHaveBeenCalledWith('/test');
  });

  it('should make POST request', async () => {
    const mockResponse = { data: 'test' };
    const data = { test: 'data' };
    (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

    const response = await apiClient.post('/test', data);
    expect(response).toEqual(mockResponse);
    expect(apiClient.post).toHaveBeenCalledWith('/test', data);
  });

  it('should make PUT request', async () => {
    const mockResponse = { data: 'test' };
    const data = { test: 'data' };
    (apiClient.put as jest.Mock).mockResolvedValue(mockResponse);

    const response = await apiClient.put('/test', data);
    expect(response).toEqual(mockResponse);
    expect(apiClient.put).toHaveBeenCalledWith('/test', data);
  });

  it('should make DELETE request', async () => {
    const mockResponse = { data: 'test' };
    (apiClient.delete as jest.Mock).mockResolvedValue(mockResponse);

    const response = await apiClient.delete('/test');
    expect(response).toEqual(mockResponse);
    expect(apiClient.delete).toHaveBeenCalledWith('/test');
  });

  it('should handle request with query parameters', async () => {
    const mockResponse = { data: 'test' };
    const params = { test: 'param' };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

    const response = await apiClient.get('/test', { params });
    expect(response).toEqual(mockResponse);
    expect(apiClient.get).toHaveBeenCalledWith('/test', { params });
  });

  it('should handle request with custom headers', async () => {
    const mockResponse = { data: 'test' };
    const headers = { 'Custom-Header': 'test' };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

    const response = await apiClient.get('/test', { headers });
    expect(response).toEqual(mockResponse);
    expect(apiClient.get).toHaveBeenCalledWith('/test', { headers });
  });

  it('should handle request with custom config', async () => {
    const mockResponse = { data: 'test' };
    const config = { timeout: 5000 };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

    const response = await apiClient.get('/test', config);
    expect(response).toEqual(mockResponse);
    expect(apiClient.get).toHaveBeenCalledWith('/test', config);
  });
}); 