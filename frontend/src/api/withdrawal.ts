import apiClient from './client';
import { Withdrawal, WithdrawalHistory, PaginatedResponse } from '@/types';

export const withdrawalApi = {
  // 获取提现规则列表
  getWithdrawals: (page: number = 1, limit: number = 10) => 
    apiClient.getPaginated<Withdrawal>(`/withdrawals?page=${page}&limit=${limit}`),

  // 创建提现规则
  createWithdrawal: (data: {
    asset: string;
    amount: number;
    address: string;
    network: string;
    memo?: string;
    rule_name: string;
    frequency: string;
  }) => apiClient.post<Withdrawal>('/withdrawals', data),

  // 获取提现规则详情
  getWithdrawalById: (withdrawalId: number) => 
    apiClient.get<Withdrawal>(`/withdrawals/${withdrawalId}`),

  // 更新提现规则
  updateWithdrawal: (withdrawalId: number, data: Partial<Withdrawal>) => 
    apiClient.put(`/withdrawals/${withdrawalId}`, data),

  // 切换提现规则状态
  toggleWithdrawal: (withdrawalId: number) => 
    apiClient.post(`/withdrawals/${withdrawalId}/toggle`),

  // 删除提现规则
  deleteWithdrawal: (withdrawalId: number) => 
    apiClient.delete(`/withdrawals/${withdrawalId}`),

  // 获取提现历史
  getWithdrawalHistory: (page: number = 1, limit: number = 20) => 
    apiClient.getPaginated<WithdrawalHistory>(`/withdrawals/history?page=${page}&limit=${limit}`),

  // 同步提现历史
  syncWithdrawalHistory: () => 
    apiClient.post('/withdrawals/history/sync'),

  // 获取提现统计
  getWithdrawalStats: () => 
    apiClient.get('/withdrawals/stats'),
};