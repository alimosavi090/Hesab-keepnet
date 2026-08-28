package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
)

type Sale struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	TotalAmount  int64          `json:"total_amount" gorm:"column:total_amount;not null"`
	Currency     enums.Currency `json:"currency" gorm:"column:currency;not null"`
	SoldAt       time.Time      `json:"sold_at" gorm:"column:sold_at;not null"`
	CustomerName *string        `json:"customer_name" gorm:"column:customer_name"`
	Description  *string        `json:"description" gorm:"column:description"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	Payments []SalePayment `gorm:"foreignKey:SaleID" json:"-"`
}

func (Sale) TableName() string { return "sales" }

type SalePayment struct {
	ID            int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	SaleID        int64          `json:"sale_id" gorm:"column:sale_id;not null"`
	BankAccountID int64          `json:"bank_account_id" gorm:"column:bank_account_id;not null"`
	Gateway       enums.Gateway  `json:"gateway" gorm:"column:gateway;not null"`
	Amount        int64          `json:"amount" gorm:"column:amount;not null"`
	Currency      enums.Currency `json:"currency" gorm:"column:currency;not null"`
	PaidAt        time.Time      `json:"paid_at" gorm:"column:paid_at;not null"`
	GatewayRef    *string        `json:"gateway_ref" gorm:"column:gateway_ref"`
	Description   *string        `json:"description" gorm:"column:description"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	BankAccount *BankAccount `gorm:"foreignKey:BankAccountID" json:"-"`
}

func (SalePayment) TableName() string { return "sale_payments" }

type SaleStatus string

const (
	SaleUnpaid  SaleStatus = "UNPAID"
	SalePartial SaleStatus = "PARTIAL"
	SalePaid    SaleStatus = "PAID"
)

type Expense struct {
	ID            int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	CategoryID    int64          `json:"category_id" gorm:"column:category_id;not null"`
	BankAccountID *int64         `json:"bank_account_id" gorm:"column:bank_account_id"`
	Amount        int64          `json:"amount" gorm:"column:amount;not null"`
	Currency      enums.Currency `json:"currency" gorm:"column:currency;not null"`
	OccurredAt    time.Time      `json:"occurred_at" gorm:"column:occurred_at;not null"`
	Description   *string        `json:"description" gorm:"column:description"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	Category    *Category    `gorm:"foreignKey:CategoryID" json:"category"`
	BankAccount *BankAccount `gorm:"foreignKey:BankAccountID" json:"bank_account"`
}

func (Expense) TableName() string { return "expenses" }
