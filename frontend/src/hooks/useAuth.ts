import { useAuthStore } from '@/stores';
import { useNavigate } from 'react-router-dom';
import { useCallback } from 'react';

export const useAuth = () => {
  const navigate = useNavigate();
  const { user, isAuthenticated, login, logout, register } = useAuthStore();

  const handleLogin = useCallback(async (username: string, password: string) => {
    try {
      await login(username, password);
      navigate('/dashboard');
    } catch (error) {
      // 错误已在store中处理
    }
  }, [login, navigate]);

  const handleRegister = useCallback(async (username: string, email: string, password: string) => {
    try {
      await register(username, email, password);
      navigate('/login');
    } catch (error) {
      // 错误已在store中处理
    }
  }, [register, navigate]);

  const handleLogout = useCallback(() => {
    logout();
    navigate('/login');
  }, [logout, navigate]);

  return {
    user,
    isAuthenticated,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
  };
};