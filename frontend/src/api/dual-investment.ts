import apiClient from './client';
import { DualInvestmentProduct, DualInvestmentStrategy, PaginatedResponse } from '@/types';

export const dualInvestmentApi = {
  // 获取双币投资产品列表
  getProducts: () => 
    apiClient.get<DualInvestmentProduct[]>('/dual/products'),

  // 获取双币投资策略列表
  getStrategies: (page: number = 1, limit: number = 10) => 
    apiClient.getPaginated<DualInvestmentStrategy>(`/dual/strategies?page=${page}&limit=${limit}`),

  // 创建双币投资策略
  createStrategy: (data: {
    name: string;
    symbol: string;
    target_price: number;
    min_apy: number;
    max_investment: number;
    single_investment: number;
    is_call: boolean;
    auto_compound: boolean;
  }) => apiClient.post<DualInvestmentStrategy>('/dual/strategy', data),

  // 获取双币投资策略详情
  getStrategyById: (strategyId: number) => 
    apiClient.get<DualInvestmentStrategy>(`/dual/strategy/${strategyId}`),

  // 更新双币投资策略
  updateStrategy: (strategyId: number, data: Partial<DualInvestmentStrategy>) => 
    apiClient.put(`/dual/strategy/${strategyId}`, data),

  // 切换双币投资策略状态
  toggleStrategy: (strategyId: number) => 
    apiClient.post(`/dual/strategy/${strategyId}/toggle`),

  // 删除双币投资策略
  deleteStrategy: (strategyId: number) => 
    apiClient.delete(`/dual/strategy/${strategyId}`),

  // 获取用户双币投资订单
  getOrders: () => 
    apiClient.get('/dual/orders'),

  // 获取双币投资统计
  getStats: () => 
    apiClient.get('/dual/stats'),

  // 同步双币投资产品（管理员）
  syncProducts: () => 
    apiClient.post('/admin/dual/products/sync'),
};