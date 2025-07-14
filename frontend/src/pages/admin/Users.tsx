import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { 
  Button, 
  Card, 
  CardContent, 
  CardHeader, 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell,
  Badge,
  Modal,
  Select,
  SkeletonTable
} from '@/components/ui';
import { formatDate, USER_STATUS } from '@/utils';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { authApi } from '@/api';
import toast from 'react-hot-toast';

export const AdminUsers: React.FC = () => {
  const [page, setPage] = useState(1);
  const [selectedUser, setSelectedUser] = useState<any>(null);
  const [isStatusModalOpen, setIsStatusModalOpen] = useState(false);
  const [isRoleModalOpen, setIsRoleModalOpen] = useState(false);
  const queryClient = useQueryClient();
  
  const { data, isLoading } = useQuery({
    queryKey: ['admin-users', page],
    queryFn: () => authApi.getAllUsers(page, 10),
  });

  const approveUser = useMutation({
    mutationFn: (userId: number) => authApi.approveUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      toast.success('用户审核通过');
    },
  });

  const updateUserStatus = useMutation({
    mutationFn: ({ userId, status }: { userId: number; status: string }) => 
      authApi.updateUserStatus(userId, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      toast.success('用户状态更新成功');
      setIsStatusModalOpen(false);
    },
  });

  const updateUserRole = useMutation({
    mutationFn: ({ userId, role }: { userId: number; role: string }) => 
      authApi.updateUserRole(userId, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      toast.success('用户角色更新成功');
      setIsRoleModalOpen(false);
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">用户管理</h1>
        </div>
        <SkeletonTable />
      </div>
    );
  }

  const pendingUsers = data?.data?.filter(u => u.status === 'pending').length || 0;
  const activeUsers = data?.data?.filter(u => u.status === 'active').length || 0;
  const totalUsers = data?.total || 0;

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">用户管理</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            管理系统用户和权限
          </p>
        </div>
      </div>

      {/* 统计信息 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">总用户数</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
              {totalUsers}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">活跃用户</p>
            <p className="text-2xl font-bold text-green-600 dark:text-green-400 mt-1">
              {activeUsers}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">待审核</p>
            <p className="text-2xl font-bold text-yellow-600 dark:text-yellow-400 mt-1">
              {pendingUsers}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">管理员</p>
            <p className="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">
              {data?.data?.filter(u => u.role === 'admin').length || 0}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 用户列表 */}
      <Card>
        <CardHeader>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">用户列表</h2>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>用户名</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead>角色</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>API密钥</TableHead>
                <TableHead>注册时间</TableHead>
                <TableHead>最后登录</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data?.data?.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>{user.id}</TableCell>
                  <TableCell className="font-medium">{user.username}</TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>
                    <Badge variant={user.role === 'admin' ? 'info' : 'default'}>
                      {user.role === 'admin' ? '管理员' : '普通用户'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={USER_STATUS[user.status]?.color as any}>
                      {USER_STATUS[user.status]?.label}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {/* 对于加密用户(如admin)，特殊处理 */}
                    {(user.has_api_key || (user.is_encrypted && user.username === 'admin')) ? (
                      <Badge variant="success">已配置</Badge>
                    ) : (
                      <a 
                        href="/profile" 
                        className="inline-flex items-center space-x-1 text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        <Badge variant="default">未配置</Badge>
                        <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                        </svg>
                      </a>
                    )}
                  </TableCell>
                  <TableCell>{formatDate(user.created_at, 'MM-DD HH:mm')}</TableCell>
                  <TableCell>
                    {user.last_login_at 
                      ? formatDate(user.last_login_at, 'MM-DD HH:mm')
                      : '-'
                    }
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end space-x-2">
                      {user.status === 'pending' && (
                        <Button
                          size="sm"
                          variant="primary"
                          onClick={() => approveUser.mutate(user.id)}
                        >
                          审核通过
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setSelectedUser(user);
                          setIsStatusModalOpen(true);
                        }}
                      >
                        状态
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setSelectedUser(user);
                          setIsRoleModalOpen(true);
                        }}
                      >
                        角色
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* 分页 */}
      {data && data.total > 10 && (
        <div className="flex items-center justify-center space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setPage(page - 1)}
            disabled={page === 1}
          >
            上一页
          </Button>
          <span className="text-sm text-gray-600 dark:text-gray-400">
            第 {page} 页，共 {Math.ceil(data.total / 10)} 页
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setPage(page + 1)}
            disabled={page === Math.ceil(data.total / 10)}
          >
            下一页
          </Button>
        </div>
      )}

      {/* 更新状态弹窗 */}
      <Modal
        isOpen={isStatusModalOpen}
        onClose={() => setIsStatusModalOpen(false)}
        title="更新用户状态"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            用户：{selectedUser?.username}
          </p>
          <Select
            label="状态"
            defaultValue={selectedUser?.status}
            options={[
              { value: 'active', label: '正常' },
              { value: 'disabled', label: '禁用' },
              { value: 'pending', label: '待审核' },
            ]}
            onChange={(e) => {
              const newStatus = e.target.value;
              updateUserStatus.mutate({ 
                userId: selectedUser.id, 
                status: newStatus 
              });
            }}
          />
          <div className="flex justify-end space-x-3 pt-4">
            <Button
              variant="ghost"
              onClick={() => setIsStatusModalOpen(false)}
            >
              取消
            </Button>
          </div>
        </div>
      </Modal>

      {/* 更新角色弹窗 */}
      <Modal
        isOpen={isRoleModalOpen}
        onClose={() => setIsRoleModalOpen(false)}
        title="更新用户角色"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            用户：{selectedUser?.username}
          </p>
          <Select
            label="角色"
            defaultValue={selectedUser?.role}
            options={[
              { value: 'user', label: '普通用户' },
              { value: 'admin', label: '管理员' },
            ]}
            onChange={(e) => {
              const newRole = e.target.value;
              updateUserRole.mutate({ 
                userId: selectedUser.id, 
                role: newRole 
              });
            }}
          />
          <div className="flex justify-end space-x-3 pt-4">
            <Button
              variant="ghost"
              onClick={() => setIsRoleModalOpen(false)}
            >
              取消
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};