import { useQuery } from '@tanstack/react-query';
import { generalApi } from '@/api';

export const useBalance = () => {
  return useQuery({
    queryKey: ['balance'],
    queryFn: () => generalApi.getBalance(),
    staleTime: 30 * 1000, // 30秒
    refetchInterval: 60 * 1000, // 每分钟自动刷新
  });
};

export const useTradingSymbols = () => {
  return useQuery({
    queryKey: ['trading-symbols'],
    queryFn: () => generalApi.getTradingSymbols(),
    staleTime: 5 * 60 * 1000, // 5分钟
  });
};

export const useFuturesTradingSymbols = () => {
  return useQuery({
    queryKey: ['futures-trading-symbols'],
    queryFn: () => generalApi.getFuturesTradingSymbols(),
    staleTime: 5 * 60 * 1000, // 5分钟
  });
};

export const usePrice = (symbol: string) => {
  return useQuery({
    queryKey: ['price', symbol],
    queryFn: () => generalApi.getPrice(symbol),
    enabled: !!symbol,
    staleTime: 5 * 1000, // 5秒
    refetchInterval: 10 * 1000, // 每10秒刷新
  });
};