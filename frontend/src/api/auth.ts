import apiClient from './client';
import { User, LoginRequest, RegisterRequest, LoginResponse, UserStats, ChangePasswordForm, UpdateAPIKeysForm, PaginatedResponse } from '@/types';

export const authApi = {
  // 注册
  register: (data: RegisterRequest) => 
    apiClient.post<User>('/register', data),

  // 登录
  login: (data: LoginRequest) => 
    apiClient.post<LoginResponse>('/login', data),

  // 获取个人信息
  getProfile: () => 
    apiClient.get<{ user: User; stats: UserStats }>('/profile'),

  // 更新个人信息
  updateProfile: (data: Partial<User>) => 
    apiClient.put('/profile', data),

  // 修改密码
  changePassword: (data: ChangePasswordForm) => 
    apiClient.post('/change-password', data),

  // 更新API密钥
  updateAPIKeys: (data: UpdateAPIKeysForm) => 
    apiClient.post('/api-keys', data),

  // 管理员接口
  getAllUsers: (page: number = 1, limit: number = 10) => 
    apiClient.getPaginated<User>(`/admin/users?page=${page}&limit=${limit}`),

  approveUser: (userId: number) => 
    apiClient.post(`/admin/users/${userId}/approve`),

  updateUserStatus: (userId: number, status: string) => 
    apiClient.put(`/admin/users/${userId}/status`, { status }),

  updateUserRole: (userId: number, role: string) => 
    apiClient.put(`/admin/users/${userId}/role`, { role }),
};