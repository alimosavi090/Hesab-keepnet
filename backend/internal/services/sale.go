package services

import (
	"context"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/repository"
	"gorm.io/gorm"
)

type SaleService struct {
	db    *gorm.DB
	audit repository.AuditRepository
}

type SalePaymentInput struct {
	BankAccountID int64         `json:"bank_account_id"`
	Gateway       enums.Gateway `json:"gateway"`
	Amount        int64         `json:"amount"`
	PaidAt        time.Time     `json:"paid_at"`
	GatewayRef    *string       `json:"gateway_ref"`
	Description   *string       `json:"description"`
}

type CreateSaleInput struct {
	TotalAmount  int64              `json:"total_amount"`
	Currency     enums.Currency     `json:"currency"`
	SoldAt       time.Time          `json:"sold_at"`
	CustomerName *string            `json:"customer_name"`
	Description  *string            `json:"description"`
	Payments     []SalePaymentInput `json:"payments"`
}

func (s *SaleService) Create(ctx context.Context, in CreateSaleInput) (*models.Sale, error) {
	if err := requirePositive(in.TotalAmount); err != nil {
		return nil, err
	}
	if err := requireCurrency(in.Currency); err != nil {
		return nil, err
	}
	soldAt := normalizeUTC(in.SoldAt)

	var created *models.Sale

	err := database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		payments, ledgerRows, err := s.buildPayments(tx, in.Currency, in.TotalAmount, in.Payments)
		if err != nil {
			return err
		}

		sale := models.Sale{
			TotalAmount:  in.TotalAmount,
			Currency:     in.Currency,
			SoldAt:       soldAt,
			CustomerName: in.CustomerName,
			Description:  in.Description,
		}
		if err := tx.Create(&sale).Error; err != nil {
			return apperr.Database(err)
		}

		for i := range payments {
			payments[i].SaleID = sale.ID
			if err := tx.Create(&payments[i]).Error; err != nil {
				return apperr.Database(err)
			}
			ledgerRows[i].SalePaymentID = &payments[i].ID
			if err := tx.Create(&ledgerRows[i]).Error; err != nil {
				return apperr.Database(err)
			}
		}

		if err := writeAudit(s.audit, tx, ActionCreate, "sale", sale.ID, map[string]any{
			"total_amount":  in.TotalAmount,
			"currency":      string(in.Currency),
			"payment_count": len(payments),
		}); err != nil {
			return apperr.Database(err)
		}

		created = &sale
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *SaleService) buildPayments(tx *gorm.DB, saleCurrency enums.Currency, totalAmount int64, inputs []SalePaymentInput) ([]models.SalePayment, []models.LedgerTransaction, error) {
	payments := make([]models.SalePayment, 0, len(inputs))
	ledgerRows := make([]models.LedgerTransaction, 0, len(inputs))
	var paidTotal int64

	for _, in := range inputs {
		if !in.Gateway.Valid() {
			return nil, nil, apperr.Validation("درگاه پرداخت نامعتبر است؛ مقادیر مجاز ZARINPAL، CARD_TO_CARD و SUPPORT هستند.")
		}
		if err := requirePositive(in.Amount); err != nil {
			return nil, nil, err
		}

		var account models.BankAccount
		if err := tx.First(&account, in.BankAccountID).Error; err != nil {
			return nil, nil, apperr.Validation("حساب بانکی پرداخت یافت نشد (id=%d).", in.BankAccountID)
		}
		if !account.IsActive {
			return nil, nil, apperr.Validation("حساب بانکی %q غیرفعال است.", account.Name)
		}
		if account.Currency != saleCurrency {
			return nil, nil, apperr.Validation(
				"ارز حساب %q (%s) با ارز فروش (%s) یکسان نیست.",
				account.Name, account.Currency, saleCurrency,
			)
		}

		paidTotal += in.Amount
		if paidTotal > totalAmount {
			return nil, nil, apperr.Validation(
				"مجموع پرداخت‌ها (%d) از مبلغ کل فروش (%d) بیشتر است.",
				paidTotal, totalAmount,
			)
		}

		payments = append(payments, models.SalePayment{
			BankAccountID: account.ID,
			Gateway:       in.Gateway,
			Amount:        in.Amount,
			Currency:      saleCurrency,
			PaidAt:        normalizeUTC(in.PaidAt),
			GatewayRef:    in.GatewayRef,
			Description:   in.Description,
		})
		ledgerRows = append(ledgerRows, models.LedgerTransaction{
			BankAccountID: account.ID,
			Type:          enums.LedgerIncome,
			Amount:        in.Amount,
			Currency:      saleCurrency,
			OccurredAt:    normalizeUTC(in.PaidAt),
			Description:   in.Description,
		})
	}

	return payments, ledgerRows, nil
}

func (s *SaleService) Get(ctx context.Context, id int64) (*models.Sale, error) {
	var sale models.Sale
	if err := s.db.WithContext(ctx).Preload("Payments").First(&sale, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	return &sale, nil
}

func (s *SaleService) Delete(ctx context.Context, id int64) error {
	return database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var sale models.Sale
		if err := tx.First(&sale, id).Error; err != nil {
			return apperr.Normalize(err)
		}

		var paymentIDs []int64
		if err := tx.Model(&models.SalePayment{}).
			Where("sale_id = ?", id).
			Pluck("id", &paymentIDs).Error; err != nil {
			return apperr.Database(err)
		}

		if len(paymentIDs) > 0 {
			if err := tx.Where("sale_payment_id IN ?", paymentIDs).
				Delete(&models.LedgerTransaction{}).Error; err != nil {
				return apperr.Database(err)
			}
		}
		if err := tx.Where("sale_id = ?", id).Delete(&models.SalePayment{}).Error; err != nil {
			return apperr.Database(err)
		}
		if err := tx.Delete(&sale).Error; err != nil {
			return apperr.Database(err)
		}

		return writeAudit(s.audit, tx, ActionDelete, "sale", id, map[string]any{
			"total_amount": sale.TotalAmount,
			"currency":     string(sale.Currency),
		})
	})
}

func (s *SaleService) PaidAmount(ctx context.Context, saleID int64) (int64, error) {
	var paid int64
	err := s.db.WithContext(ctx).
		Model(&models.SalePayment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("sale_id = ? AND deleted_at IS NULL", saleID).
		Scan(&paid).Error
	if err != nil {
		return 0, apperr.Database(err)
	}
	return paid, nil
}

func StatusOf(totalAmount, paidAmount int64) models.SaleStatus {
	switch {
	case paidAmount <= 0:
		return models.SaleUnpaid
	case paidAmount < totalAmount:
		return models.SalePartial
	default:
		return models.SalePaid
	}
}

type SaleListItem struct {
	models.Sale
	PaidAmount int64             `json:"paid_amount"`
	Status     models.SaleStatus `json:"status"`
}

func (s *SaleService) List(ctx context.Context, filter SaleFilter) ([]SaleListItem, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Sale{})
	applyTimeRange(query, "sold_at", filter.From, filter.To)
	if filter.Currency != "" {
		query = query.Where("sales.currency = ?", filter.Currency)
	}
	if filter.Gateway != "" {
		if !filter.Gateway.Valid() {
			return nil, 0, apperr.Validation("درگاه پرداخت نامعتبر است.")
		}
		query = query.Joins(
			"JOIN sale_payments sp ON sp.sale_id = sales.id AND sp.deleted_at IS NULL AND sp.gateway = ?",
			string(filter.Gateway),
		)
	}

	var total int64
	if err := query.Distinct("sales.id").Count(&total).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}

	offset, limit := filter.Normalized()
	var sales []models.Sale
	if err := query.Session(&gorm.Session{}).
		Distinct().
		Order("sales.sold_at DESC, sales.id DESC").
		Limit(limit).Offset(offset).
		Find(&sales).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}

	items := make([]SaleListItem, len(sales))
	paidBySale := s.paidTotals(ctx, saleIDs(sales))
	for i, sale := range sales {
		paid := paidBySale[sale.ID]
		items[i] = SaleListItem{Sale: sale, PaidAmount: paid, Status: StatusOf(sale.TotalAmount, paid)}
	}
	return items, total, nil
}

func (s *SaleService) paidTotals(ctx context.Context, ids []int64) map[int64]int64 {
	result := make(map[int64]int64, len(ids))
	if len(ids) == 0 {
		return result
	}
	type row struct {
		SaleID int64
		Total  int64
	}
	var rows []row
	_ = s.db.WithContext(ctx).Model(&models.SalePayment{}).
		Select("sale_id, COALESCE(SUM(amount),0) AS total").
		Where("sale_id IN ? AND deleted_at IS NULL", ids).
		Group("sale_id").Scan(&rows).Error
	for _, r := range rows {
		result[r.SaleID] = r.Total
	}
	return result
}

func saleIDs(sales []models.Sale) []int64 {
	ids := make([]int64, len(sales))
	for i, s := range sales {
		ids[i] = s.ID
	}
	return ids
}
