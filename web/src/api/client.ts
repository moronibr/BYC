import axios, { AxiosInstance } from 'axios';
import { API_CONFIG } from './config';
import { requestInterceptor, responseInterceptor, errorInterceptor } from './interceptors';

const apiClient: AxiosInstance = axios.create({
  baseURL: API_CONFIG.baseURL,
  timeout: API_CONFIG.timeout,
  headers: API_CONFIG.headers,
});

apiClient.interceptors.request.use(requestInterceptor);
apiClient.interceptors.response.use(responseInterceptor, errorInterceptor);

export default apiClient; 