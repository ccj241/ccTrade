import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { strategyApi } from '@/api';
import { CreateStrategyForm } from '@/types';
import toast from 'react-hot-toast';

export const useStrategies = (page: number = 1, limit: number = 10) => {
  return useQuery({
    queryKey: ['strategies', page, limit],
    queryFn: () => strategyApi.getStrategies(page, limit),
    staleTime: 5 * 60 * 1000, // 5分钟
  });
};

export const useStrategy = (strategyId: number) => {
  return useQuery({
    queryKey: ['strategy', strategyId],
    queryFn: () => strategyApi.getStrategyById(strategyId),
    enabled: !!strategyId,
  });
};

export const useCreateStrategy = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateStrategyForm) => strategyApi.createStrategy(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      toast.success('策略创建成功');
    },
    onError: (error: any) => {
      toast.error(`创建失败: ${error.response?.data?.message || '未知错误'}`);
    },
  });
};

export const useUpdateStrategy = (strategyId: number) => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: Partial<CreateStrategyForm>) => 
      strategyApi.updateStrategy(strategyId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      queryClient.invalidateQueries({ queryKey: ['strategy', strategyId] });
      toast.success('策略更新成功');
    },
    onError: (error: any) => {
      toast.error(`更新失败: ${error.response?.data?.message || '未知错误'}`);
    },
  });
};

export const useToggleStrategy = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (strategyId: number) => strategyApi.toggleStrategy(strategyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      toast.success('策略状态已切换');
    },
    onError: (error: any) => {
      toast.error(`操作失败: ${error.response?.data?.message || '未知错误'}`);
    },
  });
};

export const useDeleteStrategy = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (strategyId: number) => strategyApi.deleteStrategy(strategyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      toast.success('策略删除成功');
    },
    onError: (error: any) => {
      toast.error(`删除失败: ${error.response?.data?.message || '未知错误'}`);
    },
  });
};

export const useStrategyStats = (strategyId: number) => {
  return useQuery({
    queryKey: ['strategy-stats', strategyId],
    queryFn: () => strategyApi.getStrategyStats(strategyId),
    enabled: !!strategyId,
  });
};