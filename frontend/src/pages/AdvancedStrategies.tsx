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
import { useStrategies, useCreateStrategy, useToggleStrategy, useDeleteStrategy } from '@/hooks';
import { formatDate, formatCurrency, ADVANCED_STRATEGY_TYPES, STRATEGY_TYPES, ORDER_TYPES } from '@/utils';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { strategySchema } from '@/utils/validation';
import { z } from 'zod';
import { useTradingSymbols } from '@/hooks/useBalance';
import toast from 'react-hot-toast';

type StrategyFormData = z.infer<typeof strategySchema>;

export const AdvancedStrategies: React.FC = () => {
  const [page, setPage] = useState(1);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [selectedStrategy, setSelectedStrategy] = useState<any>(null);
  
  const { data, isLoading } = useStrategies(page, 10);
  const { data: symbols } = useTradingSymbols();
  const createStrategy = useCreateStrategy();
  const toggleStrategy = useToggleStrategy();
  const deleteStrategy = useDeleteStrategy();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<StrategyFormData>({
    resolver: zodResolver(strategySchema),
  });

  // 过滤出高级策略
  const advancedStrategies = data?.data?.filter(strategy => 
    ADVANCED_STRATEGY_TYPES.some(type => type.value === strategy.type)
  ) || [];

  const onSubmit = async (formData: StrategyFormData) => {
    try {
      await createStrategy.mutateAsync(formData);
      reset();
      setIsCreateModalOpen(false);
    } catch (error) {
      // 错误已在hook中处理
    }
  };

  const handleToggle = async (strategyId: number) => {
    try {
      await toggleStrategy.mutateAsync(strategyId);
    } catch (error) {
      // 错误已在hook中处理
    }
  };

  const handleDelete = async (strategyId: number) => {
    if (window.confirm('确定要删除这个策略吗？')) {
      try {
        await deleteStrategy.mutateAsync(strategyId);
      } catch (error) {
        // 错误已在hook中处理
      }
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">高级策略</h1>
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
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">高级策略</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            管理您的高级交易策略
          </p>
        </div>
        {/* 暂时隐藏创建按钮，等待策略开发完成 */}
        {ADVANCED_STRATEGY_TYPES.length > 0 && (
          <Button onClick={() => setIsCreateModalOpen(true)}>
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            创建高级策略
          </Button>
        )}
      </div>

      {/* 策略提示 */}
      <Card glass>
        <CardContent className="p-6">
          <div className="flex items-center space-x-3">
            <svg className="w-6 h-6 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
              <p className="text-gray-900 dark:text-white font-medium">高级策略功能正在开发中</p>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                即将推出更多自定义策略，敬请期待！
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 策略列表 */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>策略名称</TableHead>
                <TableHead>交易对</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>方向</TableHead>
                <TableHead>配置</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {advancedStrategies.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-gray-500 dark:text-gray-400">
                    暂无高级策略
                  </TableCell>
                </TableRow>
              ) : (
                advancedStrategies.map((strategy) => (
                  <TableRow key={strategy.id}>
                    <TableCell className="font-medium">{strategy.name}</TableCell>
                    <TableCell>
                      <Badge variant="info">{strategy.symbol}</Badge>
                    </TableCell>
                    <TableCell>
                      {ADVANCED_STRATEGY_TYPES.find(t => t.value === strategy.type)?.label || strategy.type}
                    </TableCell>
                    <TableCell>
                      <Badge variant={strategy.side === 'buy' ? 'success' : 'danger'}>
                        {strategy.side === 'buy' ? '买入' : '卖出'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setSelectedStrategy(strategy);
                          toast.success('策略配置: ' + JSON.stringify(strategy.config, null, 2));
                        }}
                      >
                        查看配置
                      </Button>
                    </TableCell>
                    <TableCell>
                      <Badge variant={strategy.is_active ? 'success' : 'default'}>
                        {strategy.is_active ? '运行中' : '已停止'}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatDate(strategy.created_at, 'MM-DD HH:mm')}</TableCell>
                    <TableCell>
                      <div className="flex items-center justify-end space-x-2">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleToggle(strategy.id)}
                        >
                          {strategy.is_active ? '停止' : '启动'}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleDelete(strategy.id)}
                        >
                          删除
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* 创建策略弹窗 */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="创建高级策略"
        size="md"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            {...register('name')}
            label="策略名称"
            placeholder="请输入策略名称"
            error={errors.name?.message}
          />

          <Select
            {...register('symbol')}
            label="交易对"
            placeholder="请选择交易对"
            options={symbols?.map(s => ({ value: s.symbol, label: s.symbol })) || []}
            error={errors.symbol?.message}
          />

          <Select
            {...register('type')}
            label="策略类型"
            placeholder="请选择策略类型"
            options={ADVANCED_STRATEGY_TYPES}
            error={errors.type?.message}
          />

          <Select
            {...register('side')}
            label="交易方向"
            placeholder="请选择交易方向"
            options={[
              { value: 'buy', label: '买入' },
              { value: 'sell', label: '卖出' },
            ]}
            error={errors.side?.message}
          />

          <Input
            {...register('quantity', { valueAsNumber: true })}
            type="number"
            step="0.00000001"
            label="数量"
            placeholder="请输入数量"
            error={errors.quantity?.message}
          />

          <Input
            {...register('price', { valueAsNumber: true })}
            type="number"
            step="0.00000001"
            label="价格"
            placeholder="请输入价格"
            error={errors.price?.message}
          />

          <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              高级策略需要额外的配置参数，创建后请在策略详情中进行配置。
            </p>
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setIsCreateModalOpen(false)}
            >
              取消
            </Button>
            <Button type="submit" loading={createStrategy.isPending}>
              创建
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
};