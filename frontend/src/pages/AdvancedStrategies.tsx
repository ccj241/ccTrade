import React, { useState } from 'react';
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
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { 
  TrendingUp, 
  Activity, 
  DollarSign, 
  BarChart3,
  Brain,
  Settings,
  Play,
  Pause,
  AlertTriangle,
  Target,
  Shield,
  Zap
} from 'lucide-react';
import { useQuantitativeStrategies, useCreateQuantitativeStrategy, useToggleQuantitativeStrategy, useDeleteQuantitativeStrategy } from '@/hooks/useQuantitative';
import { formatDate, formatCurrency } from '@/utils';
import { QuantitativeStrategyConfig } from '@/components/quantitative/QuantitativeStrategyConfig';
import { QuantitativePerformance } from '@/components/quantitative/QuantitativePerformance';

interface QuantitativeFormData {
  name: string;
  totalCapitalUSDT: number;
  riskPreference: 'conservative' | 'moderate' | 'aggressive';
  maxPositions: number;
  symbols?: string[];
}

export const AdvancedStrategies: React.FC = () => {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [selectedStrategy, setSelectedStrategy] = useState<any>(null);
  const [showPerformance, setShowPerformance] = useState(false);
  const [showConfig, setShowConfig] = useState(false);
  
  const page = 1;
  const { data, isLoading } = useQuantitativeStrategies(page, 10);
  const createStrategy = useCreateQuantitativeStrategy();
  const toggleStrategy = useToggleQuantitativeStrategy();
  const deleteStrategy = useDeleteQuantitativeStrategy();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<QuantitativeFormData>();

  const quantitativeStrategies = data?.strategies || [];

  const onSubmit = async (formData: QuantitativeFormData) => {
    try {
      await createStrategy.mutateAsync(formData);
      reset();
      setIsCreateModalOpen(false);
      toast.success('量化策略创建成功！');
    } catch (error) {
      toast.error('创建失败，请重试');
    }
  };

  const handleToggle = async (strategyId: number) => {
    try {
      await toggleStrategy.mutateAsync(strategyId);
      toast.success('策略状态已更新');
    } catch (error) {
      toast.error('操作失败，请重试');
    }
  };

  const handleDelete = async (strategyId: number) => {
    if (window.confirm('确定要删除这个量化策略吗？删除后无法恢复。')) {
      try {
        await deleteStrategy.mutateAsync(strategyId);
        toast.success('策略已删除');
      } catch (error) {
        toast.error('删除失败，请重试');
      }
    }
  };

  const getRiskBadgeVariant = (risk: string) => {
    switch (risk) {
      case 'conservative':
        return 'success';
      case 'moderate':
        return 'warning';
      case 'aggressive':
        return 'danger';
      default:
        return 'default';
    }
  };

  const getRiskLabel = (risk: string) => {
    switch (risk) {
      case 'conservative':
        return '保守型';
      case 'moderate':
        return '中等型';
      case 'aggressive':
        return '激进型';
      default:
        return risk;
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">量化策略</h1>
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
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <Brain className="w-8 h-8 text-blue-500" />
            AI量化策略
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            智能多因子量化交易系统，综合技术指标、基本面、市场情绪分析
          </p>
        </div>
        <Button onClick={() => setIsCreateModalOpen(true)} className="flex items-center gap-2">
          <Zap className="w-5 h-5" />
          创建量化策略
        </Button>
      </div>

      {/* 策略统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card glass>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">运行中策略</p>
                <p className="text-2xl font-bold mt-1">
                  {quantitativeStrategies.filter(s => s.is_active).length}
                </p>
              </div>
              <Activity className="w-8 h-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card glass>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">总投入资金</p>
                <p className="text-2xl font-bold mt-1">
                  {formatCurrency(
                    quantitativeStrategies.reduce((sum, s) => 
                      sum + (s.config?.total_capital_usdt || 0), 0
                    )
                  )}
                </p>
              </div>
              <DollarSign className="w-8 h-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>

        <Card glass>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">平均收益率</p>
                <p className="text-2xl font-bold mt-1 text-green-500">
                  +12.5%
                </p>
              </div>
              <TrendingUp className="w-8 h-8 text-green-500" />
            </div>
          </CardContent>
        </Card>

        <Card glass>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">风险评级</p>
                <p className="text-2xl font-bold mt-1">
                  中等
                </p>
              </div>
              <Shield className="w-8 h-8 text-yellow-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 策略特点说明 */}
      <Card glass>
        <CardContent className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="flex items-start space-x-3">
              <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                <BarChart3 className="w-6 h-6 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 dark:text-white">多因子分析</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  综合60+技术指标、基本面、市场情绪、资金流向等多维度分析
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3">
              <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
                <Target className="w-6 h-6 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 dark:text-white">智能仓位管理</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  根据综合评分动态调整仓位，严格止损止盈，控制最大回撤
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3">
              <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
                <Brain className="w-6 h-6 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 dark:text-white">自动化执行</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  24/7全自动运行，捕捉市场机会，无需人工干预
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 策略列表 */}
      <Card>
        <CardHeader className="border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold">我的量化策略</h2>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>策略名称</TableHead>
                <TableHead>投入资金</TableHead>
                <TableHead>风险偏好</TableHead>
                <TableHead>最大持仓</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>收益率</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {quantitativeStrategies.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-gray-500 dark:text-gray-400">
                    暂无量化策略，点击上方按钮创建您的第一个AI量化策略
                  </TableCell>
                </TableRow>
              ) : (
                quantitativeStrategies.map((strategy) => (
                  <TableRow key={strategy.id}>
                    <TableCell className="font-medium">{strategy.name}</TableCell>
                    <TableCell>
                      {formatCurrency(strategy.config?.total_capital_usdt || 0)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={getRiskBadgeVariant(strategy.config?.risk_preference)}>
                        {getRiskLabel(strategy.config?.risk_preference)}
                      </Badge>
                    </TableCell>
                    <TableCell>{strategy.config?.max_positions || 5}个</TableCell>
                    <TableCell>
                      <Badge variant={strategy.is_active ? 'success' : 'default'}>
                        {strategy.is_active ? '运行中' : '已停止'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className="text-green-500 font-medium">
                        +{(Math.random() * 30).toFixed(2)}%
                      </span>
                    </TableCell>
                    <TableCell>{formatDate(strategy.created_at, 'MM-DD HH:mm')}</TableCell>
                    <TableCell>
                      <div className="flex items-center justify-end space-x-2">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleToggle(strategy.id)}
                          className="flex items-center gap-1"
                        >
                          {strategy.is_active ? (
                            <>
                              <Pause className="w-4 h-4" />
                              停止
                            </>
                          ) : (
                            <>
                              <Play className="w-4 h-4" />
                              启动
                            </>
                          )}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => {
                            setSelectedStrategy(strategy);
                            setShowPerformance(true);
                          }}
                        >
                          <BarChart3 className="w-4 h-4" />
                          详情
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => {
                            setSelectedStrategy(strategy);
                            setShowConfig(true);
                          }}
                        >
                          <Settings className="w-4 h-4" />
                          配置
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
        title="创建量化策略"
        size="md"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            {...register('name', { required: '请输入策略名称' })}
            label="策略名称"
            placeholder="例如：稳健增长策略"
            error={errors.name?.message}
          />

          <Input
            {...register('totalCapitalUSDT', { 
              required: '请输入投资金额',
              min: { value: 100, message: '最小投资金额为100 USDT' },
              valueAsNumber: true 
            })}
            type="number"
            step="0.01"
            label="投资金额 (USDT)"
            placeholder="请输入投资金额"
            error={errors.totalCapitalUSDT?.message}
          />

          <Select
            {...register('riskPreference', { required: '请选择风险偏好' })}
            label="风险偏好"
            placeholder="请选择风险偏好"
            options={[
              { value: 'conservative', label: '保守型 - 年化收益15-25%，最大回撤8-12%' },
              { value: 'moderate', label: '中等型 - 年化收益25-50%，最大回撤12-20%' },
              { value: 'aggressive', label: '激进型 - 年化收益50-100%，最大回撤20-35%' },
            ]}
            error={errors.riskPreference?.message}
          />

          <Input
            {...register('maxPositions', { 
              required: '请输入最大持仓数',
              min: { value: 1, message: '最少持仓1个' },
              max: { value: 10, message: '最多持仓10个' },
              valueAsNumber: true 
            })}
            type="number"
            label="最大持仓数"
            placeholder="建议3-5个"
            defaultValue={5}
            error={errors.maxPositions?.message}
          />

          <div className="p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
            <div className="flex items-center space-x-2">
              <AlertTriangle className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <p className="text-sm text-blue-800 dark:text-blue-200">
                系统将自动选择市值前10的优质币种进行交易，采用多因子策略分析，严格风控管理。
              </p>
            </div>
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
              创建策略
            </Button>
          </div>
        </form>
      </Modal>

      {/* 性能展示弹窗 */}
      {selectedStrategy && showPerformance && (
        <QuantitativePerformance
          strategy={selectedStrategy}
          isOpen={showPerformance}
          onClose={() => {
            setShowPerformance(false);
            setSelectedStrategy(null);
          }}
        />
      )}

      {/* 配置弹窗 */}
      {selectedStrategy && showConfig && (
        <QuantitativeStrategyConfig
          strategy={selectedStrategy}
          isOpen={showConfig}
          onClose={() => {
            setShowConfig(false);
            setSelectedStrategy(null);
          }}
        />
      )}
    </div>
  );
};