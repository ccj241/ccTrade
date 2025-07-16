package models

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type UserStatus string

const (
	StatusPending  UserStatus = "pending"
	StatusActive   UserStatus = "active"
	StatusDisabled UserStatus = "disabled"
)

type StrategyType string

const (
	StrategySimple      StrategyType = "simple"
	StrategyIceberg     StrategyType = "iceberg"
	StrategySlowIceberg StrategyType = "slow_iceberg"
	StrategyGrid        StrategyType = "grid"
	StrategyDCA         StrategyType = "dca"
	// 高级策略类型
	StrategyQuantitative    StrategyType = "quantitative"     // 综合量化策略
	StrategyWeightedScoring StrategyType = "weighted_scoring" // 加权评分策略
	// 预留接口，后续添加自定义策略
	// StrategyCustom1   StrategyType = "custom1"
	// StrategyCustom2   StrategyType = "custom2"
)

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusNew             OrderStatus = "new"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCanceled        OrderStatus = "canceled"
	OrderStatusPendingCancel   OrderStatus = "pending_cancel"
	OrderStatusExpired         OrderStatus = "expired"
	OrderStatusRejected        OrderStatus = "rejected"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type OrderType string

const (
	OrderTypeMarket          OrderType = "market"
	OrderTypeLimit           OrderType = "limit"
	OrderTypeStop            OrderType = "stop"
	OrderTypeStopLoss        OrderType = "stop_loss"
	OrderTypeStopLossLimit   OrderType = "stop_loss_limit"
	OrderTypeTakeProfit      OrderType = "take_profit"
	OrderTypeTakeProfitLimit OrderType = "take_profit_limit"
	OrderTypeLimitMaker      OrderType = "limit_maker"
)

type PositionSide string

const (
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
	PositionSideBoth  PositionSide = "both"
)

type MarginType string

const (
	MarginTypeIsolated MarginType = "isolated"
	MarginTypeCross    MarginType = "cross"
)

// TimeInForce 有效期类型
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill or Kill
	TimeInForceGTX TimeInForce = "GTX" // Good Till Crossing
)

// WorkingType 触发价格类型
type WorkingType string

const (
	WorkingTypeContract WorkingType = "CONTRACT_PRICE"
	WorkingTypeMark     WorkingType = "MARK_PRICE"
)

// WithdrawStatus 提现状态
type WithdrawStatus string

const (
	WithdrawStatusEmailSent        WithdrawStatus = "email_sent"
	WithdrawStatusCancelled        WithdrawStatus = "cancelled"
	WithdrawStatusAwaitingApproval WithdrawStatus = "awaiting_approval"
	WithdrawStatusRejected         WithdrawStatus = "rejected"
	WithdrawStatusProcessing       WithdrawStatus = "processing"
	WithdrawStatusFailure          WithdrawStatus = "failure"
	WithdrawStatusCompleted        WithdrawStatus = "completed"
)
