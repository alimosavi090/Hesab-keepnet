package handlers

import (
	"net/http"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type salePaymentRequest struct {
	BankAccountID int64         `json:"bank_account_id" binding:"required"`
	Gateway       enums.Gateway `json:"gateway" binding:"required"`
	Amount        int64         `json:"amount" binding:"required"`
	PaidAt        timeValue     `json:"paid_at" binding:"required"`
	GatewayRef    *string       `json:"gateway_ref"`
	Description   *string       `json:"description"`
}

type createSaleRequest struct {
	TotalAmount  int64                `json:"total_amount" binding:"required"`
	Currency     enums.Currency       `json:"currency" binding:"required"`
	SoldAt       timeValue            `json:"sold_at" binding:"required"`
	CustomerName *string              `json:"customer_name"`
	Description  *string              `json:"description"`
	Payments     []salePaymentRequest `json:"payments"`
}

func toServicePayments(inputs []salePaymentRequest) []services.SalePaymentInput {
	payments := make([]services.SalePaymentInput, len(inputs))
	for i, p := range inputs {
		payments[i] = services.SalePaymentInput{
			BankAccountID: p.BankAccountID,
			Gateway:       p.Gateway,
			Amount:        p.Amount,
			PaidAt:        p.PaidAt.Time,
			GatewayRef:    p.GatewayRef,
			Description:   p.Description,
		}
	}
	return payments
}

func (h *FinancialHandlers) ListSales(c *gin.Context) {
	pq, err := queryPage(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	filter := services.SaleFilter{PageQuery: pq}

	var ok bool
	if filter.From, ok = parseTimeParam(c, "from"); !ok {
		return
	}
	if filter.To, ok = parseTimeParam(c, "to"); !ok {
		return
	}
	if raw := c.Query("currency"); raw != "" {
		filter.Currency = enums.Currency(raw)
	}
	if raw := c.Query("gateway"); raw != "" {
		filter.Gateway = enums.Gateway(raw)
	}

	sales, total, err := h.services.Sales.List(c.Request.Context(), filter)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	okPaged(c, sales, total, pq)
}

func (h *FinancialHandlers) GetSale(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	sale, err := h.services.Sales.Get(c.Request.Context(), id)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	paid, _ := h.services.Sales.PaidAmount(c.Request.Context(), id)
	httpx.OK(c, http.StatusOK, gin.H{
		"sale":        sale,
		"paid_amount": paid,
		"status":      services.StatusOf(sale.TotalAmount, paid),
	})
}

func (h *FinancialHandlers) CreateSale(c *gin.Context) {
	var in createSaleRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ مبلغ کل، ارز، تاریخ و پرداخت‌ها را بررسی کنید."))
		return
	}

	sale, err := h.services.Sales.Create(c.Request.Context(), services.CreateSaleInput{
		TotalAmount:  in.TotalAmount,
		Currency:     in.Currency,
		SoldAt:       in.SoldAt.Time,
		CustomerName: in.CustomerName,
		Description:  in.Description,
		Payments:     toServicePayments(in.Payments),
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, sale)
}

func (h *FinancialHandlers) DeleteSale(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if err := h.services.Sales.Delete(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type createTransferRequest struct {
	FromAccountID int64          `json:"from_account_id" binding:"required"`
	ToAccountID   int64          `json:"to_account_id" binding:"required"`
	Amount        int64          `json:"amount" binding:"required"`
	Currency      enums.Currency `json:"currency" binding:"required"`
	TransferredAt timeValue      `json:"transferred_at" binding:"required"`
	Description   *string        `json:"description"`
}

func (h *FinancialHandlers) ListTransfers(c *gin.Context) {
	pq, err := queryPage(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	filter := services.TransferFilter{PageQuery: pq}

	var ok bool
	if filter.From, ok = parseTimeParam(c, "from"); !ok {
		return
	}
	if filter.To, ok = parseTimeParam(c, "to"); !ok {
		return
	}
	if raw := c.Query("currency"); raw != "" {
		filter.Currency = enums.Currency(raw)
	}

	transfers, total, err := h.services.Transfers.List(c.Request.Context(), filter)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	okPaged(c, transfers, total, pq)
}

func (h *FinancialHandlers) CreateTransfer(c *gin.Context) {
	var in createTransferRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ حساب مبدأ/مقصد، مبلغ و ارز الزامی هستند."))
		return
	}

	transfer, err := h.services.Transfers.Create(c.Request.Context(), services.CreateTransferInput{
		FromAccountID: in.FromAccountID,
		ToAccountID:   in.ToAccountID,
		Amount:        in.Amount,
		Currency:      in.Currency,
		TransferredAt: in.TransferredAt.Time,
		Description:   in.Description,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, transfer)
}

func (h *FinancialHandlers) DeleteTransfer(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if err := h.services.Transfers.Delete(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
