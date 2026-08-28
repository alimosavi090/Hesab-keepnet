package services

import (
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (p PageQuery) Normalized() (offset, limit int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
	return (p.Page - 1) * p.PageSize, p.PageSize
}

type Meta struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type ExpenseFilter struct {
	PageQuery
	Currency      enums.Currency
	CategoryID    *int64
	BankAccountID *int64
	Type          enums.CategoryType
	From, To      time.Time
}

type SaleFilter struct {
	PageQuery
	Currency enums.Currency
	Gateway  enums.Gateway
	From, To time.Time
}

type TransferFilter struct {
	PageQuery
	Currency enums.Currency
	From, To time.Time
}

type RepTransactionFilter struct {
	PageQuery
	From, To time.Time
}
