package models

type Withdrawal struct {
	BaseModel
	UserID       uint    `json:"user_id" gorm:"not null;index"`
	Asset        string  `json:"asset" gorm:"size:10;not null"`
	Address      string  `json:"address" gorm:"size:255;not null"`
	Network      string  `json:"network" gorm:"size:20"`
	Amount       float64 `json:"amount" gorm:"type:decimal(20,8)"`
	MinBalance   float64 `json:"min_balance" gorm:"type:decimal(20,8)"`
	TriggerPrice float64 `json:"trigger_price" gorm:"type:decimal(20,8)"`
	IsActive     bool    `json:"is_active" gorm:"default:false"`
	AutoWithdraw bool    `json:"auto_withdraw" gorm:"default:false"`
	Description  string  `json:"description" gorm:"size:255"`

	User                User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
	WithdrawalHistories []WithdrawalHistory `json:"withdrawal_histories,omitempty" gorm:"foreignKey:WithdrawalID"`
}

func (w *Withdrawal) TableName() string {
	return "withdrawals"
}

type WithdrawalHistory struct {
	BaseModel
	UserID       uint    `json:"user_id" gorm:"not null;index"`
	WithdrawalID *uint   `json:"withdrawal_id" gorm:"index"`
	Asset        string  `json:"asset" gorm:"size:10;not null"`
	Amount       float64 `json:"amount" gorm:"type:decimal(20,8)"`
	Fee          float64 `json:"fee" gorm:"type:decimal(20,8)"`
	Address      string  `json:"address" gorm:"size:255"`
	Network      string  `json:"network" gorm:"size:20"`
	TxID         string  `json:"tx_id" gorm:"size:255"`
	Status       string  `json:"status" gorm:"size:20"`
	ApplyTime    int64   `json:"apply_time"`
	CompleteTime int64   `json:"complete_time"`

	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Withdrawal *Withdrawal `json:"withdrawal,omitempty" gorm:"foreignKey:WithdrawalID"`
}

func (wh *WithdrawalHistory) TableName() string {
	return "withdrawal_histories"
}
