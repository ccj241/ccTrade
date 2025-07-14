import { QueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error: any) => {
        // 对于 404 和 401 错误不重试
        if (error?.response?.status === 404 || error?.response?.status === 401) {
          return false;
        }
        // 其他错误最多重试2次
        return failureCount < 2;
      },
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000, // 5分钟
      gcTime: 10 * 60 * 1000, // 10分钟（替代原来的cacheTime）
    },
    mutations: {
      onError: (error: any) => {
        // 全局mutation错误处理
        const message = error?.response?.data?.message || error?.message || '操作失败';
        toast.error(message);
      },
    },
  },
});

// 设置全局错误处理
queryClient.setMutationDefaults(['createStrategy'], {
  mutationFn: async (data: any) => {
    // 策略创建的默认处理
    throw new Error('策略创建函数未实现');
  },
  onError: (error: any) => {
    toast.error('创建策略失败：' + (error?.message || '未知错误'));
  },
  onSuccess: () => {
    toast.success('策略创建成功');
  },
});

// 设置查询默认值
queryClient.setQueryDefaults(['strategies'], {
  staleTime: 30 * 1000, // 30秒
  gcTime: 5 * 60 * 1000, // 5分钟
});