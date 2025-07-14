import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { 
  Button, 
  Card, 
  CardContent, 
  CardHeader, 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell,
  Badge,
  Select,
  SkeletonTable
} from '@/components/ui';
import { formatDate, formatCurrency, ORDER_STATUS } from '@/utils';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { generalApi } from '@/api';
import { useTradingSymbols } from '@/hooks/useBalance';
import toast from 'react-hot-toast';

export const Orders: React.FC = () => {
  const [selectedSymbol, setSelectedSymbol] = useState<string>('');
  const [selectedOrders, setSelectedOrders] = useState<string[]>([]);
  const queryClient = useQueryClient();
  
  const { data: symbols } = useTradingSymbols();
  const { data: orders, isLoading } = useQuery({
    queryKey: ['orders', selectedSymbol],
    queryFn: () => generalApi.getOrders(selectedSymbol),
  });

  const cancelOrder = useMutation({
    mutationFn: generalApi.cancelOrder,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      toast.success('订单取消成功');
    },
  });

  const batchCancelOrders = useMutation({
    mutationFn: generalApi.batchCancelOrders,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      setSelectedOrders([]);
      toast.success('批量取消成功');
    },
  });

  const handleSelectOrder = (orderId: string) => {
    setSelectedOrders(prev => 
      prev.includes(orderId) 
        ? prev.filter(id => id !== orderId)
        : [...prev, orderId]
    );
  };

  const handleSelectAll = () => {
    if (selectedOrders.length === orders?.length) {
      setSelectedOrders([]);
    } else {
      setSelectedOrders(orders?.map(o => o.order_id) || []);
    }
  };

  const handleBatchCancel = () => {
    if (selectedOrders.length === 0) {
      toast.error('请选择要取消的订单');
      return;
    }
    
    if (window.confirm(`确定要取消选中的 ${selectedOrders.length} 个订单吗？`)) {
      batchCancelOrders.mutate(selectedOrders);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">订单管理</h1>
        </div>
        <SkeletonTable />
      </div>
    );
  }

  const activeOrders = orders?.filter(o => 
    ['new', 'partially_filled', 'pending'].includes(o.status)
  ) || [];

  return (
    <div className="space-y-6">
      {/* 页面标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">订单管理</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            查看和管理您的交易订单
          </p>
        </div>
        <div className="flex items-center space-x-3">
          <Select
            value={selectedSymbol}
            onChange={(e) => setSelectedSymbol(e.target.value)}
            options={[
              { value: '', label: '全部交易对' },
              ...(symbols?.map(s => ({ value: s.symbol, label: s.symbol })) || [])
            ]}
          />
          {selectedOrders.length > 0 && (
            <Button variant="danger" onClick={handleBatchCancel}>
              批量取消 ({selectedOrders.length})
            </Button>
          )}
        </div>
      </div>

      {/* 统计信息 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">总订单数</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
              {orders?.length || 0}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">活跃订单</p>
            <p className="text-2xl font-bold text-green-600 dark:text-green-400 mt-1">
              {activeOrders.length}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">已成交</p>
            <p className="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">
              {orders?.filter(o => o.status === 'filled').length || 0}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">已取消</p>
            <p className="text-2xl font-bold text-gray-600 dark:text-gray-400 mt-1">
              {orders?.filter(o => o.status === 'canceled').length || 0}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 订单列表 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">订单列表</h2>
            {orders && orders.length > 0 && (
              <label className="flex items-center space-x-2 text-sm">
                <input
                  type="checkbox"
                  checked={selectedOrders.length === orders.length}
                  onChange={handleSelectAll}
                  className="w-4 h-4 text-primary-500 border-gray-300 rounded focus:ring-primary-500"
                />
                <span className="text-gray-600 dark:text-gray-400">全选</span>
              </label>
            )}
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-10"></TableHead>
                <TableHead>订单ID</TableHead>
                <TableHead>交易对</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>方向</TableHead>
                <TableHead>数量</TableHead>
                <TableHead>价格</TableHead>
                <TableHead>已成交</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {orders?.map((order) => (
                <TableRow key={order.order_id}>
                  <TableCell>
                    <input
                      type="checkbox"
                      checked={selectedOrders.includes(order.order_id)}
                      onChange={() => handleSelectOrder(order.order_id)}
                      className="w-4 h-4 text-primary-500 border-gray-300 rounded focus:ring-primary-500"
                    />
                  </TableCell>
                  <TableCell className="font-mono text-xs">{order.order_id}</TableCell>
                  <TableCell>
                    <Badge variant="info">{order.symbol}</Badge>
                  </TableCell>
                  <TableCell>{order.type}</TableCell>
                  <TableCell>
                    <Badge variant={order.side === 'buy' ? 'success' : 'danger'}>
                      {order.side === 'buy' ? '买入' : '卖出'}
                    </Badge>
                  </TableCell>
                  <TableCell>{formatCurrency(order.quantity)}</TableCell>
                  <TableCell>{formatCurrency(order.price)}</TableCell>
                  <TableCell>{formatCurrency(order.executed_qty)}</TableCell>
                  <TableCell>
                    <Badge variant={ORDER_STATUS[order.status]?.color as any}>
                      {ORDER_STATUS[order.status]?.label}
                    </Badge>
                  </TableCell>
                  <TableCell>{formatDate(order.created_at, 'MM-DD HH:mm:ss')}</TableCell>
                  <TableCell>
                    {['new', 'partially_filled', 'pending'].includes(order.status) && (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          if (window.confirm('确定要取消这个订单吗？')) {
                            cancelOrder.mutate(order.order_id);
                          }
                        }}
                      >
                        取消
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          
          {(!orders || orders.length === 0) && (
            <div className="text-center py-12">
              <svg className="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
              <p className="text-gray-500 dark:text-gray-400">暂无订单</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};