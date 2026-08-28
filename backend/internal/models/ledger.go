package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
)

type Transfer struct {
	ID            int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	FromAccountID int64          `json:"from_account_id" gorm:"column:from_account_id;not null"`
	ToAccountID   int64          `json:"to_account_id" gorm:"column:to_account_id;not null"`
	Amount        int64          `json:"amount" gorm:"column:amount;not null"`
	Currency      enums.Currency `json:"currency" gorm:"column:currency;not null"`
	TransferredAt time.Time      `json:"transferred_at" gorm:"column:transferred_at;not null"`
	Description   *string        `json:"description" gorm:"column:description"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`

	FromAccount *BankAccount `gorm:"foreignKey:FromAccountID" json:"-"`
	ToAccount   *BankAccount `gorm:"foreignKey:ToAccountID" json:"-"`
}

func (Transfer) TableName() string { return "transfers" }

type LedgerTransaction struct {
	ID            int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	BankAccountID int64            `json:"bank_account_id" gorm:"column:bank_account_id;not null"`
	Type          enums.LedgerType `json:"type" gorm:"column:type;not null"`
	Amount        int64            `json:"amount" gorm:"column:amount;not null"`
	Currency      enums.Currency   `json:"currency" gorm:"column:currency;not null"`
	OccurredAt    time.Time        `json:"occurred_at" gorm:"column:occurred_at;not null"`
	Description   *string          `json:"description" gorm:"column:description"`
	SalePaymentID              *int64 `json:"sale_payment_id" gorm:"column:sale_payment_id"`
	ExpenseID                  *int64 `json:"expense_id" gorm:"column:expense_id"`
	TransferID                 *int64 `json:"transfer_id" gorm:"column:transfer_id"`
	RepresentativeTransactionID *int64 `json:"representative_transaction_id" gorm:"column:representative_transaction_id"`
	CreatedAt     time.Time        `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt     time.Time        `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt     gorm.DeletedAt   `json:"deleted_at" gorm:"column:deleted_at;index"`

	BankAccount *BankAccount `gorm:"foreignKey:BankAccountID" json:"-"`
}

func (LedgerTransaction) TableName() string { return "transactions" }
