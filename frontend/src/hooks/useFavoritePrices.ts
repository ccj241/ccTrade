import { useQuery } from '@tanstack/react-query';
import { generalApi } from '@/api';
import { useFavoriteStore } from '@/stores';

interface PriceInfo {
  symbol: string;
  price: number;
  change24h?: number;
  changePercent24h?: number;
}

export const useFavoritePrices = () => {
  const { favoritePairs } = useFavoriteStore();

  return useQuery({
    queryKey: ['favoritePrices', favoritePairs.map(p => p.symbol)],
    queryFn: async () => {
      if (favoritePairs.length === 0) return [];
      
      // 并行获取所有收藏币对的价格
      const pricePromises = favoritePairs.map(pair => 
        generalApi.getPrice(pair.symbol).catch(() => null)
      );
      
      const priceResponses = await Promise.all(pricePromises);
      
      return priceResponses
        .filter(res => res !== null)
        .map(res => ({
          symbol: res.data.symbol,
          price: res.data.price,
          // TODO: 后端需要返回24小时变化数据
          change24h: 0,
          changePercent24h: 0,
        })) as PriceInfo[];
    },
    refetchInterval: 5000, // 每5秒刷新一次价格
    enabled: favoritePairs.length > 0,
  });
};