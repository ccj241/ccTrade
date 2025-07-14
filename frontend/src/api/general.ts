import apiClient from './client';
import { Balance, BalanceResponse, Order, Symbol, Price } from '@/types';

export const generalApi = {
  // 获取余额
  getBalance: () => 
    apiClient.get<BalanceResponse>('/balance'),

  // 获取订单列表
  getOrders: (symbol?: string) => {
    const params = symbol ? `?symbol=${symbol}` : '';
    return apiClient.get<Order[]>(`/orders${params}`);
  },

  // 创建订单
  createOrder: (data: {
    symbol: string;
    side: string;
    type: string;
    quantity: number;
    price?: number;
  }) => apiClient.post<Order>('/order', data),

  // 取消订单
  cancelOrder: (orderId: string) => 
    apiClient.delete(`/order/${orderId}`),

  // 批量取消订单
  batchCancelOrders: (orderIds: string[]) => 
    apiClient.post('/batch-cancel-orders', { order_ids: orderIds }),

  // 获取已取消订单
  getCancelledOrders: () => 
    apiClient.get<Order[]>('/cancelled-orders'),

  // 获取交易对列表
  getTradingSymbols: () => 
    apiClient.get<Symbol[]>('/trading-symbols'),

  // 获取期货交易对列表
  getFuturesTradingSymbols: () => 
    apiClient.get<Symbol[]>('/futures-trading-symbols'),

  // 获取价格
  getPrice: (symbol: string) => 
    apiClient.get<Price>(`/price?symbol=${symbol}`),
};