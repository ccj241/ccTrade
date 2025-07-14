import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { 
  Button, 
  Card, 
  CardContent, 
  CardHeader, 
  Input,
  Modal
} from '@/components/ui';
import { useAuthStore } from '@/stores';
import { formatDate } from '@/utils';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { changePasswordSchema, apiKeysSchema } from '@/utils/validation';
import { z } from 'zod';
import { useMutation } from '@tanstack/react-query';
import { authApi } from '@/api';
import toast from 'react-hot-toast';

type ChangePasswordFormData = z.infer<typeof changePasswordSchema>;
type APIKeysFormData = z.infer<typeof apiKeysSchema>;

export const Profile: React.FC = () => {
  const { user } = useAuthStore();
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [isAPIKeysModalOpen, setIsAPIKeysModalOpen] = useState(false);

  const {
    register: registerPassword,
    handleSubmit: handlePasswordSubmit,
    reset: resetPassword,
    formState: { errors: passwordErrors },
  } = useForm<ChangePasswordFormData>({
    resolver: zodResolver(changePasswordSchema),
  });

  const {
    register: registerAPIKeys,
    handleSubmit: handleAPIKeysSubmit,
    reset: resetAPIKeys,
    formState: { errors: apiKeysErrors },
  } = useForm<APIKeysFormData>({
    resolver: zodResolver(apiKeysSchema),
  });

  const changePassword = useMutation({
    mutationFn: authApi.changePassword,
    onSuccess: () => {
      toast.success('密码修改成功');
      resetPassword();
      setIsPasswordModalOpen(false);
    },
  });

  const updateAPIKeys = useMutation({
    mutationFn: authApi.updateAPIKeys,
    onSuccess: () => {
      toast.success('API密钥更新成功');
      resetAPIKeys();
      setIsAPIKeysModalOpen(false);
    },
  });

  const onPasswordSubmit = async (data: ChangePasswordFormData) => {
    try {
      await changePassword.mutateAsync(data);
    } catch (error) {
      // 错误已在mutation中处理
    }
  };

  const onAPIKeysSubmit = async (data: APIKeysFormData) => {
    try {
      await updateAPIKeys.mutateAsync(data);
    } catch (error) {
      // 错误已在mutation中处理
    }
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      {/* 页面标题 */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">个人资料</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          管理您的账户信息和安全设置
        </p>
      </div>

      {/* 基本信息 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <Card>
          <CardHeader>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">基本信息</h2>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  用户名
                </label>
                <p className="text-gray-900 dark:text-white">{user?.username}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  邮箱
                </label>
                <p className="text-gray-900 dark:text-white">{user?.email}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  角色
                </label>
                <p className="text-gray-900 dark:text-white">
                  {user?.role === 'admin' ? '管理员' : '普通用户'}
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  状态
                </label>
                <p className="text-gray-900 dark:text-white">
                  {user?.status === 'active' ? '正常' : user?.status === 'pending' ? '待审核' : '已禁用'}
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  注册时间
                </label>
                <p className="text-gray-900 dark:text-white">
                  {user?.created_at ? formatDate(user.created_at) : '-'}
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  最后登录
                </label>
                <p className="text-gray-900 dark:text-white">
                  {user?.last_login_at ? formatDate(user.last_login_at) : '首次登录'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* 安全设置 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
      >
        <Card>
          <CardHeader>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">安全设置</h2>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between py-3 border-b border-gray-200 dark:border-gray-700">
              <div>
                <h3 className="text-sm font-medium text-gray-900 dark:text-white">登录密码</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  定期更改密码可以提高账户安全性
                </p>
              </div>
              <Button variant="secondary" onClick={() => setIsPasswordModalOpen(true)}>
                修改密码
              </Button>
            </div>

            <div className="flex items-center justify-between py-3">
              <div>
                <h3 className="text-sm font-medium text-gray-900 dark:text-white">API密钥</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  {user?.has_api_key ? '已配置API密钥' : '未配置API密钥'}
                </p>
              </div>
              <Button variant="secondary" onClick={() => setIsAPIKeysModalOpen(true)}>
                {user?.has_api_key ? '更新密钥' : '配置密钥'}
              </Button>
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* 修改密码弹窗 */}
      <Modal
        isOpen={isPasswordModalOpen}
        onClose={() => setIsPasswordModalOpen(false)}
        title="修改密码"
      >
        <form onSubmit={handlePasswordSubmit(onPasswordSubmit)} className="space-y-4">
          <Input
            {...registerPassword('old_password')}
            type="password"
            label="原密码"
            placeholder="请输入原密码"
            error={passwordErrors.old_password?.message}
          />
          <Input
            {...registerPassword('new_password')}
            type="password"
            label="新密码"
            placeholder="请输入新密码"
            error={passwordErrors.new_password?.message}
          />
          <Input
            {...registerPassword('confirm_password')}
            type="password"
            label="确认新密码"
            placeholder="请再次输入新密码"
            error={passwordErrors.confirm_password?.message}
          />
          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setIsPasswordModalOpen(false)}
            >
              取消
            </Button>
            <Button type="submit" loading={changePassword.isPending}>
              确认修改
            </Button>
          </div>
        </form>
      </Modal>

      {/* API密钥弹窗 */}
      <Modal
        isOpen={isAPIKeysModalOpen}
        onClose={() => setIsAPIKeysModalOpen(false)}
        title="配置API密钥"
      >
        <form onSubmit={handleAPIKeysSubmit(onAPIKeysSubmit)} className="space-y-4">
          <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              请确保您的API密钥具有正确的权限，并妥善保管密钥信息。
            </p>
          </div>
          <Input
            {...registerAPIKeys('api_key')}
            label="API Key"
            placeholder="请输入API Key"
            error={apiKeysErrors.api_key?.message}
          />
          <Input
            {...registerAPIKeys('secret_key')}
            type="password"
            label="Secret Key"
            placeholder="请输入Secret Key"
            error={apiKeysErrors.secret_key?.message}
          />
          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setIsAPIKeysModalOpen(false)}
            >
              取消
            </Button>
            <Button type="submit" loading={updateAPIKeys.isPending}>
              保存配置
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
};