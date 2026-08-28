package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type createRepresentativeRequest struct {
	FullName     string         `json:"full_name" binding:"required"`
	Phone        string         `json:"phone" binding:"required"`
	Email        *string        `json:"email"`
	Address      *string        `json:"address"`
	NationalCode *string        `json:"national_code"`
	Currency     enums.Currency `json:"currency"`
	Notes        *string        `json:"notes"`
	StartDate    timeValue      `json:"start_date" binding:"required"`
}

func (h *FinancialHandlers) ListRepresentatives(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true" || c.Query("include_inactive") == "1"
	reps, err := h.services.Representatives.List(c.Request.Context(), includeInactive)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, reps)
}

func (h *FinancialHandlers) CreateRepresentative(c *gin.Context) {
	var in createRepresentativeRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ نام و شماره تماس الزامی است."))
		return
	}

	rep, err := h.services.Representatives.Create(c.Request.Context(), services.CreateRepresentativeInput{
		FullName:     in.FullName,
		Phone:        in.Phone,
		Email:        in.Email,
		Address:      in.Address,
		NationalCode: in.NationalCode,
		Currency:     in.Currency,
		Notes:        in.Notes,
		StartDate:    in.StartDate.Time,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, rep)
}

func (h *FinancialHandlers) GetRepresentativeBalance(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	balance, err := h.services.Representatives.Balance(c.Request.Context(), id)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, balance)
}

type createRepTransactionRequest struct {
	Direction     enums.RepDirection `json:"direction" binding:"required"`
	Amount        int64              `json:"amount" binding:"required"`
	OccurredAt    timeValue          `json:"occurred_at" binding:"required"`
	SaleID        *int64             `json:"sale_id"`
	BankAccountID *int64             `json:"bank_account_id"`
	Description   *string            `json:"description"`
}

func (h *FinancialHandlers) ListRepresentativeTransactions(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	pq, err := queryPage(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	filter := services.RepTransactionFilter{PageQuery: pq}

	var ok bool
	if filter.From, ok = parseTimeParam(c, "from"); !ok {
		return
	}
	if filter.To, ok = parseTimeParam(c, "to"); !ok {
		return
	}

	transactions, total, err := h.services.Representatives.ListTransactions(c.Request.Context(), id, filter)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	okPaged(c, transactions, total, pq)
}

func (h *FinancialHandlers) CreateRepresentativeTransaction(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	var in createRepTransactionRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ نوع، مبلغ و تاریخ الزامی هستند."))
		return
	}

	tx, err := h.services.Representatives.RecordTransaction(c.Request.Context(), id, services.RecordRepTransactionInput{
		Direction:     in.Direction,
		Amount:        in.Amount,
		OccurredAt:    in.OccurredAt.Time,
		SaleID:        in.SaleID,
		BankAccountID: in.BankAccountID,
		Description:   in.Description,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, tx)
}

func (h *FinancialHandlers) DeleteRepresentativeTransaction(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if err := h.services.Representatives.DeleteTransaction(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type reminderRequest struct {
	Title          string               `json:"title" binding:"required"`
	Description    *string              `json:"description"`
	DueDate        timeValue            `json:"due_date" binding:"required"`
	RepeatInterval enums.RepeatInterval `json:"repeat_interval"`
}

func (h *FinancialHandlers) ListReminders(c *gin.Context) {
	reminders, err := h.services.Reminders.List(c.Request.Context())
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, reminders)
}

func (h *FinancialHandlers) UpcomingReminders(c *gin.Context) {
	days := 7
	if raw := c.Query("days"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}
	reminders, err := h.services.Reminders.Upcoming(c.Request.Context(), time.Duration(days)*24*time.Hour)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, reminders)
}

func (h *FinancialHandlers) CreateReminder(c *gin.Context) {
	var in reminderRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.HandleError(c, apperr.Validation("ورودی نامعتبر است؛ عنوان و سررسید الزامی هستند."))
		return
	}

	reminder, err := h.services.Reminders.Create(c.Request.Context(), services.CreateReminderInput{
		Title:          in.Title,
		Description:    in.Description,
		DueDate:        in.DueDate.Time,
		RepeatInterval: in.RepeatInterval,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, reminder)
}

func (h *FinancialHandlers) UpdateReminderDone(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	var body struct {
		IsDone *bool `json:"is_done" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.IsDone == nil {
		httpx.HandleError(c, apperr.Validation("فیلد is_done الزامی است."))
		return
	}
	if err := h.services.Reminders.MarkDone(c.Request.Context(), id, *body.IsDone); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"id": id, "is_done": *body.IsDone})
}

func (h *FinancialHandlers) DeleteReminder(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	if err := h.services.Reminders.Delete(c.Request.Context(), id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
