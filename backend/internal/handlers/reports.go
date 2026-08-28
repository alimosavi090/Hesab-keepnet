package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

func csvEscape(value string) string {
	if strings.ContainsAny(value, ",\"\n\r") {
		return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
	}
	return value
}

func writeCSV(c *gin.Context, filename string, headers []string, rows func(add func(cells []string))) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Status(http.StatusOK)

	c.Writer.WriteString("\uFEFF")
	c.Writer.WriteString(strings.Join(mapCSV(headers), ","))
	c.Writer.WriteString("\r\n")

	rows(func(cells []string) {
		c.Writer.WriteString(strings.Join(mapCSV(cells), ","))
		c.Writer.WriteString("\r\n")
	})
}

func mapCSV(cells []string) []string {
	out := make([]string, len(cells))
	for i, cell := range cells {
		out[i] = csvEscape(cell)
	}
	return out
}

func formatCSVTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04")
}

type datasetFilter struct {
	services.ExpenseFilter
	Dataset string
}

func (h *FinancialHandlers) ExportCSV(c *gin.Context) {
	dataset := c.DefaultQuery("dataset", "expenses")

	from, okFrom := parseTimeParam(c, "from")
	to, okTo := parseTimeParam(c, "to")
	if !okFrom || !okTo {
		return
	}
	if to.IsZero() {
		to = time.Now()
	}

	switch dataset {
	case "expenses":
		h.exportExpenses(c, from, to)
	case "sales":
		h.exportSales(c, from, to)
	case "transactions":
		h.exportTransactions(c, from, to)
	default:
		httpx.HandleError(c, apperr.Validation("نوع داده پشتیبانی نمی‌شود؛ مقادیر مجاز: expenses، sales، transactions"))
	}
}

func (h *FinancialHandlers) exportExpenses(c *gin.Context, from, to time.Time) {
	filter := services.ExpenseFilter{
		From:      from,
		To:        to.Add(24 * time.Hour),
		PageQuery: services.PageQuery{Page: 1, PageSize: services.MaxPageSize},
	}
	if raw := c.Query("currency"); raw != "" {
		currency := enums.Currency(raw)
		if currency.Valid() {
			filter.Currency = currency
		}
	}

	expenses, _, err := h.services.Expenses.List(c.Request.Context(), filter)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	writeCSV(c, "expenses.csv",
		[]string{"تاریخ", "دسته", "مبلغ", "ارز", "حساب بانکی", "توضیحات"},
		func(add func([]string)) {
			for _, e := range expenses {
				category := ""
				if e.Category != nil {
					category = e.Category.Name
				}
				account := "نقدی"
				if e.BankAccountID != nil {
					account = "-"
				}
				description := ""
				if e.Description != nil {
					description = *e.Description
				}
				add([]string{
					formatCSVTime(e.OccurredAt),
					category,
					strconvFormatInt(e.Amount),
					string(e.Currency),
					account,
					description,
				})
			}
		},
	)
}

func (h *FinancialHandlers) exportSales(c *gin.Context, from, to time.Time) {
	sales, _, err := h.services.Sales.List(c.Request.Context(), services.SaleFilter{
		From:      from,
		To:        to.Add(24 * time.Hour),
		PageQuery: services.PageQuery{Page: 1, PageSize: services.MaxPageSize},
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	writeCSV(c, "sales.csv",
		[]string{"تاریخ فروش", "مبلغ کل", "پرداخت‌شده", "وضعیت", "ارز", "خریدار"},
		func(add func([]string)) {
			for _, s := range sales {
				customer := ""
				if s.CustomerName != nil {
					customer = *s.CustomerName
				}
				add([]string{
					formatCSVTime(s.SoldAt),
					strconvFormatInt(s.TotalAmount),
					strconvFormatInt(s.PaidAmount),
					string(s.Status),
					string(s.Currency),
					customer,
				})
			}
		},
	)
}

func (h *FinancialHandlers) exportTransactions(c *gin.Context, from, to time.Time) {
	items, _, err := h.services.Reporting.LedgerFeed(c.Request.Context(), services.LedgerFilter{
		From:      from,
		To:        to.Add(24 * time.Hour),
		PageQuery: services.PageQuery{Page: 1, PageSize: services.MaxPageSize},
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	writeCSV(c, "transactions.csv",
		[]string{"تاریخ", "حساب", "نوع", "مبلغ", "ارز", "شرح"},
		func(add func([]string)) {
			for _, item := range items {
				description := ""
				if item.Description != nil {
					description = *item.Description
				}
				add([]string{
					formatCSVTime(item.OccurredAt),
					item.AccountName,
					item.Type,
					strconvFormatInt(item.Amount),
					string(item.Currency),
					description,
				})
			}
		},
	)
}

func strconvFormatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}
