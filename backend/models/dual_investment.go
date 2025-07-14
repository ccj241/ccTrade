package models

type DualInvestmentProduct struct {
	BaseModel
	ProductID      string  `json:"product_id" gorm:"size:50;uniqueIndex;not null"`
	ProductName    string  `json:"product_name" gorm:"size:200"`
	BaseAsset      string  `json:"base_asset" gorm:"size:10;not null"`
	QuoteAsset     string  `json:"quote_asset" gorm:"size:10;not null"`
	MinAmount      float64 `json:"min_amount" gorm:"type:decimal(20,8)"`
	MaxAmount      float64 `json:"max_amount" gorm:"type:decimal(20,8)"`
	Duration       int     `json:"duration"`
	SettlementDate string  `json:"settlement_date" gorm:"size:20"`
	DeliveryPrice  float64 `json:"delivery_price" gorm:"type:decimal(20,8)"`
	YieldRate      float64 `json:"yield_rate" gorm:"type:decimal(10,4)"`
	IsActive       bool    `json:"is_active" gorm:"default:true"`
}

func (dip *DualInvestmentProduct) TableName() string {
	return "dual_investment_products"
}

type DualInvestmentStrategy struct {
	BaseModel
	UserID         uint    `json:"user_id" gorm:"not null;index"`
	Name           string  `json:"name" gorm:"size:100;not null"`
	ProductID      string  `json:"product_id" gorm:"size:50;not null"`
	BaseAsset      string  `json:"base_asset" gorm:"size:10;not null"`
	QuoteAsset     string  `json:"quote_asset" gorm:"size:10;not null"`
	InvestmentType string  `json:"investment_type" gorm:"size:20;not null"` // single, auto_reinvest, ladder, price_trigger
	Amount         float64 `json:"amount" gorm:"type:decimal(20,8)"`
	TriggerPrice   float64 `json:"trigger_price" gorm:"type:decimal(20,8)"`
	MinYieldRate   float64 `json:"min_yield_rate" gorm:"type:decimal(10,4)"`
	AutoReinvest   bool    `json:"auto_reinvest" gorm:"default:false"`
	LadderSteps    int     `json:"ladder_steps" gorm:"default:1"`
	AmountPerStep  float64 `json:"amount_per_step" gorm:"type:decimal(20,8)"`
	IsActive       bool    `json:"is_active" gorm:"default:false"`

	User                 User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	DualInvestmentOrders []DualInvestmentOrder `json:"dual_investment_orders,omitempty" gorm:"foreignKey:StrategyID"`
}

func (dis *DualInvestmentStrategy) TableName() string {
	return "dual_investment_strategies"
}

type DualInvestmentOrder struct {
	BaseModel
	UserID         uint    `json:"user_id" gorm:"not null;index"`
	StrategyID     *uint   `json:"strategy_id" gorm:"index"`
	ProductID      string  `json:"product_id" gorm:"size:50;not null"`
	OrderID        string  `json:"order_id" gorm:"size:50;uniqueIndex"`
	Amount         float64 `json:"amount" gorm:"type:decimal(20,8)"`
	Currency       string  `json:"currency" gorm:"size:10"`
	YieldRate      float64 `json:"yield_rate" gorm:"type:decimal(10,4)"`
	Duration       int     `json:"duration"`
	SettlementDate string  `json:"settlement_date" gorm:"size:20"`
	Status         string  `json:"status" gorm:"size:20"`
	PurchaseTime   int64   `json:"purchase_time"`
	SettlementTime int64   `json:"settlement_time"`

	User     User                    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Strategy *DualInvestmentStrategy `json:"strategy,omitempty" gorm:"foreignKey:StrategyID"`
}

func (dio *DualInvestmentOrder) TableName() string {
	return "dual_investment_orders"
}
