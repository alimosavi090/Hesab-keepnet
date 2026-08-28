package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/calendar"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

func (h *FinancialHandlers) LedgerFeed(c *gin.Context) {
	pq, err := queryPage(c)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	filter := services.LedgerFilter{PageQuery: pq}

	var ok bool
	if filter.From, ok = parseTimeParam(c, "from"); !ok {
		return
	}
	if filter.To, ok = parseTimeParam(c, "to"); !ok {
		return
	}
	if filter.BankAccountID, ok = optionalID(c, "bank_account_id"); !ok {
		return
	}
	if raw := c.Query("type"); raw != "" {
		t := enums.LedgerType(raw)
		if !t.Valid() {
			httpx.HandleError(c, apperr.Validation("نوع تراکنش نامعتبر است."))
			return
		}
		filter.Type = t
	}
	if raw := c.Query("currency"); raw != "" {
		filter.Currency = enums.Currency(raw)
	}

	items, total, err := h.services.Reporting.LedgerFeed(c.Request.Context(), filter)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	okPaged(c, items, total, pq)
}

type dashboardSummary struct {
	Profit         []services.ProfitRow          `json:"profit"`
	SalesByGateway []services.GatewayTotal       `json:"sales_by_gateway"`
	Expenses       services.ExpenseSplit         `json:"expenses"`
	Banks          []services.AccountBalance     `json:"banks"`
	Recent         []services.LedgerItem         `json:"recent"`
	Reminders      []models.Reminder             `json:"reminders"`
	RepDebts       []services.RepresentativeDebt `json:"rep_debts"`
}

func (h *FinancialHandlers) DashboardSummary(c *gin.Context) {
	ctx := c.Request.Context()

	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()
	if raw := c.Query("from"); raw != "" {
		parsed, ok := parseTimeParam(c, "from")
		if !ok {
			return
		}
		from = parsed
	}
	if raw := c.Query("to"); raw != "" {
		parsed, ok := parseTimeParam(c, "to")
		if !ok {
			return
		}
		to = parsed
	}

	profit, err := h.services.Reporting.NetProfit(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	gatewayTotals, err := h.services.Reporting.SalesByGateway(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	expenseSplit, err := h.services.Reporting.ExpenseSplit(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	banks, err := h.services.Reporting.AccountBalances(ctx, h.services.BankAccounts)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	recent, _, err := h.services.Reporting.LedgerFeed(ctx, services.LedgerFilter{
		PageQuery: services.PageQuery{Page: 1, PageSize: 10},
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	reminders, err := h.services.Reminders.Upcoming(ctx, 7*24*time.Hour)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	repDebts, err := h.services.Representatives.Debts(ctx)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.OK(c, http.StatusOK, dashboardSummary{
		Profit:         profit,
		SalesByGateway: gatewayTotals,
		Expenses:       expenseSplit,
		Banks:          banks,
		Recent:         recent,
		Reminders:      reminders,
		RepDebts:       repDebts,
	})
}

type reportOverview struct {
	Profit             []services.ProfitRow           `json:"profit"`
	Gateways           []services.GatewayTotal        `json:"gateways"`
	ExpensesByCategory []services.CategoryExpenseRow  `json:"expenses_by_category"`
	RepSettlements     []services.RepSettlementRow    `json:"rep_settlements"`
	RepDebts           []services.RepresentativeDebt  `json:"rep_debts"`
	ExpenseSplit       services.ExpenseSplit          `json:"expense_split"`
}

// ReportOverview powers the advanced reports page: revenue vs both expense
// types, category breakdown, representative settlements and open debts.
func (h *FinancialHandlers) ReportOverview(c *gin.Context) {
	ctx := c.Request.Context()

	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()
	if raw := c.Query("from"); raw != "" {
		parsed, ok := parseTimeParam(c, "from")
		if !ok {
			return
		}
		from = parsed
	}
	if raw := c.Query("to"); raw != "" {
		parsed, ok := parseTimeParam(c, "to")
		if !ok {
			return
		}
		to = parsed
	}

	profit, err := h.services.Reporting.NetProfit(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	gateways, err := h.services.Reporting.SalesByGateway(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	categories, err := h.services.Reporting.ExpensesByCategory(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	settlements, err := h.services.Reporting.RepresentativeSettlements(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	expenseSplit, err := h.services.Reporting.ExpenseSplit(ctx, from, to)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	repDebts, err := h.services.Representatives.Debts(ctx)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.OK(c, http.StatusOK, reportOverview{
		Profit:             profit,
		Gateways:           gateways,
		ExpensesByCategory: categories,
		RepSettlements:     settlements,
		RepDebts:           repDebts,
		ExpenseSplit:       expenseSplit,
	})
}

type chartPoint struct {
	Date            string `json:"date"`
	Sales           int64  `json:"sales"`
	BusinessExpense int64  `json:"business_expense"`
}

func (h *FinancialHandlers) DashboardChart(c *gin.Context) {
	days := 30
	if raw := c.Query("days"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}
	currency := enums.CurrencyRIAL
	if raw := c.Query("currency"); raw != "" {
		currency = enums.Currency(raw)
		if !currency.Valid() {
			httpx.HandleError(c, apperr.Validation("ارز نامعتبر است."))
			return
		}
	}

	points, err := h.services.Reporting.DailySeries(
		c.Request.Context(), calendar.TehranToday(), days, currency,
	)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, points)
}
