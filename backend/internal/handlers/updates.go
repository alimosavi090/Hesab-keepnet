package handlers

import (
	"net/http"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ─── Edit endpoints (PATCH) ──────────────────────────────────────

type updateSaleRequest struct {
	TotalAmount  *int64     `json:"total_amount"`
	CustomerName *string    `json:"customer_name"`
	SoldAt       *timeValue `json:"sold_at"`
}

func (h *FinancialHandlers) UpdateSale(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	var in updateSaleRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی ویرایش فروش نامعتبر است."))
		return
	}

	input := services.UpdateSaleInput{CustomerName: in.CustomerName}
	if in.TotalAmount != nil {
		input.TotalAmount = in.TotalAmount
	}
	if in.SoldAt != nil {
		t := in.SoldAt.Time
		input.SoldAt = &t
	}

	sale, err := h.services.Sales.Update(c.Request.Context(), id, input)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, sale)
}

type updateExpenseRequest struct {
	CategoryID  *int64     `json:"category_id"`
	Amount      *int64     `json:"amount"`
	OccurredAt  *timeValue `json:"occurred_at"`
	Description *string    `json:"description"`
}

func (h *FinancialHandlers) UpdateExpense(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	var in updateExpenseRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی ویرایش هزینه نامعتبر است."))
		return
	}

	input := services.UpdateExpenseInput{
		CategoryID:  in.CategoryID,
		Amount:      in.Amount,
		Description: in.Description,
	}
	if in.OccurredAt != nil {
		t := in.OccurredAt.Time
		input.OccurredAt = &t
	}

	expense, err := h.services.Expenses.Update(c.Request.Context(), id, input)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, expense)
}

type updateBankAccountRequest struct {
	Name           *string `json:"name"`
	BankName       *string `json:"bank_name"`
	CardNumber     *string `json:"card_number"`
	Description    *string `json:"description"`
	InitialBalance *int64  `json:"initial_balance"`
}

func (h *FinancialHandlers) UpdateBankAccount(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	var in updateBankAccountRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی ویرایش حساب نامعتبر است."))
		return
	}

	account, err := h.services.BankAccounts.Update(c.Request.Context(), id, services.UpdateBankAccountInput{
		Name:           in.Name,
		BankName:       in.BankName,
		CardNumber:     in.CardNumber,
		Description:    in.Description,
		InitialBalance: in.InitialBalance,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, account)
}
