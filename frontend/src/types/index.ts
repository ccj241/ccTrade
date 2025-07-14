// API响应类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

export interface PaginatedResponse<T = any> {
  code: number;
  message: string;
  data?: T[];
  total: number;
  page: number;
  limit: number;
}

// 用户相关类型
export type UserRole = 'admin' | 'user';
export type UserStatus = 'pending' | 'active' | 'disabled';

export interface User {
  id: number;
  username: string;
  email: string;
  role: UserRole;
  status: UserStatus;
  api_key?: string;
  secret_key?: string;
  is_encrypted: boolean;
  has_api_key: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface UserStats {
  total_strategies: number;
  active_strategies: number;
  total_orders: number;
  total_profit: number;
  total_volume: number;
}

// 策略相关类型
export type StrategyType = 
  | 'simple'
  | 'iceberg'
  | 'slow_iceberg'
  | 'grid'
  | 'dca'
  | 'arbitrage'
  | 'breakout'
  | 'momentum'
  | 'mean_revert'
  | 'swing'
  | 'pyramid'
  | 'signal'
  | 'custom';

export type OrderSide = 'buy' | 'sell';
export type OrderType = 
  | 'market'
  | 'limit'
  | 'stop'
  | 'stop_loss'
  | 'stop_loss_limit'
  | 'take_profit'
  | 'take_profit_limit'
  | 'limit_maker';

export type OrderStatus = 
  | 'pending'
  | 'new'
  | 'partially_filled'
  | 'filled'
  | 'canceled'
  | 'pending_cancel'
  | 'expired'
  | 'rejected';

export interface Strategy {
  id: number;
  user_id: number;
  name: string;
  symbol: string;
  type: StrategyType;
  side: OrderSide;
  quantity: number;
  price: number;
  trigger_price: number;
  stop_price: number;
  take_profit: number;
  stop_loss: number;
  config: Record<string, any>;
  is_active: boolean;
  is_completed: boolean;
  auto_restart: boolean;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: number;
  user_id: number;
  strategy_id?: number;
  symbol: string;
  order_id: string;
  client_order_id: string;
  side: OrderSide;
  type: OrderType;
  quantity: number;
  price: number;
  stop_price: number;
  status: OrderStatus;
  executed_qty: number;
  cumulative_quote_qty: number;
  time_in_force: string;
  is_working: boolean;
  orig_qty: number;
  created_at: string;
  updated_at: string;
}

// 期货相关类型
export type PositionSide = 'long' | 'short' | 'both';
export type MarginType = 'isolated' | 'cross';

export interface FuturesStrategy {
  id: number;
  user_id: number;
  name: string;
  symbol: string;
  type: StrategyType;
  side: OrderSide;  // buy=做多，sell=做空
  margin_amount: number;  // 保证金本值（USDT）
  price: number;  // 触发价格
  float_basis_points: number;  // 万分比浮动
  take_profit_bp: number;  // 止盈万分比
  stop_loss_bp: number;  // 止损万分比
  leverage: number;  // 杠杆倍数
  margin_type: MarginType;
  config: Record<string, any>;
  is_active: boolean;
  is_completed: boolean;
  auto_restart: boolean;
  // 计算字段
  order_quantity?: number;  // 实际下单数量
  estimated_profit?: number;  // 预计盈利（USDT）
  estimated_loss?: number;  // 预计亏损（USDT）
  liquidation_price?: number;  // 预计爆仓价格
  created_at: string;
  updated_at: string;
}

export interface FuturesPosition {
  id: number;
  user_id: number;
  symbol: string;
  position_side: PositionSide;
  position_amt: number;
  entry_price: number;
  mark_price: number;
  unrealized_profit: number;
  liquidation_price: number;
  leverage: number;
  max_notional_value: number;
  margin_type: MarginType;
  isolated_margin: number;
  is_auto_add_margin: boolean;
  created_at: string;
  updated_at: string;
}

// 双币投资相关类型
export interface DualInvestmentProduct {
  id: number;
  symbol: string;
  exercise_price: number;
  delivery_date: string;
  apy: number;
  duration: number;
  is_call: boolean;
  min_amount: number;
  max_amount: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface DualInvestmentStrategy {
  id: number;
  user_id: number;
  name: string;
  symbol: string;
  target_price: number;
  min_apy: number;
  max_investment: number;
  single_investment: number;
  is_call: boolean;
  is_active: boolean;
  auto_compound: boolean;
  created_at: string;
  updated_at: string;
}

// 提现相关类型
export type WithdrawStatus = 
  | 'email_sent'
  | 'cancelled'
  | 'awaiting_approval'
  | 'rejected'
  | 'processing'
  | 'failure'
  | 'completed';

export interface Withdrawal {
  id: number;
  user_id: number;
  asset: string;
  amount: number;
  address: string;
  network: string;
  memo?: string;
  rule_name: string;
  status: WithdrawStatus;
  frequency: string;
  next_execute_time?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface WithdrawalHistory {
  id: number;
  user_id: number;
  withdrawal_id?: number;
  asset: string;
  amount: number;
  address: string;
  network: string;
  memo?: string;
  tx_id?: string;
  status: WithdrawStatus;
  created_at: string;
  updated_at: string;
}

// 其他通用类型
export interface Balance {
  asset: string;
  free: number;
  locked: number;
  total: number;
}

export interface BalanceResponse {
  spot: {
    balances?: Array<{
      asset: string;
      free: string;
      locked: string;
    }>;
  };
  futures?: any;
}

export interface Symbol {
  symbol: string;
  status: string;
  base_asset: string;
  quote_asset: string;
  base_precision: number;
  quote_precision: number;
  min_quantity: number;
  max_quantity: number;
  step_size: number;
  min_notional: number;
}

export interface Price {
  symbol: string;
  price: number;
}

// 表单类型
export interface ChangePasswordForm {
  old_password: string;
  new_password: string;
}

export interface UpdateAPIKeysForm {
  api_key: string;
  secret_key: string;
}

export interface CreateStrategyForm {
  name: string;
  symbol: string;
  type: StrategyType;
  side: OrderSide;
  quantity: number;
  price?: number;
  trigger_price?: number;
  stop_price?: number;
  take_profit?: number;
  stop_loss?: number;
  config?: Record<string, any>;
  auto_restart?: boolean;
}

export interface CreateFuturesStrategyForm {
  name: string;
  symbol: string;
  type: StrategyType;
  side: OrderSide;
  margin_amount: number;
  price: number;
  float_basis_points: number;
  take_profit_bp: number;
  stop_loss_bp: number;
  leverage: number;
  margin_type: MarginType;
  config?: Record<string, any>;
  auto_restart?: boolean;
}