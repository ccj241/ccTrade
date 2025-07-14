import apiClient from './client';
import { FuturesStrategy, FuturesPosition, CreateFuturesStrategyForm, PaginatedResponse } from '@/types';

export const futuresApi = {
  // 获取期货策略列表
  getFuturesStrategies: (page: number = 1, limit: number = 10) => 
    apiClient.getPaginated<FuturesStrategy>(`/futures/strategies?page=${page}&limit=${limit}`),

  // 创建期货策略
  createFuturesStrategy: (data: CreateFuturesStrategyForm) => 
    apiClient.post<FuturesStrategy>('/futures/strategy', data),

  // 获取期货策略详情
  getFuturesStrategyById: (strategyId: number) => 
    apiClient.get<FuturesStrategy>(`/futures/strategy/${strategyId}`),

  // 更新期货策略
  updateFuturesStrategy: (strategyId: number, data: Partial<CreateFuturesStrategyForm>) => 
    apiClient.put(`/futures/strategy/${strategyId}`, data),

  // 切换期货策略状态
  toggleFuturesStrategy: (strategyId: number) => 
    apiClient.post(`/futures/strategy/${strategyId}/toggle`),

  // 删除期货策略
  deleteFuturesStrategy: (strategyId: number) => 
    apiClient.delete(`/futures/strategy/${strategyId}`),

  // 获取持仓列表
  getPositions: () => 
    apiClient.get<FuturesPosition[]>('/futures/positions'),

  // 同步持仓
  syncPositions: () => 
    apiClient.post('/futures/positions/sync'),

  // 获取期货统计
  getFuturesStats: () => 
    apiClient.get('/futures/stats'),
};