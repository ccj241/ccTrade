// 现货策略类型
export const SPOT_STRATEGY_TYPES = [
  { value: 'simple', label: '简单策略' },
  { value: 'iceberg', label: '冰山策略' },
  { value: 'slow_iceberg', label: '慢速冰山' },
] as const;

// 期货策略类型
export const FUTURES_STRATEGY_TYPES = [
  { value: 'simple', label: '简单策略' },
  { value: 'iceberg', label: '冰山策略' },
  { value: 'slow_iceberg', label: '慢速冰山' },
] as const;

// 高级策略类型 - 预留接口，后续添加自定义策略
export const ADVANCED_STRATEGY_TYPES = [
  // 后续在这里添加自定义策略
  // 例如: { value: 'my_strategy', label: '我的策略' },
] as const;

// 所有策略类型（用于后端兼容）
export const STRATEGY_TYPES = [
  ...SPOT_STRATEGY_TYPES,
  ...ADVANCED_STRATEGY_TYPES,
] as const;

// 订单类型
export const ORDER_TYPES = [
  { value: 'market', label: '市价单' },
  { value: 'limit', label: '限价单' },
  { value: 'stop', label: '止损单' },
  { value: 'stop_loss', label: '止损限价单' },
  { value: 'take_profit', label: '止盈单' },
  { value: 'take_profit_limit', label: '止盈限价单' },
  { value: 'limit_maker', label: '限价只挂单' },
] as const;

// 订单状态
export const ORDER_STATUS = {
  pending: { label: '待处理', color: 'gray' },
  new: { label: '新订单', color: 'blue' },
  partially_filled: { label: '部分成交', color: 'yellow' },
  filled: { label: '已成交', color: 'green' },
  canceled: { label: '已取消', color: 'gray' },
  pending_cancel: { label: '待取消', color: 'orange' },
  expired: { label: '已过期', color: 'gray' },
  rejected: { label: '已拒绝', color: 'red' },
} as const;

// 用户状态
export const USER_STATUS = {
  pending: { label: '待审核', color: 'yellow' },
  active: { label: '正常', color: 'green' },
  disabled: { label: '已禁用', color: 'red' },
} as const;

// 提现状态
export const WITHDRAW_STATUS = {
  email_sent: { label: '邮件已发送', color: 'blue' },
  cancelled: { label: '已取消', color: 'gray' },
  awaiting_approval: { label: '等待审批', color: 'yellow' },
  rejected: { label: '已拒绝', color: 'red' },
  processing: { label: '处理中', color: 'blue' },
  failure: { label: '失败', color: 'red' },
  completed: { label: '已完成', color: 'green' },
} as const;

// 持仓方向
export const POSITION_SIDES = [
  { value: 'long', label: '做多' },
  { value: 'short', label: '做空' },
  { value: 'both', label: '双向' },
] as const;

// 保证金模式
export const MARGIN_TYPES = [
  { value: 'isolated', label: '逐仓' },
  { value: 'cross', label: '全仓' },
] as const;

// 时间频率
export const FREQUENCIES = [
  { value: 'hourly', label: '每小时' },
  { value: 'daily', label: '每天' },
  { value: 'weekly', label: '每周' },
  { value: 'monthly', label: '每月' },
] as const;

// 分页选项
export const PAGE_SIZES = [10, 20, 50, 100] as const;

// 预设提币币种
export const PRESET_ASSETS = [
  { value: 'USDT', label: 'USDT' },
  { value: 'USDC', label: 'USDC' },
  { value: 'BTC', label: 'BTC' },
  { value: 'ETH', label: 'ETH' },
  { value: 'BNB', label: 'BNB' },
  { value: 'BUSD', label: 'BUSD' },
  { value: 'SOL', label: 'SOL' },
  { value: 'ADA', label: 'ADA' },
  { value: 'DOT', label: 'DOT' },
  { value: 'MATIC', label: 'MATIC' },
] as const;

// 提币网络配置
export const WITHDRAWAL_NETWORKS = {
  USDT: [
    { value: 'TRC20', label: 'TRC20 (波场)', fee: '1 USDT' },
    { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '15 USDT' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.8 USDT' },
    { value: 'POLYGON', label: 'Polygon', fee: '0.8 USDT' },
    { value: 'ARBITRUM', label: 'Arbitrum One', fee: '0.8 USDT' },
  ],
  USDC: [
    { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '15 USDC' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.8 USDC' },
    { value: 'POLYGON', label: 'Polygon', fee: '0.8 USDC' },
    { value: 'ARBITRUM', label: 'Arbitrum One', fee: '0.8 USDC' },
    { value: 'SOL', label: 'Solana', fee: '0.8 USDC' },
  ],
  BTC: [
    { value: 'BTC', label: 'Bitcoin', fee: '0.0005 BTC' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.000005 BTC' },
  ],
  ETH: [
    { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '0.005 ETH' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.0005 ETH' },
    { value: 'ARBITRUM', label: 'Arbitrum One', fee: '0.001 ETH' },
  ],
  BNB: [
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.005 BNB' },
    { value: 'BEP2', label: 'BEP2', fee: '0.01 BNB' },
  ],
  BUSD: [
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.8 BUSD' },
    { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '15 BUSD' },
  ],
  SOL: [
    { value: 'SOL', label: 'Solana', fee: '0.01 SOL' },
  ],
  ADA: [
    { value: 'ADA', label: 'Cardano', fee: '1 ADA' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.8 ADA' },
  ],
  DOT: [
    { value: 'DOT', label: 'Polkadot', fee: '0.1 DOT' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.08 DOT' },
  ],
  MATIC: [
    { value: 'POLYGON', label: 'Polygon', fee: '0.1 MATIC' },
    { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '20 MATIC' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.8 MATIC' },
  ],
  DOGE: [
    { value: 'DOGE', label: 'Dogecoin', fee: '5 DOGE' },
    { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.5 DOGE' },
  ],
} as const;