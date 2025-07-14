import React, { useState, useMemo, useEffect } from 'react';
import { 
  Button, 
  Card, 
  CardContent, 
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
import { formatDate, formatCurrency, formatNumber, SPOT_STRATEGY_TYPES } from '@/utils';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useFavoriteStore } from '@/stores';
import toast from 'react-hot-toast';

// 基础策略字段
const baseStrategySchema = z.object({
  name: z.string().optional(),
  symbol: z.string().min(1, '请选择交易对'),
  type: z.string().min(1, '请选择策略类型'),
  side: z.enum(['buy', 'sell']),
  quantity: z.number().positive('数量必须大于0'),
  auto_restart: z.boolean().optional(),
});

// 简单策略验证
const simpleStrategySchema = baseStrategySchema.extend({
  trigger_price: z.number().positive('触发价格必须大于0'),
  price_float: z.number().min(0).max(10000, '价格浮动最大万分之10000').optional(),
  timeout: z.number().min(1, '超时时间至少1分钟').optional(),
});

// 冰山策略验证
const icebergStrategySchema = baseStrategySchema.extend({
  trigger_price: z.number().positive('触发价格必须大于0'),
  layers: z.number().min(5).max(10, '层数必须在5-10之间'),
  price_float_step: z.number().min(0).max(1000, '价格浮动步长最大千分之1').optional(),
  timeout: z.number().min(1, '超时时间至少1分钟').optional(),
  layer_quantities: z.array(z.number()).optional(),
  layer_price_floats: z.array(z.number()).optional(),
});

// 慢冰山策略验证（与冰山策略相同）
const slowIcebergStrategySchema = icebergStrategySchema;

// 动态验证函数
const getStrategySchema = (type: string) => {
  switch (type) {
    case 'simple':
      return simpleStrategySchema;
    case 'iceberg':
      return icebergStrategySchema;
    case 'slow_iceberg':
      return slowIcebergStrategySchema;
    default:
      return baseStrategySchema;
  }
};

// 定义更具体的类型（暂时未使用）
// type SimpleStrategyData = z.infer<typeof simpleStrategySchema>;
// type IcebergStrategyData = z.infer<typeof icebergStrategySchema>;
// type SlowIcebergStrategyData = z.infer<typeof slowIcebergStrategySchema>;

// 使用交叉类型来包含所有可能的字段
type StrategyFormData = {
  name?: string;
  symbol: string;
  type: string;
  side: 'buy' | 'sell';
  quantity: number;
  auto_restart?: boolean;
  trigger_price?: number;
  price_float?: number;
  timeout?: number;
  layers?: number;
  price_float_step?: number;
  layer_quantities?: number[];
  layer_price_floats?: number[];
};

export const Strategies: React.FC = () => {
  const [page, setPage] = useState(1);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  // const [selectedStrategy, setSelectedStrategy] = useState<any>(null);
  const [strategyType, setStrategyType] = useState('simple');
  const [editingStrategy, setEditingStrategy] = useState<any>(null);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  
  const { data, isLoading } = useStrategies(page, 10);
  const { favoritePairs } = useFavoriteStore();
  const createStrategy = useCreateStrategy();
  const toggleStrategy = useToggleStrategy();
  const deleteStrategy = useDeleteStrategy();

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    getValues,
    trigger,
    control,
    formState: { errors },
    clearErrors,
  } = useForm<StrategyFormData>({
    resolver: zodResolver(getStrategySchema(strategyType)),
    defaultValues: {
      type: 'simple',
      side: 'buy',
      auto_restart: false,
      price_float: 0,
      timeout: 5,
      layers: 10,
      price_float_step: 8,
      layer_quantities: [],
      layer_price_floats: [],
    },
    mode: 'onChange', // 实时验证
    mode: 'onChange'
  });

  const watchedValues = watch();
  const selectedSymbol = watch('symbol');
  const side = watch('side');
  const priceFloat = watch('price_float');
  const triggerPrice = watch('trigger_price');
  const watchedType = watch('type');
  const watchedLayers = watch('layers');
  const watchedLayerQuantities = watch('layer_quantities');
  const watchedLayerPriceFloats = watch('layer_price_floats');
  
  // 使用表单中的layers值，确保有默认值
  const currentLayers = typeof watchedLayers === 'number' ? watchedLayers : 10;

  // 根据交易对提取基础货币和报价货币
  const { baseCurrency, quoteCurrency } = useMemo(() => {
    if (!selectedSymbol) return { baseCurrency: '', quoteCurrency: '' };
    // 假设交易对格式为 BTCUSDT, ETHUSDT 等
    if (selectedSymbol.endsWith('USDT')) {
      return {
        baseCurrency: selectedSymbol.slice(0, -4),
        quoteCurrency: 'USDT'
      };
    } else if (selectedSymbol.endsWith('BUSD')) {
      return {
        baseCurrency: selectedSymbol.slice(0, -4),
        quoteCurrency: 'BUSD'
      };
    } else if (selectedSymbol.endsWith('BTC')) {
      return {
        baseCurrency: selectedSymbol.slice(0, -3),
        quoteCurrency: 'BTC'
      };
    } else if (selectedSymbol.endsWith('ETH')) {
      return {
        baseCurrency: selectedSymbol.slice(0, -3),
        quoteCurrency: 'ETH'
      };
    }
    return { baseCurrency: selectedSymbol, quoteCurrency: '' };
  }, [selectedSymbol]);

  // 生成策略名称
  const generateStrategyName = () => {
    if (!selectedSymbol || !strategyType || !side) return '';
    
    const typeLabel = SPOT_STRATEGY_TYPES.find(t => t.value === strategyType)?.label || strategyType;
    const sideLabel = side === 'buy' ? '买入' : '卖出';
    const triggerPrice = watchedValues.trigger_price || '';
    
    return `${sideLabel}${selectedSymbol}${typeLabel}${triggerPrice ? '@' + triggerPrice : ''}`;
  };

  // 计算预估价格
  const estimatedPrice = useMemo(() => {
    if (!triggerPrice || !priceFloat || strategyType !== 'simple') return null;
    
    const floatRatio = priceFloat / 10000; // 万分比转换为比例
    if (side === 'buy') {
      // 买入时向下浮动
      return triggerPrice * (1 - floatRatio);
    } else {
      // 卖出时向上浮动
      return triggerPrice * (1 + floatRatio);
    }
  }, [triggerPrice, priceFloat, side, strategyType]);

  // 计算冰山策略的默认层级分布
  const calculateLayerDistribution = (layerCount: number) => {
    const quantities: number[] = [];
    const priceFloats: number[] = [];
    
    if (layerCount === 10) {
      quantities.push(0.19, 0.17, 0.15, 0.13, 0.11, 0.09, 0.07, 0.05, 0.03, 0.01);
    } else {
      // 等差数列计算
      const totalSteps = (layerCount * (layerCount + 1)) / 2;
      for (let i = layerCount; i >= 1; i--) {
        quantities.push(i / totalSteps);
      }
    }
    
    // 价格浮动：每层递增万分之8
    for (let i = 0; i < layerCount; i++) {
      priceFloats.push(i * 8);
    }
    
    return { quantities, priceFloats };
  };

  // 确保层级分布与层数同步
  useEffect(() => {
    if ((watchedType === 'iceberg' || watchedType === 'slow_iceberg') && watchedLayers) {
      const currentQuantities = getValues('layer_quantities');
      const currentPriceFloats = getValues('layer_price_floats');
      
      // 如果层数改变了或者没有分布数据，更新分布
      if (!currentQuantities?.length || 
          !currentPriceFloats?.length || 
          currentQuantities.length !== watchedLayers ||
          currentPriceFloats.length !== watchedLayers) {
        const { quantities, priceFloats } = calculateLayerDistribution(watchedLayers);
        setValue('layer_quantities', quantities);
        setValue('layer_price_floats', priceFloats);
      }
    }
  }, [watchedType, watchedLayers, setValue, getValues]);

  const onSubmit = async (formData: StrategyFormData) => {
    try {
      // 如果没有设置名称，使用自动生成的名称
      const strategyName = formData.name || generateStrategyName();
      
      // 构建提交数据
      const submitData: any = {
        name: strategyName,
        symbol: formData.symbol,
        type: formData.type,
        side: formData.side,
        quantity: formData.quantity,
        auto_restart: formData.auto_restart,
      };

      // 根据策略类型添加特定字段
      if (formData.type === 'simple') {
        submitData.trigger_price = formData.trigger_price;
        submitData.config = {
          price_float: formData.price_float || 0,
          timeout: formData.timeout || 5,
        };
      } else if (formData.type === 'iceberg' || formData.type === 'slow_iceberg') {
        // 确保layers是数字类型
        const layerCount = typeof formData.layers === 'string' ? parseInt(formData.layers) : formData.layers || 10;
        
        // 如果没有层级分布数据，重新计算
        let quantities = formData.layer_quantities;
        let priceFloats = formData.layer_price_floats;
        
        if (!quantities || !priceFloats || quantities.length === 0 || priceFloats.length === 0) {
          const distribution = calculateLayerDistribution(layerCount);
          quantities = distribution.quantities;
          priceFloats = distribution.priceFloats;
        }
        
        submitData.trigger_price = formData.trigger_price;
        submitData.config = {
          layers: Number(layerCount), // 确保是数字类型
          timeout: Number(formData.timeout || 5), // 确保是数字类型
          price_float_step: Number(formData.price_float_step || 8), // 确保是数字类型
          layer_quantities: quantities,
          layer_price_floats: priceFloats,
        };
      }
      
      await createStrategy.mutateAsync(submitData);
      // 完整重置表单
      const resetValues: StrategyFormData = {
        type: 'simple',
        side: 'buy',
        symbol: '',
        quantity: 0,
        auto_restart: false,
        price_float: 0,
        timeout: 5,
        layers: 10,
        price_float_step: 8,
        layer_quantities: [],
        layer_price_floats: [],
      };
      reset(resetValues);
      setIsCreateModalOpen(false);
      setStrategyType('simple');
      toast.success('策略创建成功');
    } catch (error) {
      toast.error('策略创建失败');
    }
  };

  const handleToggle = async (strategyId: number) => {
    try {
      await toggleStrategy.mutateAsync(strategyId);
      toast.success('策略状态已更新');
    } catch (error) {
      toast.error('操作失败');
    }
  };

  const handleDelete = async (strategyId: number) => {
    if (window.confirm('确定要删除这个策略吗？')) {
      try {
        await deleteStrategy.mutateAsync(strategyId);
        toast.success('策略已删除');
      } catch (error) {
        toast.error('删除失败');
      }
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">现货策略</h1>
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
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">现货策略</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            管理您的现货交易策略
          </p>
        </div>
        <Button onClick={() => {
          // 重置表单到初始状态
          const defaultValues: StrategyFormData = {
            type: 'simple',
            side: 'buy',
            symbol: '',
            quantity: 0,
            auto_restart: false,
            price_float: 0,
            timeout: 5,
            layers: 10,
            price_float_step: 8,
            layer_quantities: [],
            layer_price_floats: [],
          };
          reset(defaultValues);
          setStrategyType('simple');
          setIsCreateModalOpen(true);
        }}>
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          创建策略
        </Button>
      </div>

      {/* 策略列表 */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>策略名称</TableHead>
                <TableHead>交易对</TableHead>
                <TableHead>策略类型</TableHead>
                <TableHead>方向</TableHead>
                <TableHead>总数量</TableHead>
                <TableHead>触发价格</TableHead>
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
                    {SPOT_STRATEGY_TYPES.find(t => t.value === strategy.type)?.label || strategy.type}
                  </TableCell>
                  <TableCell>
                    <Badge variant={strategy.side === 'buy' ? 'success' : 'danger'}>
                      {strategy.side === 'buy' ? '买入' : '卖出'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {formatCurrency(strategy.quantity)}
                    {strategy.symbol && (
                      <span className="text-gray-500 dark:text-gray-400 text-sm ml-1">
                        ({strategy.symbol.replace(/USDT$|BUSD$|BTC$|ETH$/, (match) => {
                          const baseAsset = strategy.symbol.slice(0, -match.length);
                          return baseAsset;
                        })})
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    {formatNumber(strategy.trigger_price || strategy.price, 4)}
                    {strategy.symbol && (
                      <span className="text-gray-500 dark:text-gray-400 text-sm ml-1">
                        ({strategy.symbol.replace(/.*?(USDT|BUSD|BTC|ETH)$/, '$1')})
                      </span>
                    )}
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
                        onClick={() => {
                          setEditingStrategy(strategy);
                          setIsEditModalOpen(true);
                        }}
                      >
                        编辑
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
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* 分页 */}
      {data && data.total > 10 && (
        <div className="flex items-center justify-center space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setPage(page - 1)}
            disabled={page === 1}
          >
            上一页
          </Button>
          <span className="text-sm text-gray-600 dark:text-gray-400">
            第 {page} 页，共 {Math.ceil(data.total / 10)} 页
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setPage(page + 1)}
            disabled={page === Math.ceil(data.total / 10)}
          >
            下一页
          </Button>
        </div>
      )}

      {/* 创建策略弹窗 */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          const resetValues: StrategyFormData = {
            type: 'simple',
            side: 'buy',
            symbol: '',
            quantity: 0,
            auto_restart: false,
            price_float: 0,
            timeout: 5,
            layers: 10,
            price_float_step: 8,
            layer_quantities: [],
            layer_price_floats: [],
          };
          reset(resetValues);
          setStrategyType('simple');
        }}
        title="创建策略"
        size="lg"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {/* 基础信息 */}
          <div className="space-y-4">
            <Input
              {...register('name')}
              label="策略名称（可选）"
              placeholder={generateStrategyName() || "自动生成"}
              error={errors.name?.message}
            />

            <div className="grid grid-cols-2 gap-4">
              <Controller
                name="symbol"
                control={control}
                render={({ field }) => (
                  <Select
                    {...field}
                    label="交易对"
                    placeholder="请选择交易对"
                    options={favoritePairs.map(p => ({ 
                      value: p.symbol, 
                      label: p.symbol 
                    }))}
                    error={errors.symbol?.message}
                  />
                )}
              />

              <Controller
                name="type"
                control={control}
                defaultValue="simple"
                render={({ field }) => (
                  <Select
                    {...field}
                    label="策略类型"
                    placeholder="请选择策略类型"
                    value={field.value || 'simple'}
                    options={[...SPOT_STRATEGY_TYPES]}
                    onChange={(value: string) => {
                      field.onChange(value);
                      setStrategyType(value);
                      
                      // 清除之前的验证错误
                      clearErrors();
                      
                      // 重置部分表单数据以适应新的策略类型
                      if (value === 'iceberg' || value === 'slow_iceberg') {
                        const defaultLayers = 10;
                        setValue('layers', defaultLayers);
                        setValue('timeout', 5);
                        setValue('price_float_step', 8);
                        // 立即计算并设置层级分布
                        const { quantities, priceFloats } = calculateLayerDistribution(defaultLayers);
                        setValue('layer_quantities', quantities);
                        setValue('layer_price_floats', priceFloats);
                      } else if (value === 'simple') {
                        setValue('price_float', 0);
                        setValue('timeout', 5);
                        // 清除冰山策略相关的数据
                        setValue('layers', undefined);
                        setValue('layer_quantities', []);
                        setValue('layer_price_floats', []);
                      }
                    }}
                    error={errors.type?.message}
                  />
                )}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Controller
                name="side"
                control={control}
                render={({ field }) => (
                  <Select
                    {...field}
                    label="交易方向"
                    placeholder="请选择交易方向"
                    options={[
                      { value: 'buy', label: '买入' },
                      { value: 'sell', label: '卖出' },
                    ]}
                    error={errors.side?.message}
                  />
                )}
              />

              <Input
                {...register('quantity', { valueAsNumber: true })}
                type="number"
                step="0.00000001"
                label={`总数量（${baseCurrency || '基础货币'}）`}
                placeholder="请输入总数量"
                error={errors.quantity?.message}
              />
            </div>
          </div>

          {/* 简单策略特有字段 */}
          {watchedType === 'simple' && (
            <div className="space-y-4 border-t pt-4">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">简单策略参数</h3>
              
              <div className="grid grid-cols-2 gap-4">
                <Input
                  {...register('trigger_price', { valueAsNumber: true })}
                  type="number"
                  step="0.00000001"
                  label={`触发价格（${quoteCurrency || '报价货币'}）`}
                  placeholder="达到此价格后触发"
                  error={errors.trigger_price?.message}
                />

                <Input
                  {...register('price_float', { valueAsNumber: true })}
                  type="number"
                  step="1"
                  min="0"
                  max="10000"
                  label="挂单价格浮动（万分比）"
                  placeholder="默认0"
                  error={errors.price_float?.message}
                />
              </div>

              <Input
                {...register('timeout', { valueAsNumber: true })}
                type="number"
                step="1"
                min="1"
                label="超时时间（分钟）"
                placeholder="默认5分钟"
                error={errors.timeout?.message}
              />

              <p className="text-xs text-gray-500 dark:text-gray-400">
                {side === 'buy' 
                  ? '买入时：达到触发价格后，以买1价向下浮动指定万分比挂单'
                  : '卖出时：达到触发价格后，以卖1价向上浮动指定万分比挂单'}
              </p>
              
              {estimatedPrice && (
                <div className="mt-2 p-2 bg-gray-50 dark:bg-gray-800 rounded">
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    预估挂单价格：<span className="font-medium text-gray-900 dark:text-white">
                      {formatNumber(estimatedPrice, 4)} {quoteCurrency}
                    </span>
                  </p>
                </div>
              )}
            </div>
          )}

          {/* 冰山策略特有字段 */}
          {(watchedType === 'iceberg' || watchedType === 'slow_iceberg') && (
            <div className="space-y-4 border-t pt-4">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                {watchedType === 'iceberg' ? '冰山策略参数' : '慢冰山策略参数'}
              </h3>
              
              <div className="grid grid-cols-2 gap-4">
                <Input
                  {...register('trigger_price', { valueAsNumber: true })}
                  type="number"
                  step="0.00000001"
                  label={`触发价格（${quoteCurrency || '报价货币'}）`}
                  placeholder="达到此价格后触发"
                  error={errors.trigger_price?.message}
                />

                <Controller
                  name="layers"
                  control={control}
                  render={({ field }) => (
                    <Select
                      label="层数"
                      placeholder="请选择层数"
                      value={field.value?.toString() || "10"}
                      options={[5, 6, 7, 8, 9, 10].map(n => ({ 
                        value: n.toString(), 
                        label: `${n}层` 
                      }))}
                      onChange={(value) => {
                        const numValue = parseInt(value);
                        field.onChange(numValue);
                        // 立即计算并更新层级分布，确保预览能及时更新
                        const { quantities, priceFloats } = calculateLayerDistribution(numValue);
                        setValue('layer_quantities', quantities);
                        setValue('layer_price_floats', priceFloats);
                      }}
                      error={errors.layers?.message}
                    />
                  )}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <Input
                  {...register('timeout', { valueAsNumber: true })}
                  type="number"
                  step="1"
                  min="1"
                  label="超时时间（分钟）"
                  placeholder="默认5分钟"
                  error={errors.timeout?.message}
                />
                
                <Input
                  {...register('price_float_step', { valueAsNumber: true })}
                  type="number"
                  step="1"
                  min="0"
                  max="1000"
                  label="价格浮动步长（万分比）"
                  placeholder="默认8"
                  error={errors.price_float_step?.message}
                />
              </div>

              <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg text-xs">
                <p className="font-medium mb-2">策略说明：</p>
                {watchedType === 'iceberg' ? (
                  <p>一次性将{currentLayers}层单子全部挂上，{watchedValues.timeout || 5}分钟后未完全成交则撤销，继续监测触发价格</p>
                ) : (
                  <p>一次只挂一层单子，{watchedValues.timeout || 5}分钟后未成交则撤销并将本层未成交的部分继续按新的买、卖1价的本层百分比浮动挂，直到所有层完成</p>
                )}
                <p className="mt-2">
                  {side === 'buy' 
                    ? `买入时：第1层按买1价挂单，后续每层向下浮动万分之${watchedValues.price_float_step || 8}`
                    : `卖出时：第1层按卖1价挂单，后续每层向上浮动万分之${watchedValues.price_float_step || 8}`}
                </p>
              </div>
              
              <div className="mt-4">
                <p className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">层级分布预览</p>
                <div className="space-y-2">
                  {(watchedType === 'iceberg' || watchedType === 'slow_iceberg') && 
                   watchedLayerQuantities && watchedLayerQuantities.length > 0 && 
                   watchedLayerPriceFloats && watchedLayerPriceFloats.length > 0 ? (
                    watchedLayerQuantities.map((layerQuantity, i) => {
                      const layerPriceFloat = watchedLayerPriceFloats[i] || 0;
                      return (
                        <div key={i} className="flex items-center justify-between text-sm">
                          <span className="text-gray-600 dark:text-gray-400">第{i + 1}层</span>
                          <span className="text-gray-900 dark:text-white">
                            数量: {layerQuantity > 0 ? (layerQuantity * 100).toFixed(0) : '0'}%
                          </span>
                          <span className="text-gray-900 dark:text-white">
                            浮动: 万分之{layerPriceFloat}
                          </span>
                        </div>
                      );
                    })
                  ) : (
                    <div className="text-sm text-gray-500 dark:text-gray-400">
                      {(watchedType === 'iceberg' || watchedType === 'slow_iceberg') 
                        ? '请设置层数以查看层级分布' 
                        : '请选择冰山策略查看层级分布'}
                    </div>
                  )}
                </div>
              </div>
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

          <div className="flex justify-end space-x-3 pt-4 border-t">
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                setIsCreateModalOpen(false);
                reset();
                setStrategyType('simple');
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
      
      {/* 编辑策略弹窗 */}
      <Modal
        isOpen={isEditModalOpen}
        onClose={() => {
          setIsEditModalOpen(false);
          setEditingStrategy(null);
        }}
        title="编辑策略"
        size="lg"
      >
        {editingStrategy && (
          <div className="space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              策略名称：{editingStrategy.name}
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              交易对：{editingStrategy.symbol}
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              类型：{SPOT_STRATEGY_TYPES.find(t => t.value === editingStrategy.type)?.label || editingStrategy.type}
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              方向：{editingStrategy.side === 'buy' ? '买入' : '卖出'}
            </p>
            
            <div className="mt-4 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded">
              <p className="text-sm text-yellow-800 dark:text-yellow-200">
                提示：目前暂不支持编辑策略参数，如需修改请删除后重新创建。
              </p>
            </div>
            
            <div className="flex justify-end space-x-3 pt-4 border-t">
              <Button
                variant="ghost"
                onClick={() => {
                  setIsEditModalOpen(false);
                  setEditingStrategy(null);
                }}
              >
                关闭
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};