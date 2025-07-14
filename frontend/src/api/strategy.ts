import apiClient from './client';
import { Strategy, CreateStrategyForm, PaginatedResponse } from '@/types';

export const strategyApi = {
  // 获取策略列表
  getStrategies: (page: number = 1, limit: number = 10) => 
    apiClient.getPaginated<Strategy>(`/strategies?page=${page}&limit=${limit}`),

  // 创建策略
  createStrategy: (data: CreateStrategyForm) => 
    apiClient.post<Strategy>('/strategies', data),

  // 获取策略详情
  getStrategyById: (strategyId: number) => 
    apiClient.get<Strategy>(`/strategies/${strategyId}`),

  // 更新策略
  updateStrategy: (strategyId: number, data: Partial<CreateStrategyForm>) => 
    apiClient.put(`/strategies/${strategyId}`, data),

  // 切换策略状态
  toggleStrategy: (strategyId: number) => 
    apiClient.post(`/strategies/${strategyId}/toggle`),

  // 删除策略
  deleteStrategy: (strategyId: number) => 
    apiClient.delete(`/strategies/${strategyId}`),

  // 获取策略统计
  getStrategyStats: (strategyId: number) => 
    apiClient.get(`/strategies/${strategyId}/stats`),
};