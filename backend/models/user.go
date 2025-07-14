package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	BaseModel
	Username    string     `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email       string     `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password    string     `json:"-" gorm:"size:255;not null"`
	Role        UserRole   `json:"role" gorm:"default:'user'"`
	Status      UserStatus `json:"status" gorm:"default:'pending'"`
	APIKey      string     `json:"-" gorm:"size:255"`
	SecretKey   string     `json:"-" gorm:"size:255"`
	IsEncrypted bool       `json:"is_encrypted" gorm:"default:false"`
	LastLoginAt *time.Time `json:"last_login_at"`
	HasAPIKey   bool       `json:"has_api_key" gorm:"-"`

	Strategies           []Strategy               `json:"strategies,omitempty" gorm:"foreignKey:UserID"`
	FuturesStrategies    []FuturesStrategy        `json:"futures_strategies,omitempty" gorm:"foreignKey:UserID"`
	DualInvestStrategies []DualInvestmentStrategy `json:"dual_invest_strategies,omitempty" gorm:"foreignKey:UserID"`
	Withdrawals          []Withdrawal             `json:"withdrawals,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func (u *User) TableName() string {
	return "users"
}
