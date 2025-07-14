import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type Theme = 'light' | 'dark';

interface ThemeState {
  theme: Theme;
  toggleTheme: () => void;
  setTheme: (theme: Theme) => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: 'dark',
      
      toggleTheme: () => {
        set((state) => {
          const newTheme = state.theme === 'light' ? 'dark' : 'light';
          updateThemeClass(newTheme);
          return { theme: newTheme };
        });
      },
      
      setTheme: (theme: Theme) => {
        updateThemeClass(theme);
        set({ theme });
      },
    }),
    {
      name: 'theme-storage',
    }
  )
);

// 更新HTML元素的class
function updateThemeClass(theme: Theme) {
  const root = document.documentElement;
  if (theme === 'dark') {
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
  }
}