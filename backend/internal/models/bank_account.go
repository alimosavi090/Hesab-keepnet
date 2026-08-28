package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
)

type Session struct {
	TokenHash string    `json:"token_hash" gorm:"column:token_hash;primaryKey;size:64"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at;not null"`
	IPAddress string    `json:"ip_address" gorm:"column:ip_address"`
	UserAgent string    `json:"user_agent" gorm:"column:user_agent"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (Session) TableName() string { return "sessions" }

type BankAccount struct {
	ID             int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Name           string         `json:"name" gorm:"column:name;not null"`
	BankName       string         `json:"bank_name" gorm:"column:bank_name;not null"`
	CardNumber     *string        `json:"card_number" gorm:"column:card_number"`
	Currency       enums.Currency `json:"currency" gorm:"column:currency;not null"`
	InitialBalance int64          `json:"initial_balance" gorm:"column:initial_balance;not null;default:0"`
	Description    *string        `json:"description" gorm:"column:description"`
	IsActive       bool           `json:"is_active" gorm:"column:is_active;not null;default:true"`
	CreatedAt      time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}

func (BankAccount) TableName() string { return "bank_accounts" }
