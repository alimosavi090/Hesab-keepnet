package handlers

import (
	"net/http"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

func maskCard(card *string) *string {
	if card == nil || len(*card) < 4 {
		return nil
	}
	value := *card
	masked := "•••• •••• •••• " + value[len(value)-4:]
	return &masked
}

func (h *FinancialHandlers) ListBankAccounts(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true" || c.Query("include_inactive") == "1"
	accounts, err := h.services.BankAccounts.List(c.Request.Context(), includeInactive)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	result := make([]gin.H, 0, len(accounts))
	for _, account := range accounts {
		balance, err := h.services.BankAccounts.Balance(c.Request.Context(), account.ID)
		if err != nil {
			httpx.HandleError(c, err)
			return
		}
		result = append(result, gin.H{
			"id":              account.ID,
			"name":            account.Name,
			"bank_name":       account.BankName,
			"card_number":     maskCard(account.CardNumber),
			"currency":        account.Currency,
			"initial_balance": account.InitialBalance,
			"description":     account.Description,
			"is_active":       account.IsActive,
			"balance":         balance,
		})
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *FinancialHandlers) CreateBankAccount(c *gin.Context) {
	var in services.CreateBankAccountInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ نام، نام بانک و ارز الزامی هستند."))
		return
	}
	account, err := h.services.BankAccounts.Create(c.Request.Context(), in)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, account)
}

func (h *FinancialHandlers) GetBankAccountBalance(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	balance, err := h.services.BankAccounts.Balance(c.Request.Context(), id)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, balance)
}

func (h *FinancialHandlers) UpdateBankAccountActive(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	var body struct {
		IsActive *bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.IsActive == nil {
		httpx.HandleError(c, apperr.Validation("فیلد is_active الزامی است."))
		return
	}
	if err := h.services.BankAccounts.SetActive(c.Request.Context(), id, *body.IsActive); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"id": id, "is_active": *body.IsActive})
}

func (h *FinancialHandlers) ListCategories(c *gin.Context) {
	var categoryType *enums.CategoryType
	if raw := c.Query("type"); raw != "" {
		t := enums.CategoryType(raw)
		categoryType = &t
	}
	categories, err := h.services.Categories.List(c.Request.Context(), categoryType, false)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, categories)
}

func (h *FinancialHandlers) CreateCategory(c *gin.Context) {
	var in services.CreateCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ نام و نوع دسته الزامی است."))
		return
	}
	category, err := h.services.Categories.Create(c.Request.Context(), in)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, category)
}

func (h *FinancialHandlers) DeactivateCategory(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if err := h.services.Categories.SetActive(c.Request.Context(), id, false); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
