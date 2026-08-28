package services

import (
	"context"
	"strings"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/repository"
	"gorm.io/gorm"
)

type RepresentativeService struct {
	db    *gorm.DB
	audit repository.AuditRepository
}

type CreateRepresentativeInput struct {
	FullName     string         `json:"full_name"`
	Phone        string         `json:"phone"`
	Email        *string        `json:"email"`
	Address      *string        `json:"address"`
	NationalCode *string        `json:"national_code"`
	Currency     enums.Currency `json:"currency"`
	Notes        *string        `json:"notes"`
	StartDate    time.Time      `json:"start_date"`
}

func (s *RepresentativeService) Create(ctx context.Context, in CreateRepresentativeInput) (*models.Representative, error) {
	in.FullName = strings.TrimSpace(in.FullName)
	if in.FullName == "" || strings.TrimSpace(in.Phone) == "" {
		return nil, apperr.Validation("نام کامل و شماره تماس نماینده الزامی است.")
	}
	if !in.Currency.Valid() {
		in.Currency = enums.CurrencyRIAL
	}

	if in.NationalCode != nil && *in.NationalCode != "" {
		var count int64
		if err := s.db.WithContext(ctx).Model(&models.Representative{}).
			Where("national_code = ?", *in.NationalCode).
			Count(&count).Error; err != nil {
			return nil, apperr.Database(err)
		}
		if count > 0 {
			return nil, apperr.Conflict("نماینده‌ای با این کد ملی از قبل ثبت شده است.")
		}
	} else {
		in.NationalCode = nil
	}

	rep := models.Representative{
		FullName:     in.FullName,
		Phone:        strings.TrimSpace(in.Phone),
		Email:        in.Email,
		Address:      in.Address,
		NationalCode: in.NationalCode,
		Currency:     in.Currency,
		Notes:        in.Notes,
		StartDate:    normalizeUTC(in.StartDate),
		IsActive:     true,
	}
	if err := s.db.WithContext(ctx).Create(&rep).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return &rep, nil
}

func (s *RepresentativeService) Get(ctx context.Context, id int64) (*models.Representative, error) {
	var rep models.Representative
	if err := s.db.WithContext(ctx).First(&rep, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	return &rep, nil
}

func (s *RepresentativeService) List(ctx context.Context, includeInactive bool) ([]models.Representative, error) {
	query := s.db.WithContext(ctx)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	var reps []models.Representative
	if err := query.Order("id").Find(&reps).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return reps, nil
}

func (s *RepresentativeService) SetActive(ctx context.Context, id int64, active bool) error {
	result := s.db.WithContext(ctx).Model(&models.Representative{}).
		Where("id = ?", id).
		Update("is_active", active)
	if result.Error != nil {
		return apperr.Database(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("نماینده یافت نشد.")
	}
	return nil
}

type RecordRepTransactionInput struct {
	Direction    enums.RepDirection `json:"direction"`
	Amount       int64              `json:"amount"`
	OccurredAt   time.Time          `json:"occurred_at"`
	SaleID       *int64             `json:"sale_id"`
	BankAccountID *int64            `json:"bank_account_id"`
	Description  *string            `json:"description"`
}

func (s *RepresentativeService) RecordTransaction(ctx context.Context, representativeID int64, in RecordRepTransactionInput) (*models.RepresentativeTransaction, error) {
	if !in.Direction.Valid() {
		return nil, apperr.Validation("نوع تراکنش نماینده نامعتبر است؛ فقط DEBIT و CREDIT مجاز است.")
	}
	if err := requirePositive(in.Amount); err != nil {
		return nil, err
	}
	if in.Direction == enums.RepCredit && (in.BankAccountID == nil || *in.BankAccountID <= 0) {
		return nil, apperr.Validation("برای ثبت تسویه، انتخاب حساب مقصد الزامی است.")
	}
	if in.Direction == enums.RepDebit && in.BankAccountID != nil && *in.BankAccountID > 0 {
		return nil, apperr.Validation("ثبت بدهی به حساب بانکی متصل نمی‌شود؛ فیلد حساب مقصد را خالی بگذارید.")
	}

	var created *models.RepresentativeTransaction

	err := database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var rep models.Representative
		if err := tx.First(&rep, representativeID).Error; err != nil {
			return apperr.Validation("نماینده یافت نشد (id=%d).", representativeID)
		}
		if !rep.IsActive {
			return apperr.Validation("نماینده %q غیرفعال است.", rep.FullName)
		}

		repTx := models.RepresentativeTransaction{
			RepresentativeID: rep.ID,
			Direction:        in.Direction,
			Amount:           in.Amount,
			Currency:         rep.Currency,
			OccurredAt:       normalizeUTC(in.OccurredAt),
			SaleID:           in.SaleID,
			BankAccountID:    in.BankAccountID,
			Description:      in.Description,
		}

		if err := tx.Create(&repTx).Error; err != nil {
			return apperr.Database(err)
		}

		// A settlement (CREDIT) is real money: it lands on the chosen bank
		// account as INCOME so it raises that balance and counts as revenue.
		if in.Direction == enums.RepCredit {
			var acct models.BankAccount
			if err := tx.First(&acct, *in.BankAccountID).Error; err != nil {
				return apperr.Validation("حساب مقصد یافت نشد.")
			}
			if !acct.IsActive {
				return apperr.Validation("حساب مقصد غیرفعال است.")
			}
			if acct.Currency != rep.Currency {
				return apperr.Validation("ارز حساب مقصد با ارز دفتر نماینده یکسان نیست.")
			}

			description := buildRepLedgerDescription(rep.FullName, in.Description)
			ledgerRow := models.LedgerTransaction{
				BankAccountID:               acct.ID,
				Type:                        enums.LedgerIncome,
				Amount:                      in.Amount,
				Currency:                    acct.Currency,
				OccurredAt:                  repTx.OccurredAt,
				Description:                 &description,
				RepresentativeTransactionID: &repTx.ID,
			}
			if err := tx.Create(&ledgerRow).Error; err != nil {
				return apperr.Database(err)
			}
			if err := tx.Model(&models.RepresentativeTransaction{}).
				Where("id = ?", repTx.ID).
				Update("ledger_transaction_id", ledgerRow.ID).Error; err != nil {
				return apperr.Database(err)
			}
			repTx.LedgerTransactionID = &ledgerRow.ID
		}

		if err := writeAudit(s.audit, tx, ActionCreate, "representative_transaction", repTx.ID, map[string]any{
			"representative": rep.ID,
			"direction":      string(in.Direction),
			"amount":         in.Amount,
			"currency":       string(rep.Currency),
			"bank_account":   in.BankAccountID,
		}); err != nil {
			return apperr.Database(err)
		}

		created = &repTx
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func buildRepLedgerDescription(fullName string, detail *string) string {
	base := "تسویه بدهی نماینده " + fullName
	if detail != nil && strings.TrimSpace(*detail) != "" {
		return base + " — " + strings.TrimSpace(*detail)
	}
	return base
}

type RepresentativeBalance struct {
	Currency    enums.Currency `json:"currency"`
	TotalDebit  int64          `json:"total_debit"`
	TotalCredit int64          `json:"total_credit"`
	Balance     int64          `json:"balance"`
}

func (b RepresentativeBalance) OwedToBusiness() bool {
	return b.Balance > 0
}

func (s *RepresentativeService) Balance(ctx context.Context, representativeID int64) (*RepresentativeBalance, error) {
	rep, err := s.Get(ctx, representativeID)
	if err != nil {
		return nil, err
	}

	type row struct {
		Direction enums.RepDirection
		Total     int64
	}
	var rows []row
	if err := s.db.WithContext(ctx).
		Model(&models.RepresentativeTransaction{}).
		Select("direction, COALESCE(SUM(amount),0) AS total").
		Where("representative_id = ? AND deleted_at IS NULL", representativeID).
		Group("direction").
		Scan(&rows).Error; err != nil {
		return nil, apperr.Database(err)
	}

	balance := RepresentativeBalance{Currency: rep.Currency}
	for _, r := range rows {
		switch r.Direction {
		case enums.RepDebit:
			balance.TotalDebit = r.Total
		case enums.RepCredit:
			balance.TotalCredit = r.Total
		}
	}
	balance.Balance = balance.TotalDebit - balance.TotalCredit
	return &balance, nil
}

func (s *RepresentativeService) DeleteTransaction(ctx context.Context, transactionID int64) error {
	return database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var repTx models.RepresentativeTransaction
		if err := tx.First(&repTx, transactionID).Error; err != nil {
			return apperr.Normalize(err)
		}
		if repTx.LedgerTransactionID != nil {
			// Soft-delete the linked INCOME row so the bank balance reverses.
			if err := tx.Delete(&models.LedgerTransaction{}, *repTx.LedgerTransactionID).Error; err != nil {
				return apperr.Database(err)
			}
		}
		if err := tx.Delete(&repTx).Error; err != nil {
			return apperr.Database(err)
		}
		return writeAudit(s.audit, tx, ActionDelete, "representative_transaction", transactionID, map[string]any{
			"amount":     repTx.Amount,
			"direction":  string(repTx.Direction),
			"bank_account": repTx.BankAccountID,
		})
	})
}

type RepresentativeDebt struct {
	RepresentativeID int64          `json:"representative_id"`
	FullName         string         `json:"full_name"`
	Currency         enums.Currency `json:"currency"`
	Debt             int64          `json:"debt"`
}

// Debts returns representatives whose ledger balance is positive (they owe
// the business), sorted largest first. Read-only: never touches bank balances.
func (s *RepresentativeService) Debts(ctx context.Context) ([]RepresentativeDebt, error) {
	type row struct {
		ID       int64
		FullName string
		Currency enums.Currency
		Debt     int64
	}
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("representatives AS r").
		Select("r.id AS id, r.full_name AS full_name, r.currency AS currency, "+
			"COALESCE(SUM(CASE WHEN t.direction = 'DEBIT' THEN t.amount ELSE -t.amount END), 0) AS debt").
		Joins("JOIN representative_transactions t ON t.representative_id = r.id AND t.deleted_at IS NULL").
		Where("r.deleted_at IS NULL").
		Group("r.id, r.full_name, r.currency").
		Having("debt > 0").
		Order("debt DESC").
		Scan(&rows).Error; err != nil {
		return nil, apperr.Database(err)
	}

	result := make([]RepresentativeDebt, 0, len(rows))
	for _, r := range rows {
		result = append(result, RepresentativeDebt{
			RepresentativeID: r.ID,
			FullName:         r.FullName,
			Currency:         r.Currency,
			Debt:             r.Debt,
		})
	}
	return result, nil
}

func (s *RepresentativeService) ListTransactions(ctx context.Context, representativeID int64, filter RepTransactionFilter) ([]models.RepresentativeTransaction, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.RepresentativeTransaction{}).
		Where("representative_id = ?", representativeID)
	applyTimeRange(query, "occurred_at", filter.From, filter.To)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}

	offset, limit := filter.Normalized()
	var transactions []models.RepresentativeTransaction
	if err := query.Preload("BankAccount").Order("occurred_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}
	return transactions, total, nil
}
