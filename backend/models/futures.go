package models

type FuturesStrategy struct {
	BaseModel
	UserID           uint           `json:"user_id" gorm:"not null;index"`
	Name             string         `json:"name" gorm:"size:100;not null"`
	Symbol           string         `json:"symbol" gorm:"size:20;not null;index"`
	Type             StrategyType   `json:"type" gorm:"not null"`
	Side             OrderSide      `json:"side" gorm:"not null"`                       // buy=做多，sell=做空
	MarginAmount     float64        `json:"margin_amount" gorm:"type:decimal(20,8)"`     // 保证金本值（USDT）
	Price            float64        `json:"price" gorm:"type:decimal(20,8)"`             // 触发价格
	FloatBasisPoints float64        `json:"float_basis_points" gorm:"type:decimal(10,4);default:0.1"` // 首单万分比浮动（支持小数）
	TakeProfitBP     int            `json:"take_profit_bp" gorm:"default:0"`             // 止盈万分比
	StopLossBP       int            `json:"stop_loss_bp" gorm:"default:0"`               // 止损万分比
	Leverage         int            `json:"leverage" gorm:"default:8"`                   // 杠杆倍数，默认8倍
	MarginType       MarginType     `json:"margin_type" gorm:"default:'isolated'"`       // 默认逐仓
	Config           StrategyConfig `json:"config" gorm:"type:json"`
	State            StrategyState  `json:"state" gorm:"type:json"`
	IsActive         bool           `json:"is_active" gorm:"default:false"`
	IsCompleted      bool           `json:"is_completed" gorm:"default:false"`
	AutoRestart      bool           `json:"auto_restart" gorm:"default:false"`
	
	// 计算字段（不存储）
	OrderQuantity    float64        `json:"order_quantity" gorm:"-"`     // 实际下单数量
	EstimatedProfit  float64        `json:"estimated_profit" gorm:"-"`   // 预计盈利（USDT）
	EstimatedLoss    float64        `json:"estimated_loss" gorm:"-"`     // 预计亏损（USDT）
	LiquidationPrice float64        `json:"liquidation_price" gorm:"-"`  // 预计爆仓价格

	User          User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	FuturesOrders []FuturesOrder `json:"futures_orders,omitempty" gorm:"foreignKey:StrategyID"`
}

func (fs *FuturesStrategy) TableName() string {
	return "futures_strategies"
}

type FuturesOrder struct {
	BaseModel
	UserID             uint         `json:"user_id" gorm:"not null;index"`
	StrategyID         *uint        `json:"strategy_id" gorm:"index"`
	Symbol             string       `json:"symbol" gorm:"size:20;not null;index"`
	OrderID            string       `json:"order_id" gorm:"size:50;uniqueIndex"`
	ClientOrderID      string       `json:"client_order_id" gorm:"size:50"`
	Side               OrderSide    `json:"side" gorm:"not null"`
	PositionSide       PositionSide `json:"position_side" gorm:"not null"`
	Type               OrderType    `json:"type" gorm:"not null"`
	Quantity           float64      `json:"quantity" gorm:"type:decimal(20,8)"`
	Price              float64      `json:"price" gorm:"type:decimal(20,8)"`
	StopPrice          float64      `json:"stop_price" gorm:"type:decimal(20,8)"`
	Status             OrderStatus  `json:"status" gorm:"default:'pending'"`
	ExecutedQty        float64      `json:"executed_qty" gorm:"type:decimal(20,8);default:0"`
	CumulativeQuoteQty float64      `json:"cumulative_quote_qty" gorm:"type:decimal(20,8);default:0"`
	TimeInForce        string       `json:"time_in_force" gorm:"size:10"`
	ReduceOnly         bool         `json:"reduce_only" gorm:"default:false"`
	WorkingType        string       `json:"working_type" gorm:"size:20"`

	User     User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Strategy *FuturesStrategy `json:"strategy,omitempty" gorm:"foreignKey:StrategyID"`
}

func (fo *FuturesOrder) TableName() string {
	return "futures_orders"
}

type FuturesPosition struct {
	BaseModel
	UserID           uint         `json:"user_id" gorm:"not null;index"`
	Symbol           string       `json:"symbol" gorm:"size:20;not null;index"`
	PositionSide     PositionSide `json:"position_side" gorm:"not null"`
	PositionAmt      float64      `json:"position_amt" gorm:"type:decimal(20,8)"`
	EntryPrice       float64      `json:"entry_price" gorm:"type:decimal(20,8)"`
	MarkPrice        float64      `json:"mark_price" gorm:"type:decimal(20,8)"`
	UnRealizedProfit float64      `json:"unrealized_profit" gorm:"type:decimal(20,8)"`
	LiquidationPrice float64      `json:"liquidation_price" gorm:"type:decimal(20,8)"`
	Leverage         int          `json:"leverage"`
	MaxNotionalValue float64      `json:"max_notional_value" gorm:"type:decimal(20,8)"`
	MarginType       MarginType   `json:"margin_type"`
	IsolatedMargin   float64      `json:"isolated_margin" gorm:"type:decimal(20,8)"`
	IsAutoAddMargin  bool         `json:"is_auto_add_margin"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (fp *FuturesPosition) TableName() string {
	return "futures_positions"
}
