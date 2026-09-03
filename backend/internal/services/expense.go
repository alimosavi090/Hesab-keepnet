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

type ExpenseService struct {
	db    *gorm.DB
	audit repository.AuditRepository
}

type CreateExpenseInput struct {
	CategoryID    int64          `json:"category_id"`
	BankAccountID *int64         `json:"bank_account_id"`
	Amount        int64          `json:"amount"`
	Currency      enums.Currency `json:"currency"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Description   *string        `json:"description"`
}

func (s *ExpenseService) Create(ctx context.Context, in CreateExpenseInput) (*models.Expense, error) {
	if err := requirePositive(in.Amount); err != nil {
		return nil, err
	}
	if err := requireCurrency(in.Currency); err != nil {
		return nil, err
	}
	occurredAt := normalizeUTC(in.OccurredAt)

	var created *models.Expense

	err := database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var category models.Category
		if err := tx.First(&category, in.CategoryID).Error; err != nil {
			return apperr.Validation("دسته‌بندی هزینه یافت نشد (id=%d).", in.CategoryID)
		}
		if !category.IsActive {
			return apperr.Validation("دسته‌بندی %q غیرفعال است.", category.Name)
		}

		var account *models.BankAccount
		if in.BankAccountID != nil {
			account = &models.BankAccount{}
			if err := tx.First(account, *in.BankAccountID).Error; err != nil {
				return apperr.Validation("حساب بانکی هزینه یافت نشد (id=%d).", *in.BankAccountID)
			}
			if !account.IsActive {
				return apperr.Validation("حساب بانکی %q غیرفعال است.", account.Name)
			}
			if account.Currency != in.Currency {
				return apperr.Validation(
					"ارز حساب %q (%s) با ارز هزینه (%s) یکسان نیست.",
					account.Name, account.Currency, in.Currency,
				)
			}
		} else if category.Type == enums.CategoryBusiness {
			return apperr.Validation(
				"هزینه‌های کسب‌وکار (%s) باید به یک حساب بانکی متصل شوند.",
				category.Name,
			)
		}

		expense := models.Expense{
			CategoryID:  category.ID,
			Amount:      in.Amount,
			Currency:    in.Currency,
			OccurredAt:  occurredAt,
			Description: in.Description,
		}
		if account != nil {
			expense.BankAccountID = &account.ID
		}

		if err := tx.Create(&expense).Error; err != nil {
			return apperr.Database(err)
		}

		if account != nil {
			ledgerRow := models.LedgerTransaction{
				BankAccountID: account.ID,
				Type:          enums.LedgerExpense,
				Amount:        expense.Amount,
				Currency:      expense.Currency,
				OccurredAt:    expense.OccurredAt,
				Description:   expense.Description,
				ExpenseID:     &expense.ID,
			}
			if err := tx.Create(&ledgerRow).Error; err != nil {
				return apperr.Database(err)
			}
		}

		metadata := map[string]any{
			"amount":       in.Amount,
			"currency":     string(in.Currency),
			"category":     category.Name,
			"type":         string(category.Type),
			"bank_account": account != nil,
		}
		if err := writeAudit(s.audit, tx, ActionCreate, "expense", expense.ID, metadata); err != nil {
			return apperr.Database(err)
		}

		created = &expense
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *ExpenseService) Get(ctx context.Context, id int64) (*models.Expense, error) {
	var expense models.Expense
	if err := s.db.WithContext(ctx).Preload("Category").First(&expense, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	return &expense, nil
}

type UpdateExpenseInput struct {
	CategoryID  *int64     `json:"category_id"`
	Amount      *int64     `json:"amount"`
	OccurredAt  *time.Time `json:"occurred_at"`
	Description *string    `json:"description"`
}

// Update edits an expense and keeps its linked ledger row (bank-account
// side of double-entry) in sync with the new amount/date/description.
func (s *ExpenseService) Update(ctx context.Context, id int64, in UpdateExpenseInput) (*models.Expense, error) {
	if in.CategoryID == nil && in.Amount == nil && in.OccurredAt == nil && in.Description == nil {
		return nil, apperr.Validation("حداقل یک فیلد برای ویرایش لازم است.")
	}
	if in.Amount != nil {
		if err := requirePositive(*in.Amount); err != nil {
			return nil, err
		}
	}

	var expense models.Expense
	err := database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		if err := tx.First(&expense, id).Error; err != nil {
			return apperr.Normalize(err)
		}
		if in.CategoryID != nil {
			var category models.Category
			if err := tx.First(&category, *in.CategoryID).Error; err != nil {
				return apperr.Validation("دسته انتخابی یافت نشد.")
			}
			expense.CategoryID = category.ID
		}
		if in.Amount != nil {
			expense.Amount = *in.Amount
		}
		if in.OccurredAt != nil {
			expense.OccurredAt = normalizeUTC(*in.OccurredAt)
		}
		if in.Description != nil {
			desc := strings.TrimSpace(*in.Description)
			if desc == "" {
				expense.Description = nil
			} else {
				expense.Description = &desc
			}
		}
		if err := tx.Save(&expense).Error; err != nil {
			return apperr.Database(err)
		}

		// Mirror the change onto the ledger row so the linked bank
		// account balance stays correct.
		if expense.BankAccountID != nil {
			updates := map[string]any{
				"amount":      expense.Amount,
				"occurred_at": expense.OccurredAt,
			}
			if expense.Description != nil {
				updates["description"] = *expense.Description
			} else {
				updates["description"] = nil
			}
			if err := tx.Model(&models.LedgerTransaction{}).
				Where("expense_id = ? AND deleted_at IS NULL", expense.ID).
				Updates(updates).Error; err != nil {
				return apperr.Database(err)
			}
		}

		return writeAudit(s.audit, tx, ActionUpdate, "expense", expense.ID, map[string]any{
			"amount": expense.Amount,
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *ExpenseService) Delete(ctx context.Context, id int64) error {
	return database.WithImmediateTx(ctx, s.db, func(tx *gorm.DB) error {
		var expense models.Expense
		if err := tx.First(&expense, id).Error; err != nil {
			return apperr.Normalize(err)
		}

		if err := tx.Where("expense_id = ?", id).
			Delete(&models.LedgerTransaction{}).Error; err != nil {
			return apperr.Database(err)
		}
		if err := tx.Delete(&expense).Error; err != nil {
			return apperr.Database(err)
		}

		return writeAudit(s.audit, tx, ActionDelete, "expense", id, map[string]any{
			"amount":   expense.Amount,
			"currency": string(expense.Currency),
		})
	})
}

func (s *ExpenseService) List(ctx context.Context, filter ExpenseFilter) ([]models.Expense, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Expense{})
	applyTimeRange(query, "occurred_at", filter.From, filter.To)
	if filter.Currency != "" {
		query = query.Where("expenses.currency = ?", filter.Currency)
	}
	if filter.CategoryID != nil {
		query = query.Where("expenses.category_id = ?", *filter.CategoryID)
	}
	if filter.BankAccountID != nil {
		if *filter.BankAccountID == 0 {
			query = query.Where("expenses.bank_account_id IS NULL")
		} else {
			query = query.Where("expenses.bank_account_id = ?", *filter.BankAccountID)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}

	offset, limit := filter.Normalized()
	var expenses []models.Expense
	db := query.Preload("Category")
	if filter.Type != "" {
		db = db.Joins("JOIN categories ON categories.id = expenses.category_id AND categories.type = ?", string(filter.Type))
	}
	if err := db.Order("expenses.occurred_at DESC, expenses.id DESC").
		Limit(limit).Offset(offset).
		Find(&expenses).Error; err != nil {
		return nil, 0, apperr.Database(err)
	}
	return expenses, total, nil
}
