import React, { useState, useEffect } from 'react';
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
import { formatDate, formatCurrency, FUTURES_STRATEGY_TYPES, STRATEGY_TYPES, MARGIN_TYPES } from '@/utils';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { futuresStrategySchema } from '@/utils/validation';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { futuresApi } from '@/api';
import { useFavoriteStore } from '@/stores/favoriteStore';
import toast from 'react-hot-toast';

type FuturesStrategyFormData = z.infer<typeof futuresStrategySchema>;

export const Futures: React.FC = () => {
  const [page, setPage] = useState(1);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [layerFloatRates, setLayerFloatRates] = useState<Record<number, number>>({});
  const queryClient = useQueryClient();
  
  const { data, isLoading } = useQuery({
    queryKey: ['futures-strategies', page],
    queryFn: () => futuresApi.getFuturesStrategies(page, 10),
  });

  const favoritePairs = useFavoriteStore((state) => state.favoritePairs);
  const { data: positions } = useQuery({
    queryKey: ['futures-positions'],
    queryFn: () => futuresApi.getPositions(),
  });

  const createStrategy = useMutation({
    mutationFn: futuresApi.createFuturesStrategy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['futures-strategies'] });
      toast.success('期货策略创建成功');
      setIsCreateModalOpen(false);
    },
  });

  const toggleStrategy = useMutation({
    mutationFn: (strategyId: number) => futuresApi.toggleFuturesStrategy(strategyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['futures-strategies'] });
      toast.success('策略状态已切换');
    },
  });

  const deleteStrategy = useMutation({
    mutationFn: futuresApi.deleteFuturesStrategy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['futures-strategies'] });
      toast.success('策略删除成功');
    },
  });

  const syncPositions = useMutation({
    mutationFn: futuresApi.syncPositions,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['futures-positions'] });
      toast.success('持仓同步成功');
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<FuturesStrategyFormData>({
    resolver: zodResolver(futuresStrategySchema),
    defaultValues: {
      leverage: 8,
      margin_type: 'cross',
      float_basis_points: 0.1,
      take_profit_bp: 0,
      stop_loss_bp: 0,
    },
  });

  // 监听表单变化以计算预估值
  const watchMarginAmount = watch('margin_amount');
  const watchType = watch('type');
  const watchLeverage = watch('leverage');
  const watchPrice = watch('price');
  const watchSide = watch('side');
  const watchTakeProfitBP = watch('take_profit_bp');
  const watchStopLossBP = watch('stop_loss_bp');
  const watchFloatBP = watch('float_basis_points');
  const watchLayers = watch('config.layers');

  // 监听层数变化，清空自定义浮动比例
  useEffect(() => {
    setLayerFloatRates({});
  }, [watchLayers]);

  // 计算预估值
  const calculateEstimates = () => {
    if (!watchMarginAmount || !watchLeverage || !watchPrice) return null;

    const marginAmount = Number(watchMarginAmount);
    const leverage = Number(watchLeverage);
    const price = Number(watchPrice);
    const takeProfitBP = Number(watchTakeProfitBP) || 0;
    const stopLossBP = Number(watchStopLossBP) || 0;
    const floatBP = Number(watchFloatBP) || 0;

    // 计算实际下单数量
    const orderValue = marginAmount * leverage;
    const orderQuantity = orderValue / price;

    // 计算浮动后的实际下单价格
    let actualOrderPrice = price;
    if (watchSide === 'buy') {
      // 做多时，万分比向下浮动
      actualOrderPrice = price * (1 - floatBP / 10000);
    } else {
      // 做空时，万分比向上浮动
      actualOrderPrice = price * (1 + floatBP / 10000);
    }

    // 预计盈利（扣除手续费，假设手续费率为0.04%）
    const feeRate = 0.0004;
    const profitRate = takeProfitBP / 10000;
    const estimatedProfit = orderValue * profitRate - orderValue * feeRate * 2;

    // 预计亏损（包含手续费）
    const lossRate = stopLossBP / 10000;
    const estimatedLoss = orderValue * lossRate + orderValue * feeRate * 2;

    // 计算爆仓价格（逐仓模式）
    let liquidationPrice = 0;
    if (watchSide === 'buy') {
      // 做多爆仓价格 = 开仓价格 * (1 - 1/杠杆 + 手续费率)
      liquidationPrice = actualOrderPrice * (1 - 1/leverage + feeRate);
    } else {
      // 做空爆仓价格 = 开仓价格 * (1 + 1/杠杆 - 手续费率)
      liquidationPrice = actualOrderPrice * (1 + 1/leverage - feeRate);
    }

    return {
      orderQuantity,
      actualOrderPrice,
      estimatedProfit,
      estimatedLoss,
      liquidationPrice,
      orderValue
    };
  };

  const estimates = calculateEstimates();

  // 计算冰山策略层级详情
  const calculateLayerDetails = () => {
    if (!watchType || (watchType !== 'iceberg' && watchType !== 'slow_iceberg')) return null;
    if (!watchMarginAmount || !watchLeverage || !watchPrice) return null;

    const marginAmount = Number(watchMarginAmount);
    const leverage = Number(watchLeverage);
    const basePrice = Number(watchPrice);
    const floatBP = Number(watchFloatBP) || 0.1;
    const layers = Number(watch('config.layers')) || 10;

    const totalOrderValue = marginAmount * leverage;
    const totalQuantity = totalOrderValue / basePrice;

    // 获取数量分配比例（从大到小）
    const getQuantityRatios = (layerCount: number) => {
      if (layerCount === 10) {
        return [0.19, 0.17, 0.15, 0.13, 0.11, 0.09, 0.07, 0.05, 0.03, 0.01];
      }
      // 其他层数使用等差数列
      const diff = 2.0 / (layerCount * (layerCount - 1) / 2);
      const ratios = [];
      for (let i = 0; i < layerCount; i++) {
        ratios.push((layerCount - i) * diff);
      }
      return ratios;
    };

    const quantityRatios = getQuantityRatios(layers);
    const layerDetails = [];

    for (let i = 0; i < layers; i++) {
      const layerQuantity = totalQuantity * quantityRatios[i];
      // 首层使用用户设置的浮动，其他层使用自定义值或默认每层增加万分之8
      const defaultFloatRate = i === 0 ? floatBP : i * 8;
      const layerFloatRate = layerFloatRates[i + 1] !== undefined ? layerFloatRates[i + 1] : defaultFloatRate;
      
      let layerPrice = basePrice;
      if (layerFloatRate > 0) {
        const floatRate = layerFloatRate / 10000.0;
        if (watchSide === 'buy') {
          layerPrice = basePrice * (1 - floatRate);
        } else {
          layerPrice = basePrice * (1 + floatRate);
        }
      }

      layerDetails.push({
        layer: i + 1,
        quantity: layerQuantity,
        floatRate: layerFloatRate,
        price: layerPrice,
        value: layerQuantity * layerPrice
      });
    }

    return layerDetails;
  };

  const layerDetails = calculateLayerDetails();

  // 监听表单变化自动生成策略名称
  useEffect(() => {
    const generateStrategyName = () => {
      const side = watchSide === 'buy' ? '做多' : '做空';
      const symbol = watch('symbol');
      const type = watchType;
      const price = watchPrice;
      
      if (symbol && type && price) {
        const strategyTypeLabel = FUTURES_STRATEGY_TYPES.find(t => t.value === type)?.label || type;
        const autoName = `${side}${symbol}${strategyTypeLabel}${price}`;
        
        // 只有在策略名称为空或者是之前自动生成的名称时才更新
        const currentName = watch('name');
        if (!currentName || currentName.match(/^(做多|做空).+/)) {
          setValue('name', autoName, { shouldValidate: false });
        }
      }
    };
    
    generateStrategyName();
  }, [watchSide, watch('symbol'), watchType, watchPrice, setValue, watch]);

  const onSubmit = async (formData: FuturesStrategyFormData) => {
    try {
      // 如果是冰山策略，确保config字段正确设置
      const submitData = { ...formData };
      if (formData.type === 'iceberg' || formData.type === 'slow_iceberg') {
        const layers = formData.config?.layers || 10;
        
        // 构建层级浮动比例数组
        const layer_price_floats = [];
        for (let i = 0; i < layers; i++) {
          const layerNum = i + 1;
          if (layerFloatRates[layerNum] !== undefined) {
            layer_price_floats.push(layerFloatRates[layerNum]);
          } else {
            // 使用默认值：首层使用用户设置的浮动，其他层每层增加万分之8
            layer_price_floats.push(i === 0 ? (formData.float_basis_points || 0.1) : i * 8);
          }
        }
        
        submitData.config = {
          layers: layers,
          timeout_minutes: formData.config?.timeout_minutes || 5,
          layer_price_floats: layer_price_floats,
          first_layer_float: formData.float_basis_points || 0.1,
        };
      }
      await createStrategy.mutateAsync(submitData);
      reset();
      setLayerFloatRates({}); // 重置层级浮动比例
    } catch (error) {
      // 错误已在mutation中处理
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">期货策略</h1>
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
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">期货策略</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            管理您的期货交易策略和持仓
          </p>
        </div>
        <div className="flex items-center space-x-3">
          <Button variant="secondary" onClick={() => syncPositions.mutate()}>
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            同步持仓
          </Button>
          <Button onClick={() => setIsCreateModalOpen(true)}>
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            创建策略
          </Button>
        </div>
      </div>

      {/* 持仓信息 */}
      {positions && positions.length > 0 && (
        <Card>
          <CardHeader>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">当前持仓</h2>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {positions.map((position) => (
                <div key={`${position.symbol}-${position.position_side}`} className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-gray-900 dark:text-white">{position.symbol}</span>
                    <Badge variant={position.position_side === 'long' ? 'success' : 'danger'}>
                      {position.position_side === 'long' ? '做多' : '做空'}
                    </Badge>
                  </div>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-500">持仓数量</span>
                      <span className="text-gray-900 dark:text-white">{formatCurrency(position.position_amt)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">开仓均价</span>
                      <span className="text-gray-900 dark:text-white">{formatCurrency(position.entry_price)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">未实现盈亏</span>
                      <span className={position.unrealized_profit >= 0 ? 'text-green-500' : 'text-red-500'}>
                        {position.unrealized_profit >= 0 ? '+' : ''}{formatCurrency(position.unrealized_profit)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">杠杆</span>
                      <span className="text-gray-900 dark:text-white">{position.leverage}x</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

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
                <TableHead>保证金</TableHead>
                <TableHead>杠杆</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data?.data?.map((strategy) => (
                <TableRow key={strategy.id}>
                  <TableCell className="font-medium">{strategy.name}</TableCell>
                  <TableCell>
                    <Badge variant="info">{strategy.symbol}</Badge>
                  </TableCell>
                  <TableCell>
                    {STRATEGY_TYPES.find(t => t.value === strategy.type)?.label || strategy.type}
                  </TableCell>
                  <TableCell>
                    <Badge variant={strategy.side === 'buy' ? 'success' : 'danger'}>
                      {strategy.side === 'buy' ? '做多' : '做空'}
                    </Badge>
                  </TableCell>
                  <TableCell>{formatCurrency(strategy.margin_amount)} USDT</TableCell>
                  <TableCell>{strategy.leverage}x</TableCell>
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
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* 创建策略弹窗 */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="创建期货策略"
        size="md"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
          <Input
            {...register('name')}
            label="策略名称"
            placeholder="留空将自动生成（交易方向+交易对+策略类型+触发价格）"
            error={errors.name?.message}
          />

          <div className="grid grid-cols-2 gap-3">
            <Select
              {...register('symbol')}
              label="交易对"
              placeholder="请选择交易对"
              options={favoritePairs.map(pair => ({ value: pair.symbol, label: pair.symbol }))}
              error={errors.symbol?.message}
            />

            <Select
              {...register('type')}
              label="策略类型"
              placeholder="请选择策略类型"
              options={FUTURES_STRATEGY_TYPES}
              error={errors.type?.message}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <Select
              {...register('side')}
              label="交易方向"
              placeholder="请选择交易方向"
              options={[
                { value: 'buy', label: '做多' },
                { value: 'sell', label: '做空' },
              ]}
              error={errors.side?.message}
            />

            <Select
              {...register('leverage', { valueAsNumber: true })}
              label="杠杆倍数"
              placeholder="请选择杠杆倍数"
              options={Array.from({ length: 20 }, (_, i) => ({ 
                value: i + 1, 
                label: `${i + 1}倍` 
              }))}
              error={errors.leverage?.message}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <Input
              {...register('margin_amount', { valueAsNumber: true })}
              type="number"
              step="0.01"
              label="保证金本值 (USDT)"
              placeholder="请输入保证金金额"
              error={errors.margin_amount?.message}
            />

            <Input
              {...register('price', { valueAsNumber: true })}
              type="number"
              step="0.01"
              label="触发价格"
              placeholder="请输入触发价格"
              error={errors.price?.message}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <Input
              {...register('float_basis_points', { valueAsNumber: true })}
              type="number"
              step="0.1"
              min="0"
              label="首单万分比浮动"
              placeholder="请输入首单万分比浮动（默认0.1）"
              error={errors.float_basis_points?.message}
            />

            <Select
              {...register('margin_type')}
              label="保证金模式"
              options={MARGIN_TYPES}
              error={errors.margin_type?.message}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <Input
              {...register('take_profit_bp', { valueAsNumber: true })}
              type="number"
              step="1"
              min="0"
              label="止盈万分比"
              placeholder="请输入止盈万分比"
              error={errors.take_profit_bp?.message}
            />

            <Input
              {...register('stop_loss_bp', { valueAsNumber: true })}
              type="number"
              step="1"
              min="0"
              label="止损万分比"
              placeholder="请输入止损万分比"
              error={errors.stop_loss_bp?.message}
            />
          </div>

          {/* 冰山策略特有配置 */}
          {(watchType === 'iceberg' || watchType === 'slow_iceberg') && (
            <>
              <div className="grid grid-cols-2 gap-3">
                <Select
                  {...register('config.layers', { valueAsNumber: true })}
                  label="层数"
                  placeholder="请选择层数"
                  options={[
                    { value: 5, label: '5层' },
                    { value: 6, label: '6层' },
                    { value: 7, label: '7层' },
                    { value: 8, label: '8层' },
                    { value: 9, label: '9层' },
                    { value: 10, label: '10层' },
                  ]}
                  defaultValue={10}
                  error={errors.config?.layers?.message}
                />

                <Input
                  {...register('config.timeout_minutes', { valueAsNumber: true })}
                  type="number"
                  step="1"
                  min="1"
                  label="超时时间(分钟)"
                  placeholder="请输入超时时间"
                  defaultValue={5}
                  error={errors.config?.timeout_minutes?.message}
                />
              </div>

              <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
                <p className="text-sm text-blue-800 dark:text-blue-200">
                  <strong>冰山策略说明：</strong>
                </p>
                <ul className="mt-2 text-xs text-blue-700 dark:text-blue-300 space-y-1">
                  {watchType === 'iceberg' && (
                    <>
                      <li>• 快速冰山：一次性将所有层的单子全部挂上</li>
                      <li>• 价格将根据各层的浮动比例自动计算</li>
                      <li>• 超时后将自动撤销未成交的订单</li>
                    </>
                  )}
                  {watchType === 'slow_iceberg' && (
                    <>
                      <li>• 慢速冰山：每次只挂一层单子</li>
                      <li>• 当前层成交后或超时后才会挂下一层</li>
                      <li>• 更加隐蔽的交易方式</li>
                    </>
                  )}
                  <li>• 默认价格浮动：每层递增万分之八</li>
                  <li>• 默认数量分配：根据层数自动计算</li>
                </ul>
              </div>

              {/* 层级详情选项卡 */}
              {layerDetails && (
                <div className="mt-4">
                  <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
                    <div className="px-4 py-3 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
                      <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">层级详情</h4>
                    </div>
                    <div className="max-h-60 overflow-y-auto">
                      <table className="w-full text-xs">
                        <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0">
                          <tr>
                            <th className="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-400">层级</th>
                            <th className="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-400">数量</th>
                            <th className="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-400">
                              <div className="flex items-center">
                                <span>浮动比例</span>
                                <div className="group relative ml-1">
                                  <svg className="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                  </svg>
                                  <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                                    万分比，可修改
                                  </div>
                                </div>
                              </div>
                            </th>
                            <th className="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-400">预计挂单价格</th>
                            <th className="px-3 py-2 text-left font-medium text-gray-600 dark:text-gray-400">价值(USDT)</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                          {layerDetails.map((detail, index) => (
                            <tr key={detail.layer} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                              <td className="px-3 py-2 text-gray-900 dark:text-white">第{detail.layer}层</td>
                              <td className="px-3 py-2 text-gray-900 dark:text-white">{detail.quantity.toFixed(4)}</td>
                              <td className="px-3 py-2">
                                <input
                                  type="number"
                                  step="0.1"
                                  className="w-20 px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded focus:ring-1 focus:ring-primary-500 dark:bg-gray-800 dark:text-white"
                                  value={detail.floatRate}
                                  onChange={(e) => {
                                    const newRate = parseFloat(e.target.value) || 0;
                                    setLayerFloatRates(prev => ({
                                      ...prev,
                                      [detail.layer]: newRate
                                    }));
                                  }}
                                />
                              </td>
                              <td className="px-3 py-2 text-gray-900 dark:text-white">{formatCurrency(detail.price)}</td>
                              <td className="px-3 py-2 text-gray-900 dark:text-white">{formatCurrency(detail.value)}</td>
                            </tr>
                          ))}
                        </tbody>
                        <tfoot className="bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
                          <tr>
                            <td colSpan={2} className="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">总计</td>
                            <td className="px-3 py-2"></td>
                            <td className="px-3 py-2"></td>
                            <td className="px-3 py-2 font-medium text-gray-900 dark:text-white">
                              {formatCurrency(layerDetails.reduce((sum, d) => sum + d.value, 0))}
                            </td>
                          </tr>
                        </tfoot>
                      </table>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}

          {/* 显示预估计算结果 */}
          {estimates && (
            <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg space-y-2">
              <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">预估计算</h4>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-gray-500">开仓价值：</span>
                  <span className="ml-2 text-gray-900 dark:text-white font-medium">
                    {formatCurrency(estimates.orderValue)} USDT
                  </span>
                </div>
                <div>
                  <span className="text-gray-500">开仓数量：</span>
                  <span className="ml-2 text-gray-900 dark:text-white font-medium">
                    {estimates.orderQuantity.toFixed(4)}
                  </span>
                </div>
                <div>
                  <span className="text-gray-500">实际开仓价格：</span>
                  <span className="ml-2 text-gray-900 dark:text-white font-medium">
                    {formatCurrency(estimates.actualOrderPrice)}
                  </span>
                </div>
                <div>
                  <span className="text-gray-500">预计爆仓价格：</span>
                  <span className="ml-2 text-orange-500 font-medium">
                    {formatCurrency(estimates.liquidationPrice)}
                  </span>
                </div>
                {watchTakeProfitBP > 0 && (
                  <div>
                    <span className="text-gray-500">预计盈利：</span>
                    <span className="ml-2 text-green-500 font-medium">
                      +{formatCurrency(estimates.estimatedProfit)} USDT
                    </span>
                  </div>
                )}
                {watchStopLossBP > 0 && (
                  <div>
                    <span className="text-gray-500">预计亏损：</span>
                    <span className="ml-2 text-red-500 font-medium">
                      -{formatCurrency(estimates.estimatedLoss)} USDT
                    </span>
                  </div>
                )}
              </div>
              {watchSide && (
                <div className="mt-2 text-xs text-gray-500">
                  提示：{watchSide === 'buy' ? '做多' : '做空'}时，万分比浮动将{watchSide === 'buy' ? '向下' : '向上'}调整价格
                </div>
              )}
            </div>
          )}

          <div className="flex items-center">
            <input
              {...register('auto_restart')}
              type="checkbox"
              className="w-4 h-4 text-primary-500 border-gray-300 rounded focus:ring-primary-500"
            />
            <label className="ml-2 text-sm text-gray-700 dark:text-gray-300">
              自动重启
            </label>
          </div>

          <div className="flex justify-end space-x-3 pt-3">
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