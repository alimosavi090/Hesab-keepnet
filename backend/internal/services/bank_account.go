package services

import (
	"context"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"gorm.io/gorm"
)

type BankAccountService struct {
	db *gorm.DB
}

type CreateBankAccountInput struct {
	Name           string         `json:"name"`
	BankName       string         `json:"bank_name"`
	CardNumber     *string        `json:"card_number"`
	Currency       enums.Currency `json:"currency"`
	InitialBalance int64          `json:"initial_balance"`
	Description    *string        `json:"description"`
}

func (s *BankAccountService) Create(ctx context.Context, in CreateBankAccountInput) (*models.BankAccount, error) {
	if in.Name == "" || in.BankName == "" {
		return nil, apperr.Validation("نام حساب و نام بانک الزامی است.")
	}
	if err := requireCurrency(in.Currency); err != nil {
		return nil, err
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.BankAccount{}).
		Where("name = ?", in.Name).
		Count(&count).Error; err != nil {
		return nil, apperr.Database(err)
	}
	if count > 0 {
		return nil, apperr.Conflict("حسابی با نام %q از قبل وجود دارد.", in.Name)
	}

	account := models.BankAccount{
		Name:           in.Name,
		BankName:       in.BankName,
		CardNumber:     in.CardNumber,
		Currency:       in.Currency,
		InitialBalance: in.InitialBalance,
		Description:    in.Description,
		IsActive:       true,
	}
	if err := s.db.WithContext(ctx).Create(&account).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return &account, nil
}

func (s *BankAccountService) Get(ctx context.Context, id int64) (*models.BankAccount, error) {
	var account models.BankAccount
	if err := s.db.WithContext(ctx).First(&account, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	return &account, nil
}

func (s *BankAccountService) List(ctx context.Context, includeInactive bool) ([]models.BankAccount, error) {
	query := s.db.WithContext(ctx)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	var accounts []models.BankAccount
	if err := query.Order("id").Find(&accounts).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return accounts, nil
}

func (s *BankAccountService) SetActive(ctx context.Context, id int64, active bool) error {
	result := s.db.WithContext(ctx).Model(&models.BankAccount{}).
		Where("id = ?", id).
		Update("is_active", active)
	if result.Error != nil {
		return apperr.Database(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("حساب بانکی یافت نشد.")
	}
	return nil
}

type AccountBalance struct {
	AccountID      int64          `json:"account_id"`
	Currency       enums.Currency `json:"currency"`
	InitialBalance int64          `json:"initial_balance"`
	Incoming       int64          `json:"incoming"`
	Outgoing       int64          `json:"outgoing"`
	Balance        int64          `json:"balance"`
}

func (s *BankAccountService) Balance(ctx context.Context, accountID int64) (*AccountBalance, error) {
	account, err := s.Get(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var net int64
	err = s.db.WithContext(ctx).
		Model(&models.LedgerTransaction{}).
		Select("COALESCE(SUM(CASE WHEN type IN ('INCOME','TRANSFER_IN') THEN amount ELSE -amount END), 0)").
		Where("bank_account_id = ? AND deleted_at IS NULL", accountID).
		Scan(&net).Error
	if err != nil {
		return nil, apperr.Database(err)
	}

	incoming, outgoing, _ := s.flows(ctx, accountID)

	return &AccountBalance{
		AccountID:      account.ID,
		Currency:       account.Currency,
		InitialBalance: account.InitialBalance,
		Incoming:       incoming,
		Outgoing:       outgoing,
		Balance:        account.InitialBalance + net,
	}, nil
}

func (s *BankAccountService) flows(ctx context.Context, accountID int64) (incoming, outgoing int64, err error) {
	type row struct {
		Type  enums.LedgerType
		Total int64
	}
	var rows []row
	err = s.db.WithContext(ctx).
		Model(&models.LedgerTransaction{}).
		Select("type, COALESCE(SUM(amount),0) AS total").
		Where("bank_account_id = ? AND deleted_at IS NULL", accountID).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, apperr.Database(err)
	}
	for _, r := range rows {
		if r.Type.IsOutflow() {
			outgoing += r.Total
		} else {
			incoming += r.Total
		}
	}
	return incoming, outgoing, nil
}
