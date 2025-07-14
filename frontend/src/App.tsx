import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { Toaster } from 'react-hot-toast';
import { useAuthStore, useThemeStore } from '@/stores';
import { Layout } from '@/components/layout';
import { PrivateRoute, ErrorBoundary, LoadingSpinner } from '@/components/common';
import { Login, Register, Dashboard } from '@/pages';
import { StrategiesNew as Strategies } from '@/pages/StrategiesNew';
import { queryClient } from '@/utils/queryClient';

// 懒加载页面组件
const Futures = React.lazy(() => import('@/pages/Futures').then(module => ({ default: module.Futures })));
const DualInvestment = React.lazy(() => import('@/pages/DualInvestment').then(module => ({ default: module.DualInvestment })));
const Withdrawals = React.lazy(() => import('@/pages/Withdrawals').then(module => ({ default: module.Withdrawals })));
const Orders = React.lazy(() => import('@/pages/Orders').then(module => ({ default: module.Orders })));
const AdminUsers = React.lazy(() => import('@/pages/admin/Users').then(module => ({ default: module.AdminUsers })));
const Profile = React.lazy(() => import('@/pages/Profile').then(module => ({ default: module.Profile })));
const AdvancedStrategies = React.lazy(() => import('@/pages/AdvancedStrategies').then(module => ({ default: module.AdvancedStrategies })));

function App() {
  const { checkAuth } = useAuthStore();
  const { theme } = useThemeStore();

  useEffect(() => {
    // 检查认证状态
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    // 设置主题
    const root = document.documentElement;
    if (theme === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }, [theme]);

  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
      <Router>
        <Routes>
          {/* 公开路由 */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          
          {/* 私有路由 */}
          <Route
            path="/"
            element={
              <PrivateRoute>
                <Layout />
              </PrivateRoute>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="strategies" element={<Strategies />} />
            <Route
              path="futures"
              element={
                <React.Suspense fallback={<LoadingSpinner />}>
                  <Futures />
                </React.Suspense>
              }
            />
            <Route
              path="advanced-strategies"
              element={
                <React.Suspense fallback={<LoadingSpinner />}>
                  <AdvancedStrategies />
                </React.Suspense>
              }
            />
            <Route
              path="dual-investment"
              element={
                <React.Suspense fallback={<LoadingSpinner />}>
                  <DualInvestment />
                </React.Suspense>
              }
            />
            <Route
              path="withdrawals"
              element={
                <React.Suspense fallback={<LoadingSpinner />}>
                  <Withdrawals />
                </React.Suspense>
              }
            />
            <Route
              path="orders"
              element={
                <React.Suspense fallback={<LoadingSpinner />}>
                  <Orders />
                </React.Suspense>
              }
            />
            <Route
              path="profile"
              element={
                <React.Suspense fallback={<LoadingSpinner />}>
                  <Profile />
                </React.Suspense>
              }
            />
            
            {/* 管理员路由 */}
            <Route
              path="admin/users"
              element={
                <PrivateRoute adminOnly>
                  <React.Suspense fallback={<LoadingSpinner />}>
                    <AdminUsers />
                  </React.Suspense>
                </PrivateRoute>
              }
            />
          </Route>
          
          {/* 404 页面 */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </Router>
      
      {/* Toast 通知 */}
      <Toaster
        position="top-right"
        toastOptions={{
          duration: 3000,
          style: {
            background: theme === 'dark' ? '#1f2937' : '#fff',
            color: theme === 'dark' ? '#f3f4f6' : '#111827',
          },
        }}
      />
      
      {/* React Query 开发工具 */}
      <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-left" />
    </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;