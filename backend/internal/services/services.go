package services

import (
	"encoding/json"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/repository"
	"gorm.io/gorm"
)

type Services struct {
	DB              *gorm.DB
	BankAccounts    *BankAccountService
	Categories      *CategoryService
	Sales           *SaleService
	Expenses        *ExpenseService
	Transfers       *TransferService
	Representatives *RepresentativeService
	Reminders       *ReminderService
	Reporting       *ReportingService
	Notes           *NotesService
	Backups         *BackupService
}

func NewServices(db *gorm.DB) *Services {
	return &Services{
		DB:              db,
		BankAccounts:    &BankAccountService{db},
		Categories:      &CategoryService{db, repository.NewCategoryRepository(db)},
		Sales:           &SaleService{db, repository.NewAuditRepository(db)},
		Expenses:        &ExpenseService{db, repository.NewAuditRepository(db)},
		Transfers:       &TransferService{db, repository.NewAuditRepository(db)},
		Representatives: &RepresentativeService{db, repository.NewAuditRepository(db)},
		Reminders:       &ReminderService{db},
		Reporting:       &ReportingService{db},
		Notes:           &NotesService{db, repository.NewAuditRepository(db)},
	}
}

const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)

func normalizeUTC(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
}

func requirePositive(amount int64) error {
	if amount <= 0 {
		return apperr.Validation("مبلغ باید بزرگ‌تر از صفر باشد.")
	}
	return nil
}

func requireCurrency(currency enums.Currency) error {
	if !currency.Valid() {
		return apperr.Validation("ارز نامعتبر است؛ فقط RIAL و USD مجاز است.")
	}
	return nil
}

func metadataJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
}

func writeAudit(
	repo repository.AuditRepository,
	tx *gorm.DB,
	action, entityType string,
	entityID int64,
	meta any,
) error {
	return repo.Log(tx, repository.AuditEntry{
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   metadataJSON(meta),
	})
}
