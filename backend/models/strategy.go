package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type StrategyConfig map[string]interface{}

func (c StrategyConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *StrategyConfig) Scan(value interface{}) error {
	if value == nil {
		*c = make(StrategyConfig)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return errors.New("cannot scan StrategyConfig")
	}
}

type StrategyState map[string]interface{}

func (s StrategyState) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StrategyState) Scan(value interface{}) error {
	if value == nil {
		*s = make(StrategyState)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan StrategyState")
	}
}

type Strategy struct {
	BaseModel
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	Name         string         `json:"name" gorm:"size:100;not null"`
	Symbol       string         `json:"symbol" gorm:"size:20;not null;index"`
	Type         StrategyType   `json:"type" gorm:"not null"`
	Side         OrderSide      `json:"side" gorm:"not null"`
	Quantity     float64        `json:"quantity" gorm:"type:decimal(20,8)"`
	Price        float64        `json:"price" gorm:"type:decimal(20,8)"`
	TriggerPrice float64        `json:"trigger_price" gorm:"type:decimal(20,8)"`
	StopPrice    float64        `json:"stop_price" gorm:"type:decimal(20,8)"`
	TakeProfit   float64        `json:"take_profit" gorm:"type:decimal(20,8)"`
	StopLoss     float64        `json:"stop_loss" gorm:"type:decimal(20,8)"`
	Config       StrategyConfig `json:"config" gorm:"type:json"`
	State        StrategyState  `json:"state" gorm:"type:json"`
	IsActive     bool           `json:"is_active" gorm:"default:false"`
	IsCompleted  bool           `json:"is_completed" gorm:"default:false"`
	AutoRestart  bool           `json:"auto_restart" gorm:"default:false"`

	User   User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Orders []Order `json:"orders,omitempty" gorm:"foreignKey:StrategyID"`
}

func (s *Strategy) TableName() string {
	return "strategies"
}

type Order struct {
	BaseModel
	UserID             uint        `json:"user_id" gorm:"not null;index"`
	StrategyID         *uint       `json:"strategy_id" gorm:"index"`
	Symbol             string      `json:"symbol" gorm:"size:20;not null;index"`
	OrderID            string      `json:"order_id" gorm:"size:50;uniqueIndex"`
	ClientOrderID      string      `json:"client_order_id" gorm:"size:50"`
	Side               OrderSide   `json:"side" gorm:"not null"`
	Type               OrderType   `json:"type" gorm:"not null"`
	Quantity           float64     `json:"quantity" gorm:"type:decimal(20,8)"`
	Price              float64     `json:"price" gorm:"type:decimal(20,8)"`
	StopPrice          float64     `json:"stop_price" gorm:"type:decimal(20,8)"`
	Status             OrderStatus `json:"status" gorm:"default:'pending'"`
	ExecutedQty        float64     `json:"executed_qty" gorm:"type:decimal(20,8);default:0"`
	CumulativeQuoteQty float64     `json:"cumulative_quote_qty" gorm:"type:decimal(20,8);default:0"`
	TimeInForce        string      `json:"time_in_force" gorm:"size:10"`
	IsWorking          bool        `json:"is_working" gorm:"default:true"`
	OrigQty            float64     `json:"orig_qty" gorm:"type:decimal(20,8)"`

	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Strategy *Strategy `json:"strategy,omitempty" gorm:"foreignKey:StrategyID"`
	Trades   []Trade   `json:"trades,omitempty" gorm:"foreignKey:OrderID;references:OrderID"`
}

func (o *Order) TableName() string {
	return "orders"
}

type Trade struct {
	BaseModel
	OrderID         string  `json:"order_id" gorm:"size:50;not null;index"`
	TradeID         string  `json:"trade_id" gorm:"size:50;uniqueIndex"`
	Symbol          string  `json:"symbol" gorm:"size:20;not null;index"`
	Price           float64 `json:"price" gorm:"type:decimal(20,8)"`
	Quantity        float64 `json:"quantity" gorm:"type:decimal(20,8)"`
	Commission      float64 `json:"commission" gorm:"type:decimal(20,8)"`
	CommissionAsset string  `json:"commission_asset" gorm:"size:10"`
	IsBuyer         bool    `json:"is_buyer"`
	IsMaker         bool    `json:"is_maker"`
	IsBestMatch     bool    `json:"is_best_match"`

	Order Order `json:"order,omitempty" gorm:"foreignKey:OrderID;references:OrderID"`
}

func (t *Trade) TableName() string {
	return "trades"
}

type Price struct {
	BaseModel
	Symbol string  `json:"symbol" gorm:"size:20;uniqueIndex;not null"`
	Price  float64 `json:"price" gorm:"type:decimal(20,8)"`
}

func (p *Price) TableName() string {
	return "prices"
}
