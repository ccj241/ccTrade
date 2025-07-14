import axios, { AxiosError, AxiosInstance, AxiosRequestConfig } from 'axios';
import { ApiResponse } from '@/types';
import toast from 'react-hot-toast';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

class ApiClient {
  private instance: AxiosInstance;

  constructor() {
    this.instance = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // 请求拦截器
    this.instance.interceptors.request.use(
      (config) => {
        console.log('Request interceptor - URL:', config.url);
        console.log('Request interceptor - Base URL:', config.baseURL);
        console.log('Request interceptor - Full URL:', config.baseURL + config.url);
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.instance.interceptors.response.use(
      (response) => {
        return response;
      },
      (error: AxiosError<ApiResponse>) => {
        if (error.response) {
          const { status, data } = error.response;
          
          switch (status) {
            case 401:
              // 未授权，清除token并跳转到登录页
              localStorage.removeItem('token');
              window.location.href = '/login';
              toast.error(data.message || '登录已过期，请重新登录');
              break;
            case 403:
              toast.error(data.message || '没有权限执行此操作');
              break;
            case 404:
              toast.error(data.message || '请求的资源不存在');
              break;
            case 429:
              toast.error(data.message || '请求过于频繁，请稍后再试');
              break;
            case 500:
              toast.error(data.message || '服务器错误，请稍后再试');
              break;
            default:
              toast.error(data.message || '请求失败');
          }
        } else if (error.request) {
          toast.error('网络错误，请检查网络连接');
        } else {
          toast.error('请求配置错误');
        }
        
        return Promise.reject(error);
      }
    );
  }

  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.get<ApiResponse<T>>(url, config);
    return response.data.data as T;
  }

  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    console.log('POST request to:', url);
    console.log('Base URL:', this.instance.defaults.baseURL);
    console.log('Full URL will be:', this.instance.defaults.baseURL + url);
    const response = await this.instance.post<ApiResponse<T>>(url, data, config);
    return response.data.data as T;
  }

  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.put<ApiResponse<T>>(url, data, config);
    return response.data.data as T;
  }

  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.delete<ApiResponse<T>>(url, config);
    return response.data.data as T;
  }

  // 分页请求的特殊处理
  async getPaginated<T = any>(url: string, config?: AxiosRequestConfig) {
    const response = await this.instance.get(url, config);
    return response.data;
  }
}

export default new ApiClient();