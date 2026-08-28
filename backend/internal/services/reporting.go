package services

import (
	"context"
	"github.com/ali/hesab-keepnet/backend/internal/calendar"
	"strings"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"gorm.io/gorm"
)

type ReportingService struct {
	db *gorm.DB
}

type ProfitRow struct {
	Currency        enums.Currency `json:"currency"`
	Sales           int64          `json:"sales"`
	BusinessExpense int64          `json:"business_expense"`
	NetProfit       int64          `json:"net_profit"`
}

func (s *ReportingService) NetProfit(ctx context.Context, from, to time.Time) ([]ProfitRow, error) {
	from = from.UTC()
	to = to.UTC()

	type sumRow struct {
		Currency enums.Currency
		Total    int64
	}

	var salesRows []sumRow
	if err := s.db.WithContext(ctx).
		Model(&models.LedgerTransaction{}).
		Select("currency, COALESCE(SUM(amount),0) AS total").
		Where("type = ? AND deleted_at IS NULL AND occurred_at BETWEEN ? AND ?",
			enums.LedgerIncome, from, to).
		Group("currency").
		Scan(&salesRows).Error; err != nil {
		return nil, err
	}

	var expenseRows []sumRow
	if err := s.db.WithContext(ctx).
		Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Select("expenses.currency AS currency, COALESCE(SUM(expenses.amount),0) AS total").
		Where("categories.type = ? AND expenses.deleted_at IS NULL AND expenses.occurred_at BETWEEN ? AND ?",
			enums.CategoryBusiness, from, to).
		Group("expenses.currency").
		Scan(&expenseRows).Error; err != nil {
		return nil, err
	}

	rows := map[enums.Currency]*ProfitRow{}
	for _, r := range salesRows {
		rows[r.Currency] = &ProfitRow{Currency: r.Currency, Sales: r.Total}
	}
	for _, r := range expenseRows {
		row, ok := rows[r.Currency]
		if !ok {
			row = &ProfitRow{Currency: r.Currency}
			rows[r.Currency] = row
		}
		row.BusinessExpense = r.Total
	}

	result := make([]ProfitRow, 0, len(rows))
	currencies := []enums.Currency{enums.CurrencyRIAL, enums.CurrencyUSD}
	for _, currency := range currencies {
		row, ok := rows[currency]
		if !ok {
			continue
		}
		row.NetProfit = row.Sales - row.BusinessExpense
		result = append(result, *row)
	}
	return result, nil
}

func (s *ReportingService) AccountBalances(ctx context.Context, bankAccounts *BankAccountService) ([]AccountBalance, error) {
	accounts, err := bankAccounts.List(ctx, true)
	if err != nil {
		return nil, err
	}

	balances := make([]AccountBalance, 0, len(accounts))
	for _, account := range accounts {
		balance, err := bankAccounts.Balance(ctx, account.ID)
		if err != nil {
			return nil, err
		}
		balances = append(balances, *balance)
	}
	return balances, nil
}

type LedgerItem struct {
	ID            int64          `json:"id"`
	BankAccountID int64          `json:"bank_account_id"`
	AccountName   string         `json:"account_name"`
	Type          string         `json:"type"`
	Amount        int64          `json:"amount"`
	Currency      enums.Currency `json:"currency"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Description   *string        `json:"description"`
	SourceType    string         `json:"source_type"`
	SourceID      *int64         `json:"source_id"`
}

type LedgerFilter struct {
	PageQuery
	BankAccountID *int64
	Type          enums.LedgerType
	Currency      enums.Currency
	From, To      time.Time
}

func (s *ReportingService) LedgerFeed(ctx context.Context, filter LedgerFilter) ([]LedgerItem, int64, error) {
	base := s.db.WithContext(ctx).Table("transactions t").
		Joins("JOIN bank_accounts a ON a.id = t.bank_account_id").
		Where("t.deleted_at IS NULL")

	if filter.BankAccountID != nil {
		base = base.Where("t.bank_account_id = ?", *filter.BankAccountID)
	}
	if filter.Type != "" && filter.Type.Valid() {
		base = base.Where("t.type = ?", string(filter.Type))
	}
	if filter.Currency != "" {
		base = base.Where("t.currency = ?", filter.Currency)
	}
	if !filter.From.IsZero() {
		base = base.Where("t.occurred_at >= ?", filter.From.UTC())
	}
	if !filter.To.IsZero() {
		base = base.Where("t.occurred_at <= ?", filter.To.UTC())
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}

	offset, limit := filter.Normalized()
	var rows []struct {
		LedgerItem
		PaymentGateway *string `gorm:"column:payment_gateway"`
		CategoryName   *string `gorm:"column:category_name"`
	}
	err := base.
		Select(`t.id, a.name AS account_name, t.type, t.amount, t.currency,
			t.occurred_at, t.description,
			CASE
				WHEN t.sale_payment_id IS NOT NULL THEN 'sale_payment'
				WHEN t.expense_id IS NOT NULL THEN 'expense'
				ELSE 'transfer'
			END AS source_type,
			COALESCE(t.sale_payment_id, t.expense_id, t.transfer_id) AS source_id,
			sp.gateway AS payment_gateway,
			c.name AS category_name`).
		Joins("LEFT JOIN sale_payments sp ON sp.id = t.sale_payment_id").
		Joins("LEFT JOIN expenses e ON e.id = t.expense_id").
		Joins("LEFT JOIN categories c ON c.id = e.category_id").
		Order("t.occurred_at DESC, t.id DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, apperr.Database(err)
	}

	items := make([]LedgerItem, len(rows))
	for i, r := range rows {
		item := r.LedgerItem
		switch item.SourceType {
		case "expense":
			item.Description = orDefault(item.Description, "هزینه: "+valueOr(r.CategoryName, "-"))
		case "sale_payment":
			item.Description = orDefault(item.Description, "دریافت از درگاه "+valueOr(r.PaymentGateway, "-"))
		default:
			item.Description = orDefault(item.Description, "انتقال بین حساب‌ها")
		}
		items[i] = item
	}
	return items, total, nil
}

func applyTimeRange(query *gorm.DB, column string, from, to time.Time) {
	if !from.IsZero() {
		query = query.Where(column+" >= ?", from.UTC())
	}
	if !to.IsZero() {
		query = query.Where(column+" <= ?", to.UTC())
	}
}

func valueOr(s *string, fallback string) string {
	if s == nil || *s == "" {
		return fallback
	}
	return *s
}

func orDefault(s *string, fallback string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return &fallback
	}
	return s
}

type TotalRow struct {
	Currency enums.Currency `json:"currency"`
	Total    int64          `json:"total"`
}

type GatewayTotal struct {
	Gateway  enums.Gateway  `json:"gateway"`
	Currency enums.Currency `json:"currency"`
	Total    int64          `json:"total"`
}

type ExpenseSplit struct {
	Business []TotalRow `json:"business"`
	Personal []TotalRow `json:"personal"`
}

func (s *ReportingService) SalesByGateway(ctx context.Context, from, to time.Time) ([]GatewayTotal, error) {
	var rows []GatewayTotal
	err := s.db.WithContext(ctx).
		Model(&models.SalePayment{}).
		Select("gateway, currency, COALESCE(SUM(amount),0) AS total").
		Where("deleted_at IS NULL AND paid_at BETWEEN ? AND ?", from.UTC(), to.UTC()).
		Group("gateway, currency").
		Order("currency, total DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, apperr.Database(err)
	}
	return rows, nil
}

func splitTotals(ctx context.Context, db *gorm.DB, categoryType enums.CategoryType, from, to time.Time) ([]TotalRow, error) {
	var rows []TotalRow
	err := db.WithContext(ctx).
		Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Select("expenses.currency AS currency, COALESCE(SUM(expenses.amount),0) AS total").
		Where("categories.type = ? AND expenses.deleted_at IS NULL AND expenses.occurred_at BETWEEN ? AND ?",
			categoryType, from.UTC(), to.UTC()).
		Group("expenses.currency").
		Scan(&rows).Error
	if err != nil {
		return nil, apperr.Database(err)
	}
	return rows, nil
}

func (s *ReportingService) ExpenseSplit(ctx context.Context, from, to time.Time) (ExpenseSplit, error) {
	business, err := splitTotals(ctx, s.db, enums.CategoryBusiness, from, to)
	if err != nil {
		return ExpenseSplit{}, err
	}
	personal, err := splitTotals(ctx, s.db, enums.CategoryPersonal, from, to)
	if err != nil {
		return ExpenseSplit{}, err
	}
	if business == nil {
		business = []TotalRow{}
	}
	if personal == nil {
		personal = []TotalRow{}
	}
	return ExpenseSplit{Business: business, Personal: personal}, nil
}

type CategoryExpenseRow struct {
	CategoryName string             `json:"category_name"`
	CategoryType enums.CategoryType `json:"category_type"`
	Currency     enums.Currency     `json:"currency"`
	Total        int64              `json:"total"`
}

// ExpensesByCategory aggregates every expense (business AND personal) per
// category so the reports view can show a ranked breakdown.
func (s *ReportingService) ExpensesByCategory(ctx context.Context, from, to time.Time) ([]CategoryExpenseRow, error) {
	var rows []CategoryExpenseRow
	err := s.db.WithContext(ctx).
		Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Select("categories.name AS category_name, categories.type AS category_type, expenses.currency AS currency, COALESCE(SUM(expenses.amount),0) AS total").
		Where("expenses.deleted_at IS NULL AND expenses.occurred_at BETWEEN ? AND ?", from.UTC(), to.UTC()).
		Group("categories.name, categories.type, expenses.currency").
		Order("currency, total DESC").
		Limit(60).
		Scan(&rows).Error
	if err != nil {
		return nil, apperr.Database(err)
	}
	if rows == nil {
		rows = []CategoryExpenseRow{}
	}
	return rows, nil
}

type RepSettlementRow struct {
	Currency enums.Currency `json:"currency"`
	Total    int64          `json:"total"`
	Count    int64          `json:"count"`
}

// RepresentativeSettlements sums INCOME ledger rows that were created by
// representative debt settlements within the period.
func (s *ReportingService) RepresentativeSettlements(ctx context.Context, from, to time.Time) ([]RepSettlementRow, error) {
	var rows []RepSettlementRow
	err := s.db.WithContext(ctx).
		Model(&models.LedgerTransaction{}).
		Select("currency, COALESCE(SUM(amount),0) AS total, COUNT(*) AS count").
		Where("type = ? AND representative_transaction_id IS NOT NULL AND deleted_at IS NULL AND occurred_at BETWEEN ? AND ?",
			enums.LedgerIncome, from.UTC(), to.UTC()).
		Group("currency").
		Scan(&rows).Error
	if err != nil {
		return nil, apperr.Database(err)
	}
	return rows, nil
}

type DailyPoint struct {
	Date            string `json:"date"`
	Sales           int64  `json:"sales"`
	BusinessExpense int64  `json:"business_expense"`
}

func (s *ReportingService) DailySeries(ctx context.Context, today time.Time, days int, currency enums.Currency) ([]DailyPoint, error) {
	if days < 1 {
		days = 1
	}
	firstDay := today.AddDate(0, 0, -(days - 1))
	from, _ := calendar.DayRangeUTC(firstDay)
	_, to := calendar.DayRangeUTC(today)

	type rawRow struct {
		OccurredAt time.Time
		Amount     int64
	}

	var incomeRows []rawRow
	if err := s.db.WithContext(ctx).
		Model(&models.LedgerTransaction{}).
		Select("occurred_at, amount").
		Where("type = ? AND deleted_at IS NULL AND currency = ? AND occurred_at BETWEEN ? AND ?",
			enums.LedgerIncome, currency, from, to).
		Scan(&incomeRows).Error; err != nil {
		return nil, apperr.Database(err)
	}

	var expenseRows []rawRow
	if err := s.db.WithContext(ctx).
		Model(&models.Expense{}).
		Joins("JOIN categories ON categories.id = expenses.category_id").
		Select("expenses.occurred_at, expenses.amount").
		Where("categories.type = ? AND expenses.deleted_at IS NULL AND expenses.currency = ? AND expenses.occurred_at BETWEEN ? AND ?",
			enums.CategoryBusiness, currency, from, to).
		Scan(&expenseRows).Error; err != nil {
		return nil, apperr.Database(err)
	}

	incomeByDay := make(map[string]int64)
	for _, r := range incomeRows {
		key := r.OccurredAt.In(calendar.Location()).Format("2006-01-02")
		incomeByDay[key] += r.Amount
	}
	expenseByDay := make(map[string]int64)
	for _, r := range expenseRows {
		key := r.OccurredAt.In(calendar.Location()).Format("2006-01-02")
		expenseByDay[key] += r.Amount
	}

	points := make([]DailyPoint, 0, days)
	for i := 0; i < days; i++ {
		day := firstDay.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		points = append(points, DailyPoint{
			Date:            key,
			Sales:           incomeByDay[key],
			BusinessExpense: expenseByDay[key],
		})
	}
	return points, nil
}
