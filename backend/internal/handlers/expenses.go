package handlers

import (
	"net/http"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type createExpenseRequest struct {
	CategoryID    int64          `json:"category_id" binding:"required"`
	BankAccountID *int64         `json:"bank_account_id"`
	Amount        int64          `json:"amount" binding:"required"`
	Currency      enums.Currency `json:"currency" binding:"required"`
	OccurredAt    timeValue      `json:"occurred_at" binding:"required"`
	Description   *string        `json:"description"`
}

func (h *FinancialHandlers) ListExpenses(c *gin.Context) {
	pq, err := queryPage(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	filter := services.ExpenseFilter{PageQuery: pq}

	from, ok := parseTimeParam(c, "from")
	if !ok {
		return
	}
	filter.From = from

	to, ok := parseTimeParam(c, "to")
	if !ok {
		return
	}
	filter.To = to

	categoryID, ok := optionalID(c, "category_id")
	if !ok {
		return
	}
	filter.CategoryID = categoryID

	accountID, ok := optionalID(c, "bank_account_id")
	if !ok {
		return
	}
	filter.BankAccountID = accountID

	if raw := c.Query("currency"); raw != "" {
		currency := enums.Currency(raw)
		if !currency.Valid() {
			httpx.HandleError(c, apperr.Validation("ارز نامعتبر است."))
			return
		}
		filter.Currency = currency
	}

	if raw := c.Query("type"); raw != "" {
		t := enums.CategoryType(raw)
		if !t.Valid() {
			httpx.HandleError(c, apperr.Validation("نوع هزینه نامعتبر است."))
			return
		}
		filter.Type = t
	}

	expenses, total, err := h.services.Expenses.List(c.Request.Context(), filter)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	okPaged(c, expenses, total, pq)
}

func (h *FinancialHandlers) CreateExpense(c *gin.Context) {
	var in createExpenseRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ دسته، مبلغ، ارز و تاریخ الزامی هستند."))
		return
	}

	expense, err := h.services.Expenses.Create(c.Request.Context(), services.CreateExpenseInput{
		CategoryID:    in.CategoryID,
		BankAccountID: in.BankAccountID,
		Amount:        in.Amount,
		Currency:      in.Currency,
		OccurredAt:    in.OccurredAt.Time,
		Description:   in.Description,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, expense)
}

func (h *FinancialHandlers) DeleteExpense(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if err := h.services.Expenses.Delete(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
