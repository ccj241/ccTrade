import React, { useState, useMemo, useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Plus, ChevronLeft, ChevronRight, Trash2 } from 'lucide-react';
import { toast } from 'react-hot-toast';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { Modal } from '../components/ui/Modal';
import { Badge } from '../components/ui/Badge';
import { 
  useStrategies, 
  useCreateStrategy, 
  useToggleStrategy, 
  useDeleteStrategy 
} from '../hooks/useStrategies';
import { Strategy } from '../types';
import { useFavoriteStore } from '../stores/favoriteStore';

// 策略类型选项
const SPOT_STRATEGY_TYPES = [
  { value: 'simple', label: '简单策略' },
  { value: 'iceberg', label: '冰山策略' },
  { value: 'slow_iceberg', label: '慢冰山策略' },
];

// 基础验证模式
const baseStrategySchema = z.object({
  name: z.string().optional(),
  symbol: z.string().min(1, '请选择交易对'),
  type: z.enum(['simple', 'iceberg', 'slow_iceberg']),
  side: z.enum(['buy', 'sell']),
  quantity: z.number().positive('数量必须大于0'),
  auto_restart: z.boolean(),
  trigger_price: z.number().positive('触发价格必须大于0').optional(),
  price_float: z.number().optional(),
  timeout: z.number().optional(),
  layers: z.number().optional(),
  price_float_step: z.number().optional(),
  layer_quantities: z.array(z.number()).optional(),
  layer_price_floats: z.array(z.number()).optional(),
});

// 简单策略验证
const simpleStrategySchema = baseStrategySchema.extend({
  trigger_price: z.number().positive('触发价格必须大于0'),
  price_float: z.number().min(0).max(1000, '价格浮动最大千分之1').optional(),
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

// 动态验证函数
const getStrategySchema = (type: string) => {
  switch (type) {
    case 'simple':
      return simpleStrategySchema;
    case 'iceberg':
    case 'slow_iceberg':
      return icebergStrategySchema;
    default:
      return baseStrategySchema;
  }
};

type StrategyFormData = z.infer<typeof baseStrategySchema>;

export function StrategiesNew() {
  const [page, setPage] = useState(1);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [strategyToDelete, setStrategyToDelete] = useState<Strategy | null>(null);
  const [strategyType, setStrategyType] = useState<string>('simple');

  const { data, isLoading } = useStrategies(page, 10);
  const { favoritePairs } = useFavoriteStore();
  const createStrategy = useCreateStrategy();
  const toggleStrategy = useToggleStrategy();
  const deleteStrategy = useDeleteStrategy();

  // 初始化表单
  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    control,
    formState: { errors },
    clearErrors,
  } = useForm<StrategyFormData>({
    resolver: zodResolver(getStrategySchema(strategyType)),
    defaultValues: {
      type: 'simple',
      side: 'buy',
      symbol: '',
      quantity: 0,
      trigger_price: 0,
      auto_restart: false,
      price_float: 0,
      timeout: 5,
      layers: 10,
      price_float_step: 8,
      layer_quantities: [],
      layer_price_floats: [],
    },
    mode: 'onChange',
  });

  // 监听表单值
  const watchedType = watch('type');
  const selectedSymbol = watch('symbol');
  const side = watch('side');
  const triggerPrice = watch('trigger_price');
  const priceFloat = watch('price_float');
  const watchedLayers = watch('layers');
  const currentLayers = watchedLayers || 10;

  // 根据交易对提取基础货币和报价货币
  const { baseCurrency, quoteCurrency } = useMemo(() => {
    if (!selectedSymbol) return { baseCurrency: '', quoteCurrency: '' };
    
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

  // 计算冰山策略的层级分布
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
    
    // 价格浮动：每层递增
    const priceFloatStep = 8; // 默认万分之8
    for (let i = 0; i < layerCount; i++) {
      priceFloats.push(i * priceFloatStep);
    }
    
    return { quantities, priceFloats };
  };

  // 生成策略名称
  const generateStrategyName = () => {
    if (!selectedSymbol || !watchedType || !side) return '';
    
    const typeLabel = SPOT_STRATEGY_TYPES.find(t => t.value === watchedType)?.label || watchedType;
    const sideLabel = side === 'buy' ? '买入' : '卖出';
    const price = triggerPrice || '';
    
    return `${sideLabel}${selectedSymbol}${typeLabel}${price ? '@' + price : ''}`;
  };

  // 计算预估价格（简单策略）
  const estimatedPrice = useMemo(() => {
    if (!triggerPrice || !priceFloat || watchedType !== 'simple') return null;
    
    const floatRatio = priceFloat / 10000;
    if (side === 'buy') {
      return triggerPrice * (1 - floatRatio);
    } else {
      return triggerPrice * (1 + floatRatio);
    }
  }, [triggerPrice, priceFloat, side, watchedType]);

  // 当策略类型改变时，更新验证模式并重置相关字段
  useEffect(() => {
    clearErrors();
    
    if (watchedType === 'iceberg' || watchedType === 'slow_iceberg') {
      setValue('layers', 10);
      setValue('timeout', 5);
      setValue('price_float_step', 8);
      const { quantities, priceFloats } = calculateLayerDistribution(10);
      setValue('layer_quantities', quantities);
      setValue('layer_price_floats', priceFloats);
    } else if (watchedType === 'simple') {
      setValue('price_float', 0);
      setValue('timeout', 5);
    }
  }, [watchedType, setValue, clearErrors]);

  // 提交表单
  const onSubmit = async (formData: StrategyFormData) => {
    try {
      const strategyName = formData.name || generateStrategyName();
      
      const submitData: any = {
        name: strategyName,
        symbol: formData.symbol,
        type: formData.type,
        side: formData.side,
        quantity: formData.quantity,
        trigger_price: formData.trigger_price,
        auto_restart: formData.auto_restart,
      };

      // 根据策略类型添加配置
      if (formData.type === 'simple') {
        submitData.config = {
          price_float: formData.price_float || 0,
          timeout: formData.timeout || 5,
        };
      } else if (formData.type === 'iceberg' || formData.type === 'slow_iceberg') {
        const layerCount = formData.layers || 10;
        
        let quantities = formData.layer_quantities;
        let priceFloats = formData.layer_price_floats;
        
        if (!quantities || quantities.length === 0) {
          const distribution = calculateLayerDistribution(layerCount);
          quantities = distribution.quantities;
          priceFloats = distribution.priceFloats;
        }
        
        submitData.config = {
          layers: layerCount,
          timeout: formData.timeout || 5,
          price_float_step: formData.price_float_step || 8,
          layer_quantities: quantities,
          layer_price_floats: priceFloats,
        };
      }
      
      await createStrategy.mutateAsync(submitData);
      
      // 重置表单
      reset({
        type: 'simple',
        side: 'buy',
        symbol: '',
        quantity: 0,
        trigger_price: 0,
        auto_restart: false,
        price_float: 0,
        timeout: 5,
        layers: 10,
        price_float_step: 8,
        layer_quantities: [],
        layer_price_floats: [],
      });
      
      setIsCreateModalOpen(false);
      setStrategyType('simple');
      toast.success('策略创建成功');
    } catch (error: any) {
      console.error('创建策略失败:', error);
      toast.error(error.message || '策略创建失败');
    }
  };

  // 切换策略状态
  const handleToggle = async (strategyId: number) => {
    try {
      await toggleStrategy.mutateAsync(strategyId);
      toast.success('策略状态已更新');
    } catch (error) {
      toast.error('操作失败');
    }
  };

  // 删除策略
  const handleDelete = async () => {
    if (!strategyToDelete) return;
    
    try {
      await deleteStrategy.mutateAsync(strategyToDelete.id);
      toast.success('策略已删除');
      setDeleteModalOpen(false);
      setStrategyToDelete(null);
    } catch (error) {
      toast.error('删除失败');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">现货策略</h1>
        <Button onClick={() => setIsCreateModalOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          创建策略
        </Button>
      </div>

      {/* 策略列表 */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  策略名称
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  类型
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  交易对
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  方向
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  数量
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  触发价格
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  状态
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {isLoading ? (
                <tr>
                  <td colSpan={8} className="px-6 py-4 text-center text-gray-500">
                    加载中...
                  </td>
                </tr>
              ) : data?.data?.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-6 py-4 text-center text-gray-500">
                    暂无策略
                  </td>
                </tr>
              ) : (
                data?.data?.map((strategy) => (
                  <tr key={strategy.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                      {strategy.name}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {SPOT_STRATEGY_TYPES.find(t => t.value === strategy.type)?.label || strategy.type}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {strategy.symbol}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {strategy.side === 'buy' ? '买入' : '卖出'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {strategy.quantity}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {strategy.trigger_price}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <Badge variant={strategy.is_active ? 'success' : 'default'} size="sm">
                        {strategy.is_active ? '运行中' : '已停止'}
                      </Badge>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <div className="flex space-x-2">
                        <Button
                          size="sm"
                          variant={strategy.is_active ? 'secondary' : 'primary'}
                          onClick={() => handleToggle(strategy.id)}
                        >
                          {strategy.is_active ? '暂停' : '启动'}
                        </Button>
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() => {
                            setStrategyToDelete(strategy);
                            setDeleteModalOpen(true);
                          }}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* 分页 */}
        {data && data.total > 10 && (
          <div className="px-6 py-4 flex items-center justify-between border-t border-gray-200 dark:border-gray-700">
            <div className="flex items-center space-x-2">
              <Button
                size="sm"
                variant="secondary"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <span className="text-sm text-gray-700 dark:text-gray-300">
                第 {page} 页，共 {Math.ceil(data.total / 10)} 页
              </span>
              <Button
                size="sm"
                variant="secondary"
                onClick={() => setPage(p => p + 1)}
                disabled={page >= Math.ceil(data.total / 10)}
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* 创建策略弹窗 */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          reset();
          setStrategyType('simple');
        }}
        title="创建现货策略"
        size="lg"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {/* 策略类型 */}
          <Controller
            name="type"
            control={control}
            render={({ field }) => (
              <Select
                label="策略类型"
                value={field.value}
                options={SPOT_STRATEGY_TYPES}
                onChange={(value) => {
                  field.onChange(value);
                  setStrategyType(value);
                }}
                error={errors.type?.message}
              />
            )}
          />

          {/* 基础信息 */}
          <div className="grid grid-cols-2 gap-4">
            <Controller
              name="side"
              control={control}
              render={({ field }) => (
                <Select
                  label="交易方向"
                  value={field.value}
                  options={[
                    { value: 'buy', label: '买入' },
                    { value: 'sell', label: '卖出' }
                  ]}
                  onChange={field.onChange}
                  error={errors.side?.message}
                />
              )}
            />

            <Controller
              name="symbol"
              control={control}
              render={({ field }) => (
                <Select
                  label="交易对"
                  placeholder="请选择交易对"
                  value={field.value}
                  options={favoritePairs.map(pair => ({
                    value: pair.symbol,
                    label: pair.symbol
                  }))}
                  onChange={field.onChange}
                  error={errors.symbol?.message}
                />
              )}
            />
          </div>

          {/* 触发价格（所有策略都需要） */}
          <Input
            {...register('trigger_price', { valueAsNumber: true })}
            type="number"
            step="0.00000001"
            label={`触发价格（${quoteCurrency || '报价货币'}）`}
            placeholder="达到此价格后触发"
            error={errors.trigger_price?.message}
          />

          {/* 数量和策略名称 */}
          <div className="grid grid-cols-2 gap-4">
            <Input
              {...register('quantity', { valueAsNumber: true })}
              type="number"
              step="0.00000001"
              label={`数量（${baseCurrency || '基础货币'}）`}
              placeholder="请输入数量"
              error={errors.quantity?.message}
            />

            <Input
              {...register('name')}
              label="策略名称（可选）"
              placeholder={generateStrategyName() || '自动生成'}
              error={errors.name?.message}
            />
          </div>

          {/* 简单策略特有字段 */}
          {watchedType === 'simple' && (
            <div className="space-y-4 border-t pt-4">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                简单策略参数
              </h3>
              
              <div className="grid grid-cols-2 gap-4">
                <Input
                  {...register('price_float', { valueAsNumber: true })}
                  type="number"
                  step="1"
                  min="0"
                  max="1000"
                  label="价格浮动（万分比）"
                  placeholder="默认0"
                  error={errors.price_float?.message}
                />
                
                <Input
                  {...register('timeout', { valueAsNumber: true })}
                  type="number"
                  step="1"
                  min="1"
                  label="超时时间（分钟）"
                  placeholder="默认5分钟"
                  error={errors.timeout?.message}
                />
              </div>

              {estimatedPrice && (
                <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
                  <p className="text-sm text-blue-700 dark:text-blue-300">
                    预估成交价格：{estimatedPrice.toFixed(8)} {quoteCurrency}
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
                <Controller
                  name="layers"
                  control={control}
                  render={({ field }) => (
                    <Select
                      label="层数"
                      placeholder="请选择层数"
                      value={field.value?.toString()}
                      options={[5, 6, 7, 8, 9, 10].map(n => ({ 
                        value: n.toString(), 
                        label: `${n}层` 
                      }))}
                      onChange={(e) => {
                        const numValue = parseInt(e.target.value);
                        field.onChange(numValue);
                        const { quantities, priceFloats } = calculateLayerDistribution(numValue);
                        setValue('layer_quantities', quantities);
                        setValue('layer_price_floats', priceFloats);
                      }}
                      error={errors.layers?.message}
                    />
                  )}
                />

                <Input
                  {...register('timeout', { valueAsNumber: true })}
                  type="number"
                  step="1"
                  min="1"
                  label="超时时间（分钟）"
                  placeholder="默认5分钟"
                  error={errors.timeout?.message}
                />
              </div>

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

              <div className="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg text-xs">
                <p className="font-medium mb-2">策略说明：</p>
                <p>
                  {watchedType === 'iceberg' 
                    ? `一次性将${currentLayers}层单子全部挂上，${watch('timeout') || 5}分钟后未完全成交则撤销，继续监测触发价格`
                    : `一次只挂一层单子，${watch('timeout') || 5}分钟后未成交则撤销并挂下一层，直到所有层完成`}
                </p>
                <p className="mt-2">
                  {side === 'buy' 
                    ? `买入时：第1层按买1价挂单，后续每层向下浮动万分之${watch('price_float_step') || 8}`
                    : `卖出时：第1层按卖1价挂单，后续每层向上浮动万分之${watch('price_float_step') || 8}`}
                </p>
              </div>

              {/* 策略预览 */}
              {watch('quantity') > 0 && triggerPrice > 0 && (
                <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg text-xs mt-3">
                  <p className="font-medium mb-2 text-blue-800 dark:text-blue-200">策略预览：</p>
                  <div className="space-y-1">
                    {(() => {
                      const quantity = watch('quantity') || 0;
                      const layers = watch('layers') || 10;
                      const priceFloatStep = (watch('price_float_step') || 8) / 10000;
                      const { quantities } = calculateLayerDistribution(layers);
                      
                      return quantities.map((ratio, index) => {
                        const layerQuantity = (quantity * ratio).toFixed(8);
                        let layerPrice = triggerPrice;
                        
                        if (index > 0) {
                          if (side === 'buy') {
                            layerPrice = triggerPrice * (1 - priceFloatStep * index);
                          } else {
                            layerPrice = triggerPrice * (1 + priceFloatStep * index);
                          }
                        }
                        
                        return (
                          <div key={index} className="flex justify-between text-gray-700 dark:text-gray-300">
                            <span>第{index + 1}层：</span>
                            <span>{layerQuantity} {baseCurrency} @ {layerPrice.toFixed(8)} {quoteCurrency}</span>
                          </div>
                        );
                      });
                    })()}
                  </div>
                  <div className="mt-2 pt-2 border-t border-blue-200 dark:border-blue-700">
                    <div className="flex justify-between font-medium text-blue-800 dark:text-blue-200">
                      <span>总计：</span>
                      <span>{watch('quantity')} {baseCurrency}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* 自动重启 */}
          <div className="flex items-center">
            <input
              {...register('auto_restart')}
              type="checkbox"
              className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
            />
            <label className="ml-2 text-sm text-gray-700 dark:text-gray-300">
              完成后自动重启
            </label>
          </div>

          {/* 提交按钮 */}
          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                setIsCreateModalOpen(false);
                reset();
                setStrategyType('simple');
              }}
            >
              取消
            </Button>
            <Button type="submit" loading={createStrategy.isLoading}>
              创建策略
            </Button>
          </div>
        </form>
      </Modal>

      {/* 删除确认弹窗 */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => {
          setDeleteModalOpen(false);
          setStrategyToDelete(null);
        }}
        title="确认删除"
      >
        <div className="space-y-4">
          <p className="text-gray-600 dark:text-gray-400">
            确定要删除策略 "{strategyToDelete?.name}" 吗？此操作不可恢复。
          </p>
          <div className="flex justify-end space-x-3">
            <Button
              variant="secondary"
              onClick={() => {
                setDeleteModalOpen(false);
                setStrategyToDelete(null);
              }}
            >
              取消
            </Button>
            <Button
              variant="danger"
              onClick={handleDelete}
              loading={deleteStrategy.isLoading}
            >
              确认删除
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}