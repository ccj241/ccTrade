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
  Modal,
  Input,
  Select,
  SkeletonTable
} from '@/components/ui';
import { formatDate, formatCurrency, WITHDRAW_STATUS, FREQUENCIES, PRESET_ASSETS, WITHDRAWAL_NETWORKS } from '@/utils';
import { useFavoriteStore } from '@/stores';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { withdrawalSchema } from '@/utils/validation';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { withdrawalApi } from '@/api';
import toast from 'react-hot-toast';

type WithdrawalFormData = z.infer<typeof withdrawalSchema>;

export const Withdrawals: React.FC = () => {
  const [page, setPage] = useState(1);
  const [historyPage, setHistoryPage] = useState(1);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<'rules' | 'history'>('rules');
  const queryClient = useQueryClient();
  
  // 获取收藏的币对
  const { favoritePairs } = useFavoriteStore();
  
  const { data, isLoading } = useQuery({
    queryKey: ['withdrawals', page],
    queryFn: () => withdrawalApi.getWithdrawals(page, 10),
  });

  const { data: historyData } = useQuery({
    queryKey: ['withdrawal-history', historyPage],
    queryFn: () => withdrawalApi.getWithdrawalHistory(historyPage, 20),
    enabled: activeTab === 'history',
  });

  const { data: stats } = useQuery({
    queryKey: ['withdrawal-stats'],
    queryFn: () => withdrawalApi.getWithdrawalStats(),
  });

  const createWithdrawal = useMutation({
    mutationFn: withdrawalApi.createWithdrawal,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['withdrawals'] });
      toast.success('提现规则创建成功');
      setIsCreateModalOpen(false);
    },
  });

  const toggleWithdrawal = useMutation({
    mutationFn: (withdrawalId: number) => withdrawalApi.toggleWithdrawal(withdrawalId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['withdrawals'] });
      toast.success('规则状态已切换');
    },
  });

  const deleteWithdrawal = useMutation({
    mutationFn: withdrawalApi.deleteWithdrawal,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['withdrawals'] });
      toast.success('规则删除成功');
    },
  });

  const syncHistory = useMutation({
    mutationFn: withdrawalApi.syncWithdrawalHistory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['withdrawal-history'] });
      toast.success('提现历史同步成功');
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<WithdrawalFormData>({
    resolver: zodResolver(withdrawalSchema),
  });

  const selectedAsset = watch('asset');
  const availableNetworks = selectedAsset && WITHDRAWAL_NETWORKS[selectedAsset as keyof typeof WITHDRAWAL_NETWORKS] 
    ? WITHDRAWAL_NETWORKS[selectedAsset as keyof typeof WITHDRAWAL_NETWORKS]
    : [];

  const onSubmit = async (formData: WithdrawalFormData) => {
    try {
      await createWithdrawal.mutateAsync(formData);
      reset();
      setIsCreateModalOpen(false);
    } catch (error) {
      // 错误已在mutation中处理
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">提现管理</h1>
        </div>
        <SkeletonTable />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 页面标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">提现管理</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            自动化提现规则和历史记录
          </p>
        </div>
        <div className="flex items-center space-x-3">
          {activeTab === 'history' && (
            <Button variant="secondary" onClick={() => syncHistory.mutate()}>
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              同步历史
            </Button>
          )}
          {activeTab === 'rules' && (
            <Button onClick={() => setIsCreateModalOpen(true)}>
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              创建规则
            </Button>
          )}
        </div>
      </div>

      {/* 统计信息 */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">总提现额</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {formatCurrency(stats.total_withdrawn || 0)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">本月提现</p>
              <p className="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">
                {formatCurrency(stats.monthly_withdrawn || 0)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">活跃规则</p>
              <p className="text-2xl font-bold text-green-600 dark:text-green-400 mt-1">
                {stats.active_rules || 0}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">总提现次数</p>
              <p className="text-2xl font-bold text-purple-600 dark:text-purple-400 mt-1">
                {stats.total_count || 0}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 标签页 */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('rules')}
            className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'rules'
                ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            提现规则
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'history'
                ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            提现历史
          </button>
        </nav>
      </div>

      {/* 提现规则 */}
      {activeTab === 'rules' && (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>规则名称</TableHead>
                  <TableHead>币种</TableHead>
                  <TableHead>金额</TableHead>
                  <TableHead>地址</TableHead>
                  <TableHead>网络</TableHead>
                  <TableHead>频率</TableHead>
                  <TableHead>下次执行</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data?.map((withdrawal) => (
                  <TableRow key={withdrawal.id}>
                    <TableCell className="font-medium">{withdrawal.rule_name}</TableCell>
                    <TableCell>
                      <Badge variant="info">{withdrawal.asset}</Badge>
                    </TableCell>
                    <TableCell>{formatCurrency(withdrawal.amount)}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {withdrawal.address.slice(0, 6)}...{withdrawal.address.slice(-4)}
                    </TableCell>
                    <TableCell>{withdrawal.network}</TableCell>
                    <TableCell>
                      {FREQUENCIES.find(f => f.value === withdrawal.frequency)?.label}
                    </TableCell>
                    <TableCell>
                      {withdrawal.next_execute_time 
                        ? formatDate(withdrawal.next_execute_time, 'MM-DD HH:mm')
                        : '-'
                      }
                    </TableCell>
                    <TableCell>
                      <Badge variant={withdrawal.is_active ? 'success' : 'default'}>
                        {withdrawal.is_active ? '启用' : '停用'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center justify-end space-x-2">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => toggleWithdrawal.mutate(withdrawal.id)}
                        >
                          {withdrawal.is_active ? '停用' : '启用'}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => {
                            if (window.confirm('确定要删除这个提现规则吗？')) {
                              deleteWithdrawal.mutate(withdrawal.id);
                            }
                          }}
                        >
                          删除
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* 提现历史 */}
      {activeTab === 'history' && (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>时间</TableHead>
                  <TableHead>币种</TableHead>
                  <TableHead>金额</TableHead>
                  <TableHead>地址</TableHead>
                  <TableHead>网络</TableHead>
                  <TableHead>交易ID</TableHead>
                  <TableHead>状态</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {historyData?.data?.map((history) => (
                  <TableRow key={history.id}>
                    <TableCell>{formatDate(history.created_at, 'MM-DD HH:mm')}</TableCell>
                    <TableCell>
                      <Badge variant="info">{history.asset}</Badge>
                    </TableCell>
                    <TableCell>{formatCurrency(history.amount)}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {history.address.slice(0, 6)}...{history.address.slice(-4)}
                    </TableCell>
                    <TableCell>{history.network}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {history.tx_id ? (
                        <span className="text-blue-600 dark:text-blue-400">
                          {history.tx_id.slice(0, 8)}...
                        </span>
                      ) : '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant={WITHDRAW_STATUS[history.status]?.color as any}>
                        {WITHDRAW_STATUS[history.status]?.label}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* 创建提现规则弹窗 */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="创建提现规则"
        size="lg"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            {...register('rule_name')}
            label="规则名称"
            placeholder="请输入规则名称"
            error={errors.rule_name?.message}
          />

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                币种
              </label>
              <select
                {...register('asset')}
                className="w-full px-3 py-2 border rounded-lg bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600"
              >
                <option value="">请选择币种</option>
                {/* 添加USDT作为默认选项 */}
                <option value="USDT">USDT</option>
                {favoritePairs.map(pair => {
                  // 从币对中提取基础币种（去掉USDT后缀）
                  const baseAsset = pair.symbol.replace('USDT', '');
                  // 避免重复添加USDT
                  if (baseAsset !== '' && baseAsset !== pair.symbol) {
                    return (
                      <option key={pair.symbol} value={baseAsset}>
                        {baseAsset}
                      </option>
                    );
                  }
                  return null;
                }).filter(Boolean)}
              </select>
              {errors.asset && (
                <p className="mt-1 text-sm text-red-500">{errors.asset.message}</p>
              )}
            </div>

            <Input
              {...register('amount', { valueAsNumber: true })}
              type="number"
              step="0.00000001"
              label="提现金额"
              placeholder="请输入提现金额"
              error={errors.amount?.message}
            />
          </div>

          <Input
            {...register('address')}
            label="提现地址"
            placeholder="请输入提现地址"
            error={errors.address?.message}
          />

          <div className="grid grid-cols-2 gap-4">
            <Select
              {...register('network')}
              label="提现网络"
              placeholder={selectedAsset ? "请选择网络" : "请先选择币种"}
              options={availableNetworks.map(net => ({ 
                value: net.value, 
                label: `${net.label} (手续费: ${net.fee})` 
              }))}
              disabled={!selectedAsset}
              error={errors.network?.message}
            />

            <Select
              {...register('frequency')}
              label="执行频率"
              placeholder="请选择执行频率"
              options={FREQUENCIES}
              error={errors.frequency?.message}
            />
          </div>

          <Input
            {...register('memo')}
            label="备注信息（可选）"
            placeholder="如需要可填写备注"
            error={errors.memo?.message}
          />

          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setIsCreateModalOpen(false)}
            >
              取消
            </Button>
            <Button type="submit" loading={createWithdrawal.isPending}>
              创建
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
};