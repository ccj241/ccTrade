import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CardHeader, SkeletonCard, Button, Modal, Input } from '@/components/ui';
import { useAuthStore, useFavoriteStore } from '@/stores';
import { useBalance, useFavoritePrices } from '@/hooks';
import { formatCurrency, formatPercent, formatDate, formatNumber } from '@/utils';
import { useQuery } from '@tanstack/react-query';
import { authApi } from '@/api';
import { Balance } from '@/types';

// 统计卡片组件
const StatCard: React.FC<{
  title: string;
  value: string | number;
  change?: number;
  icon: React.ReactNode;
  color: string;
}> = ({ title, value, change, icon, color }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    whileHover={{ y: -4 }}
    className="relative overflow-hidden"
  >
    <Card glass hover>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
              {value}
            </p>
            {change !== undefined && (
              <p className={`text-sm mt-2 ${change >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                {change >= 0 ? '↑' : '↓'} {formatPercent(Math.abs(change))}
              </p>
            )}
          </div>
          <div className={`p-3 rounded-lg ${color}`}>
            {icon}
          </div>
        </div>
        <div className={`absolute -bottom-2 -right-2 w-24 h-24 rounded-full ${color} opacity-10`} />
      </CardContent>
    </Card>
  </motion.div>
);

export const Dashboard: React.FC = () => {
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const { data: balance, isLoading: balanceLoading } = useBalance();
  const { data: profileData, isLoading: profileLoading } = useQuery({
    queryKey: ['profile'],
    queryFn: () => authApi.getProfile(),
  });
  const { data: favoritePrices, isLoading: pricesLoading } = useFavoritePrices();

  const stats = profileData?.stats;
  
  // 收藏币对相关
  const { favoritePairs, addFavoritePair, removeFavoritePair, isFavorite } = useFavoriteStore();
  const [showAddPairModal, setShowAddPairModal] = useState(false);
  const [newPairSymbol, setNewPairSymbol] = useState('');
  
  // 资产分布分页
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  // 处理余额数据，转换为统一的 Balance 数组格式
  const balances: Balance[] = React.useMemo(() => {
    if (!balance?.spot?.balances) return [];
    
    return balance.spot.balances
      .map(b => ({
        asset: b.asset,
        free: parseFloat(b.free) || 0,
        locked: parseFloat(b.locked) || 0,
        total: (parseFloat(b.free) || 0) + (parseFloat(b.locked) || 0)
      }))
      .filter(b => b.total > 0); // 只显示有余额的资产
  }, [balance]);

  // 计算总资产
  const totalAssets = balances.reduce((sum, item) => sum + (item.total * 1), 0) || 0;
  
  // 分页逻辑
  const totalPages = Math.ceil(balances.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const currentBalances = balances.slice(startIndex, endIndex);
  
  // 添加币对
  const handleAddPair = () => {
    if (!newPairSymbol.trim()) return;
    
    const symbol = newPairSymbol.toUpperCase();
    // 简单解析币对，假设格式为 BTCUSDT
    const baseAsset = symbol.slice(0, -4); // 去掉最后4位（USDT）
    const quoteAsset = symbol.slice(-4);
    
    addFavoritePair({
      symbol,
      baseAsset,
      quoteAsset,
    });
    
    setNewPairSymbol('');
    setShowAddPairModal(false);
  };

  const statCards = [
    {
      title: '总资产 (USDT)',
      value: formatCurrency(totalAssets),
      change: 0.125,
      icon: <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>,
      color: 'bg-gradient-to-br from-blue-500 to-blue-600',
    },
    {
      title: '活跃策略',
      value: stats?.active_strategies || 0,
      icon: <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>,
      color: 'bg-gradient-to-br from-green-500 to-green-600',
    },
    {
      title: '总订单数',
      value: stats?.total_orders || 0,
      icon: <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
      </svg>,
      color: 'bg-gradient-to-br from-purple-500 to-purple-600',
    },
    {
      title: '总收益 (USDT)',
      value: formatCurrency(stats?.total_profit || 0),
      change: stats?.total_profit ? (stats.total_profit / totalAssets) : 0,
      icon: <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
      </svg>,
      color: 'bg-gradient-to-br from-orange-500 to-orange-600',
    },
  ];

  if (profileLoading || balanceLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">仪表盘</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">欢迎回来，{user?.username}</p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">仪表盘</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          欢迎回来，{user?.username} • 最后登录时间：{user?.last_login_at ? formatDate(user.last_login_at) : '首次登录'}
        </p>
      </div>

      {/* API密钥提示 */}
      {!user?.has_api_key && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg"
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <svg className="w-5 h-5 text-yellow-600 dark:text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <div>
                <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                  您还未设置API密钥
                </p>
                <p className="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                  请先配置币安API密钥后才能使用交易功能
                </p>
              </div>
            </div>
            <button
              onClick={() => navigate('/profile')}
              className="px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white text-sm font-medium rounded-lg transition-colors"
            >
              立即设置
            </button>
          </div>
        </motion.div>
      )}

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <motion.div
            key={stat.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <StatCard {...stat} />
          </motion.div>
        ))}
      </div>

      {/* 收藏币对 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
      >
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">收藏币对</h2>
              <Button
                size="sm"
                variant="primary"
                onClick={() => setShowAddPairModal(true)}
              >
                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                添加币对
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {favoritePairs.length === 0 ? (
              <p className="text-center text-gray-500 dark:text-gray-400 py-4">
                暂无收藏的币对，点击上方按钮添加
              </p>
            ) : (
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
                {favoritePairs.map((pair) => {
                  const priceInfo = favoritePrices?.find(p => p.symbol === pair.symbol);
                  return (
                    <motion.div
                      key={pair.symbol}
                      initial={{ opacity: 0, scale: 0.9 }}
                      animate={{ opacity: 1, scale: 1 }}
                      className="relative group"
                    >
                      <div className="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors">
                        <div className="flex items-center justify-between mb-2">
                          <span className="font-medium text-sm text-gray-900 dark:text-white">
                            {pair.symbol}
                          </span>
                          <button
                            onClick={() => removeFavoritePair(pair.symbol)}
                            className="opacity-0 group-hover:opacity-100 transition-opacity text-red-500 hover:text-red-600"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                          </button>
                        </div>
                        <div className="text-lg font-semibold text-gray-900 dark:text-white">
                          {priceInfo ? formatNumber(priceInfo.price, 4) : '-'}
                        </div>
                        {priceInfo && priceInfo.changePercent24h !== undefined && (
                          <div className={`text-sm mt-1 ${priceInfo.changePercent24h >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                            {priceInfo.changePercent24h >= 0 ? '+' : ''}{formatPercent(priceInfo.changePercent24h / 100, 2)}
                          </div>
                        )}
                      </div>
                    </motion.div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </motion.div>

      {/* 资产分布 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
      >
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                资产分布 ({balances.length} 个资产)
              </h2>
              {totalPages > 1 && (
                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                    className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                    </svg>
                  </button>
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    {currentPage} / {totalPages}
                  </span>
                  <button
                    onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                    disabled={currentPage === totalPages}
                    className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {currentBalances.map((asset) => {
                const percentage = (asset.total * 1) / totalAssets;
                return (
                  <div key={asset.asset} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                        {asset.asset}
                      </span>
                      <span className="text-sm text-gray-500 dark:text-gray-400">
                        {formatCurrency(asset.total)} ({formatPercent(percentage)})
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                      <motion.div
                        initial={{ width: 0 }}
                        animate={{ width: `${percentage * 100}%` }}
                        transition={{ duration: 1, ease: 'easeOut' }}
                        className="bg-gradient-to-r from-primary-400 to-primary-600 h-2 rounded-full"
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* 快速操作 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
      >
        <Card>
          <CardHeader>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">快速操作</h2>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="p-4 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg hover:border-primary-500 dark:hover:border-primary-400 transition-colors group"
              >
                <svg className="w-8 h-8 mx-auto text-gray-400 group-hover:text-primary-500 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-100">
                  创建新策略
                </p>
              </motion.button>

              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="p-4 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg hover:border-primary-500 dark:hover:border-primary-400 transition-colors group"
              >
                <svg className="w-8 h-8 mx-auto text-gray-400 group-hover:text-primary-500 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-100">
                  同步账户数据
                </p>
              </motion.button>

              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="p-4 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg hover:border-primary-500 dark:hover:border-primary-400 transition-colors group"
              >
                <svg className="w-8 h-8 mx-auto text-gray-400 group-hover:text-primary-500 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-100">
                  查看报告
                </p>
              </motion.button>
            </div>
          </CardContent>
        </Card>
      </motion.div>
      
      {/* 添加币对模态框 */}
      <Modal
        isOpen={showAddPairModal}
        onClose={() => {
          setShowAddPairModal(false);
          setNewPairSymbol('');
        }}
        title="添加收藏币对"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              币对符号
            </label>
            <Input
              value={newPairSymbol}
              onChange={(e) => setNewPairSymbol(e.target.value.toUpperCase())}
              placeholder="例如: BTCUSDT, ETHUSDT"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  handleAddPair();
                }
              }}
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              请输入完整的币对符号，通常以USDT结尾
            </p>
          </div>
          <div className="flex justify-end space-x-3">
            <Button
              variant="secondary"
              onClick={() => {
                setShowAddPairModal(false);
                setNewPairSymbol('');
              }}
            >
              取消
            </Button>
            <Button
              variant="primary"
              onClick={handleAddPair}
              disabled={!newPairSymbol.trim()}
            >
              添加
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};