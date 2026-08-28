package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
)

type Representative struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	FullName     string         `json:"full_name" gorm:"column:full_name;not null"`
	Phone        string         `json:"phone" gorm:"column:phone;not null"`
	Email        *string        `json:"email" gorm:"column:email"`
	Address      *string        `json:"address" gorm:"column:address"`
	NationalCode *string        `json:"national_code" gorm:"column:national_code"`
	Currency     enums.Currency `json:"currency" gorm:"column:currency;not null;default:'RIAL'"`
	Notes        *string        `json:"notes" gorm:"column:notes"`
	StartDate    time.Time      `json:"start_date" gorm:"column:start_date;not null"`
	IsActive     bool           `json:"is_active" gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}

func (Representative) TableName() string { return "representatives" }

type RepresentativeTransaction struct {
	ID                 int64              `json:"id" gorm:"primaryKey;autoIncrement"`
	RepresentativeID   int64              `json:"representative_id" gorm:"column:representative_id;not null"`
	Direction          enums.RepDirection `json:"direction" gorm:"column:direction;not null"`
	Amount             int64              `json:"amount" gorm:"column:amount;not null"`
	Currency           enums.Currency     `json:"currency" gorm:"column:currency;not null"`
	OccurredAt         time.Time          `json:"occurred_at" gorm:"column:occurred_at;not null"`
	SaleID             *int64             `json:"sale_id" gorm:"column:sale_id"`
	BankAccountID      *int64             `json:"bank_account_id" gorm:"column:bank_account_id"`
	LedgerTransactionID *int64            `json:"ledger_transaction_id" gorm:"column:ledger_transaction_id"`
	Description        *string            `json:"description" gorm:"column:description"`
	CreatedAt          time.Time          `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt          time.Time          `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt          gorm.DeletedAt     `json:"deleted_at" gorm:"column:deleted_at;index"`

	Representative *Representative `gorm:"foreignKey:RepresentativeID" json:"-"`
	BankAccount    *BankAccount    `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
}

func (RepresentativeTransaction) TableName() string { return "representative_transactions" }

type Reminder struct {
	ID             int64                `json:"id" gorm:"primaryKey;autoIncrement"`
	Title          string               `json:"title" gorm:"column:title;not null"`
	Description    *string              `json:"description" gorm:"column:description"`
	DueDate        time.Time            `json:"due_date" gorm:"column:due_date;not null"`
	RepeatInterval enums.RepeatInterval `json:"repeat_interval" gorm:"column:repeat_interval;not null;default:'NONE'"`
	IsDone         bool                 `json:"is_done" gorm:"column:is_done;not null;default:false"`
	CompletedAt    *time.Time           `json:"completed_at" gorm:"column:completed_at"`
	CreatedAt      time.Time            `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt      time.Time            `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt       `json:"deleted_at" gorm:"column:deleted_at;index"`
}

func (Reminder) TableName() string { return "reminders" }
