import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface FavoritePair {
  symbol: string;
  baseAsset: string;
  quoteAsset: string;
  addedAt: Date;
}

interface FavoriteStore {
  favoritePairs: FavoritePair[];
  addFavoritePair: (pair: Omit<FavoritePair, 'addedAt'>) => void;
  removeFavoritePair: (symbol: string) => void;
  isFavorite: (symbol: string) => boolean;
  clearFavorites: () => void;
}

export const useFavoriteStore = create<FavoriteStore>()(
  persist(
    (set, get) => ({
      favoritePairs: [],
      
      addFavoritePair: (pair) => {
        const currentPairs = get().favoritePairs;
        // 避免重复添加
        if (currentPairs.some(p => p.symbol === pair.symbol)) {
          return;
        }
        
        set({
          favoritePairs: [
            ...currentPairs,
            {
              ...pair,
              addedAt: new Date(),
            },
          ],
        });
      },
      
      removeFavoritePair: (symbol) => {
        set((state) => ({
          favoritePairs: state.favoritePairs.filter(p => p.symbol !== symbol),
        }));
      },
      
      isFavorite: (symbol) => {
        return get().favoritePairs.some(p => p.symbol === symbol);
      },
      
      clearFavorites: () => {
        set({ favoritePairs: [] });
      },
    }),
    {
      name: 'favorite-pairs-storage',
    }
  )
);