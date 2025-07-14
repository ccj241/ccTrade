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
  SkeletonTable
} from '@/components/ui';
import { formatDate, formatCurrency, formatPercent } from '@/utils';
import { useFavoriteStore } from '@/stores';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dualInvestmentApi } from '@/api';
import toast from 'react-hot-toast';

const createStrategySchema = z.object({
  name: z.string().min(1, '请输入策略名称'),
  symbol: z.string().min(1, '请选择币种'),
  target_price: z.number().positive('目标价格必须大于0'),
  min_apy: z.number().min(0).max(100, 'APY必须在0-100之间'),
  max_investment: z.number().positive('最大投资额必须大于0'),
  single_investment: z.number().positive('单次投资额必须大于0'),
  is_call: z.boolean(),
  auto_compound: z.boolean(),
});

type CreateStrategyFormData = z.infer<typeof createStrategySchema>;

export const DualInvestment: React.FC = () => {
  const [page, setPage] = useState(1);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<any>(null);
  const queryClient = useQueryClient();
  
  // 获取收藏的币对 - 将所有hooks调用移到最前面
  const { favoritePairs } = useFavoriteStore();
  
  const { data: products } = useQuery({
    queryKey: ['dual-products'],
    queryFn: () => dualInvestmentApi.getProducts(),
  });

  const { data, isLoading } = useQuery({
    queryKey: ['dual-strategies', page],
    queryFn: () => dualInvestmentApi.getStrategies(page, 10),
  });

  const { data: stats } = useQuery({
    queryKey: ['dual-stats'],
    queryFn: () => dualInvestmentApi.getStats(),
  });

  const createStrategy = useMutation({
    mutationFn: dualInvestmentApi.createStrategy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dual-strategies'] });
      toast.success('双币投资策略创建成功');
      setIsCreateModalOpen(false);
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || '创建策略失败，请重试');
    },
  });

  const toggleStrategy = useMutation({
    mutationFn: (strategyId: number) => dualInvestmentApi.toggleStrategy(strategyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dual-strategies'] });
      toast.success('策略状态已切换');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || '切换策略状态失败');
    },
  });

  const deleteStrategy = useMutation({
    mutationFn: dualInvestmentApi.deleteStrategy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dual-strategies'] });
      toast.success('策略删除成功');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || '删除策略失败');
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateStrategyFormData>({
    defaultValues: {
      is_call: true,
      auto_compound: false,
      min_apy: 5,
    },
  });

  const onSubmit = async (formData: CreateStrategyFormData) => {
    try {
      await createStrategy.mutateAsync(formData);
      reset();
    } catch (error) {
      // 错误已在mutation中处理
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">双币投资</h1>
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
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">双币投资</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            自动化双币投资策略管理
          </p>
        </div>
        <Button onClick={() => setIsCreateModalOpen(true)}>
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          创建策略
        </Button>
      </div>

      {/* 统计信息 */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">总投资额</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {formatCurrency(stats.total_investment || 0)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">总收益</p>
              <p className="text-2xl font-bold text-green-600 dark:text-green-400 mt-1">
                {formatCurrency(stats.total_profit || 0)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">平均APY</p>
              <p className="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">
                {formatPercent((stats.average_apy || 0) / 100)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <p className="text-sm text-gray-500 dark:text-gray-400">活跃策略</p>
              <p className="text-2xl font-bold text-purple-600 dark:text-purple-400 mt-1">
                {stats.active_strategies || 0}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 可用产品 */}
      {products && products.length > 0 && (
        <Card>
          <CardHeader>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">可用产品</h2>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {products.slice(0, 6).map((product) => (
                <motion.div
                  key={product.id}
                  whileHover={{ scale: 1.02 }}
                  className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg cursor-pointer hover:border-primary-500 transition-colors"
                  onClick={() => {
                    setSelectedProduct(product);
                    setIsCreateModalOpen(true);
                  }}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-gray-900 dark:text-white">
                      {product.symbol}
                    </span>
                    <Badge variant={product.is_call ? 'success' : 'danger'}>
                      {product.is_call ? '低买' : '高卖'}
                    </Badge>
                  </div>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-500">执行价</span>
                      <span className="text-gray-900 dark:text-white">
                        {formatCurrency(product.exercise_price)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">APY</span>
                      <span className="text-green-600 dark:text-green-400 font-medium">
                        {formatPercent(product.apy / 100)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">期限</span>
                      <span className="text-gray-900 dark:text-white">
                        {product.duration}天
                      </span>
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* 策略列表 */}
      <Card>
        <CardHeader>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">我的策略</h2>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>策略名称</TableHead>
                <TableHead>币种</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>目标价</TableHead>
                <TableHead>最小APY</TableHead>
                <TableHead>最大投资</TableHead>
                <TableHead>单次投资</TableHead>
                <TableHead>自动复投</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data?.data && data.data.length > 0 ? (
                data.data.map((strategy) => (
                <TableRow key={strategy.id}>
                  <TableCell className="font-medium">{strategy.name}</TableCell>
                  <TableCell>
                    <Badge variant="info">{strategy.symbol}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={strategy.is_call ? 'success' : 'danger'}>
                      {strategy.is_call ? '低买' : '高卖'}
                    </Badge>
                  </TableCell>
                  <TableCell>{formatCurrency(strategy.target_price)}</TableCell>
                  <TableCell>{formatPercent(strategy.min_apy / 100)}</TableCell>
                  <TableCell>{formatCurrency(strategy.max_investment)}</TableCell>
                  <TableCell>{formatCurrency(strategy.single_investment)}</TableCell>
                  <TableCell>
                    <Badge variant={strategy.auto_compound ? 'success' : 'default'}>
                      {strategy.auto_compound ? '是' : '否'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={strategy.is_active ? 'success' : 'default'}>
                      {strategy.is_active ? '运行中' : '已停止'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end space-x-2">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => toggleStrategy.mutate(strategy.id)}
                      >
                        {strategy.is_active ? '停止' : '启动'}
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          if (window.confirm('确定要删除这个策略吗？')) {
                            deleteStrategy.mutate(strategy.id);
                          }
                        }}
                      >
                        删除
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
              ) : (
                <TableRow>
                  <TableCell colSpan={10} className="text-center py-8 text-gray-500">
                    暂无策略数据
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* 创建策略弹窗 */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          setSelectedProduct(null);
        }}
        title="创建双币投资策略"
        size="lg"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            {...register('name')}
            label="策略名称"
            placeholder="请输入策略名称"
            error={errors.name?.message}
          />

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                币种
              </label>
              <select
                {...register('symbol')}
                className="w-full px-3 py-2 border rounded-lg bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600"
                defaultValue={selectedProduct?.symbol}
              >
                <option value="">请选择币种</option>
                {favoritePairs.map(pair => (
                  <option key={pair.symbol} value={pair.symbol}>{pair.symbol}</option>
                ))}
              </select>
              {errors.symbol && (
                <p className="mt-1 text-sm text-red-500">{errors.symbol.message}</p>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                投资类型
              </label>
              <select
                {...register('is_call', { 
                  setValueAs: (value) => value === 'true' 
                })}
                className="w-full px-3 py-2 border rounded-lg bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600"
                defaultValue={selectedProduct?.is_call ? 'true' : 'false'}
              >
                <option value="true">低买</option>
                <option value="false">高卖</option>
              </select>
            </div>
          </div>

          <Input
            {...register('target_price', { valueAsNumber: true })}
            type="number"
            step="0.01"
            label="目标价格"
            placeholder="请输入目标价格"
            defaultValue={selectedProduct?.exercise_price}
            error={errors.target_price?.message}
          />

          <Input
            {...register('min_apy', { valueAsNumber: true })}
            type="number"
            step="0.1"
            min="0"
            max="100"
            label="最小APY (%)"
            placeholder="请输入最小APY"
            error={errors.min_apy?.message}
          />

          <div className="grid grid-cols-2 gap-4">
            <Input
              {...register('max_investment', { valueAsNumber: true })}
              type="number"
              step="0.01"
              label="最大投资额"
              placeholder="请输入最大投资额"
              error={errors.max_investment?.message}
            />

            <Input
              {...register('single_investment', { valueAsNumber: true })}
              type="number"
              step="0.01"
              label="单次投资额"
              placeholder="请输入单次投资额"
              error={errors.single_investment?.message}
            />
          </div>

          <div className="flex items-center">
            <input
              {...register('auto_compound')}
              type="checkbox"
              className="w-4 h-4 text-primary-500 border-gray-300 rounded focus:ring-primary-500"
            />
            <label className="ml-2 text-sm text-gray-700 dark:text-gray-300">
              自动复投
            </label>
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                setIsCreateModalOpen(false);
                setSelectedProduct(null);
              }}
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