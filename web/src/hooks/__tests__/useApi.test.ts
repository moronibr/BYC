import { renderHook, act } from '@testing-library/react';
import { useApi } from '../useApi';
import apiClient from '../../api/client';

// Mock the API client
jest.mock('../../api/client');

describe('useApi Hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should handle successful API calls', async () => {
    const mockResponse = { data: 'test' };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

    const { result } = renderHook(() => useApi());

    let response;
    await act(async () => {
      response = await result.current.getNodeStatus();
    });

    expect(response).toEqual(mockResponse);
    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/node/status');
  });

  it('should handle API errors', async () => {
    const mockError = new Error('API Error');
    (apiClient.get as jest.Mock).mockRejectedValue(mockError);

    const { result } = renderHook(() => useApi());

    let error;
    await act(async () => {
      try {
        await result.current.getNodeStatus();
      } catch (e) {
        error = e;
      }
    });

    expect(error).toEqual(mockError);
  });

  it('should handle loading state', async () => {
    const mockResponse = { data: 'test' };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

    const { result } = renderHook(() => useApi());

    expect(result.current.loading).toBe(false);

    let promise;
    act(() => {
      promise = result.current.getNodeStatus();
    });

    expect(result.current.loading).toBe(true);

    await act(async () => {
      await promise;
    });

    expect(result.current.loading).toBe(false);
  });

  it('should handle multiple concurrent requests', async () => {
    const mockResponse1 = { data: 'test1' };
    const mockResponse2 = { data: 'test2' };
    (apiClient.get as jest.Mock)
      .mockResolvedValueOnce(mockResponse1)
      .mockResolvedValueOnce(mockResponse2);

    const { result } = renderHook(() => useApi());

    let response1, response2;
    await act(async () => {
      const [res1, res2] = await Promise.all([
        result.current.getNodeStatus(),
        result.current.getNetworkStats()
      ]);
      response1 = res1;
      response2 = res2;
    });

    expect(response1).toEqual(mockResponse1);
    expect(response2).toEqual(mockResponse2);
    expect(apiClient.get).toHaveBeenCalledTimes(2);
  });

  it('should handle request cancellation', async () => {
    const mockResponse = { data: 'test' };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);

    const { result, unmount } = renderHook(() => useApi());

    let promise;
    act(() => {
      promise = result.current.getNodeStatus();
    });

    unmount();

    await act(async () => {
      await promise;
    });

    expect(apiClient.get).toHaveBeenCalled();
  });

  it('should handle request retries', async () => {
    const mockError = new Error('API Error');
    const mockResponse = { data: 'test' };
    (apiClient.get as jest.Mock)
      .mockRejectedValueOnce(mockError)
      .mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useApi());

    let response;
    await act(async () => {
      response = await result.current.getNodeStatus();
    });

    expect(response).toEqual(mockResponse);
    expect(apiClient.get).toHaveBeenCalledTimes(2);
  });

  it('should handle request timeout', async () => {
    const mockError = new Error('Request timeout');
    (apiClient.get as jest.Mock).mockRejectedValue(mockError);

    const { result } = renderHook(() => useApi());

    let error;
    await act(async () => {
      try {
        await result.current.getNodeStatus();
      } catch (e) {
        error = e;
      }
    });

    expect(error).toEqual(mockError);
  });
}); 