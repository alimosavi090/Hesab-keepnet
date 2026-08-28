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

type TransferService struct {
	db    *gorm.DB
	audit repository.AuditRepository
}

var AfterOutLedgerWrite func(tx *gorm.DB) error

type CreateTransferInput struct {
	FromAccountID int64          `json:"from_account_id"`
	ToAccountID   int64          `json:"to_account_id"`
	Amount        int64          `json:"amount"`
	Currency      enums.Currency `json:"currency"`
	TransferredAt time.Time      `json:"transferred_at"`
	Description   *string        `json:"description"`
}

func (s *TransferService) Create(ctx context.Context, in CreateTransferInput) (*models.Transfer, error) {
	if err := requirePositive(in.Amount); err != nil {
		return nil, err
	}
	if err := requireCurrency(in.Currency); err != nil {
		return nil, err
	}
	if in.FromAccountID == in.ToAccountID {
		return nil, apperr.Validation("حساب مبدأ و مقصد انتقال نباید یکسان باشند.")
	}
	transferredAt := normalizeUTC(in.TransferredAt)

	var created *models.Transfer

	err := database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var from, to models.BankAccount
		if err := tx.First(&from, in.FromAccountID).Error; err != nil {
			return apperr.Validation("حساب مبدأ یافت نشد (id=%d).", in.FromAccountID)
		}
		if err := tx.First(&to, in.ToAccountID).Error; err != nil {
			return apperr.Validation("حساب مقصد یافت نشد (id=%d).", in.ToAccountID)
		}
		if !from.IsActive || !to.IsActive {
			return apperr.Validation("حساب‌های مبدأ و مقصد باید فعال باشند.")
		}

		if from.Currency != to.Currency {
			return apperr.Validation(
				"انتقال بین دو حساب با ارز متفاوت مجاز نیست (مبدأ: %s، مقصد: %s).",
				from.Currency, to.Currency,
			)
		}
		if from.Currency != in.Currency {
			return apperr.Validation(
				"ارز انتقال (%s) با ارز حساب‌ها (%s) یکسان نیست.",
				in.Currency, from.Currency,
			)
		}

		transfer := models.Transfer{
			FromAccountID: from.ID,
			ToAccountID:   to.ID,
			Amount:        in.Amount,
			Currency:      in.Currency,
			TransferredAt: transferredAt,
			Description:   in.Description,
		}
		if err := tx.Create(&transfer).Error; err != nil {
			return apperr.Database(err)
		}

		outRow := models.LedgerTransaction{
			BankAccountID: from.ID,
			Type:          enums.LedgerTransferOut,
			Amount:        transfer.Amount,
			Currency:      transfer.Currency,
			OccurredAt:    transfer.TransferredAt,
			Description:   transfer.Description,
			TransferID:    &transfer.ID,
		}
		inRow := outRow
		inRow.BankAccountID = to.ID
		inRow.Type = enums.LedgerTransferIn

		if err := tx.Create(&outRow).Error; err != nil {
			return apperr.Database(err)
		}
		if AfterOutLedgerWrite != nil {
			if err := AfterOutLedgerWrite(tx); err != nil {
				return apperr.Internal(err)
			}
		}
		if err := tx.Create(&inRow).Error; err != nil {
			return apperr.Database(err)
		}

		if err := writeAudit(s.audit, tx, ActionCreate, "transfer", transfer.ID, map[string]any{
			"amount":   transfer.Amount,
			"currency": string(transfer.Currency),
			"from":     from.ID,
			"to":       to.ID,
		}); err != nil {
			return apperr.Database(err)
		}

		created = &transfer
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *TransferService) Get(ctx context.Context, id int64) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := s.db.WithContext(ctx).First(&transfer, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	return &transfer, nil
}

func (s *TransferService) Delete(ctx context.Context, id int64) error {
	return database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var transfer models.Transfer
		if err := tx.First(&transfer, id).Error; err != nil {
			return apperr.Normalize(err)
		}

		if err := tx.Where("transfer_id = ?", id).
			Delete(&models.LedgerTransaction{}).Error; err != nil {
			return apperr.Database(err)
		}
		if err := tx.Delete(&transfer).Error; err != nil {
			return apperr.Database(err)
		}

		return writeAudit(s.audit, tx, ActionDelete, "transfer", id, map[string]any{
			"amount":   transfer.Amount,
			"currency": string(transfer.Currency),
		})
	})
}

func (s *TransferService) List(ctx context.Context, filter TransferFilter) ([]models.Transfer, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Transfer{})
	applyTimeRange(query, "transferred_at", filter.From, filter.To)
	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}

	offset, limit := filter.Normalized()
	var transfers []models.Transfer
	if err := query.Preload("FromAccount").Preload("ToAccount").
		Order("transferred_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&transfers).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}
	return transfers, total, nil
}
