import { z } from 'zod';

// 用户验证
export const loginSchema = z.object({
  username: z.string().min(3, '用户名至少3个字符'),
  password: z.string().min(8, '密码至少8个字符'),
});

export const registerSchema = z.object({
  username: z.string().min(3, '用户名至少3个字符').max(50, '用户名最多50个字符'),
  email: z.string().email('请输入有效的邮箱地址'),
  password: z.string().min(8, '密码至少8个字符').regex(
    /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
    '密码必须包含大小写字母、数字和特殊字符'
  ),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: '两次输入的密码不一致',
  path: ['confirmPassword'],
});

export const changePasswordSchema = z.object({
  old_password: z.string().min(1, '请输入原密码'),
  new_password: z.string().min(8, '新密码至少8个字符'),
  confirm_password: z.string(),
}).refine((data) => data.new_password === data.confirm_password, {
  message: '两次输入的密码不一致',
  path: ['confirm_password'],
});

export const apiKeysSchema = z.object({
  api_key: z.string().min(1, '请输入API Key'),
  secret_key: z.string().min(1, '请输入Secret Key'),
});

// 策略验证
export const strategySchema = z.object({
  name: z.string().min(1, '请输入策略名称').max(100, '策略名称最多100个字符'),
  symbol: z.string().min(1, '请选择交易对'),
  type: z.string().min(1, '请选择策略类型'),
  side: z.enum(['buy', 'sell']),
  quantity: z.number().positive('数量必须大于0'),
  price: z.number().positive('价格必须大于0').optional(),
  trigger_price: z.number().positive('触发价格必须大于0').optional(),
  stop_price: z.number().positive('止损价格必须大于0').optional(),
  take_profit: z.number().positive('止盈价格必须大于0').optional(),
  stop_loss: z.number().positive('止损价格必须大于0').optional(),
  auto_restart: z.boolean().optional(),
});

// 期货策略验证
export const futuresStrategySchema = z.object({
  name: z.string().min(1, '请输入策略名称').max(100, '策略名称最多100个字符'),
  symbol: z.string().min(1, '请选择交易对'),
  type: z.string().min(1, '请选择策略类型'),
  side: z.enum(['buy', 'sell']),
  margin_amount: z.number().positive('保证金金额必须大于0'),
  price: z.number().positive('触发价格必须大于0'),
  float_basis_points: z.number().min(0, '首单万分比浮动不能为负数').max(10000, '首单万分比浮动不能超过10000'),
  take_profit_bp: z.number().int().min(0, '止盈万分比不能为负数').optional(),
  stop_loss_bp: z.number().int().min(0, '止损万分比不能为负数').optional(),
  leverage: z.number().int().min(1).max(20, '杠杆倍数必须在1-20之间'),
  margin_type: z.enum(['isolated', 'cross']),
  auto_restart: z.boolean().optional(),
  config: z.object({
    layers: z.number().int().min(5).max(10).optional(),
    timeout_minutes: z.number().int().min(1).optional(),
  }).optional(),
});

// 提现验证
export const withdrawalSchema = z.object({
  asset: z.string().min(1, '请选择币种'),
  amount: z.number().positive('提现金额必须大于0'),
  address: z.string().min(1, '请输入提现地址'),
  network: z.string().min(1, '请选择网络'),
  memo: z.string().optional(),
  rule_name: z.string().min(1, '请输入规则名称'),
  frequency: z.string().min(1, '请选择执行频率'),
});