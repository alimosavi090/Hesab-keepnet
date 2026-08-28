package repository

import (
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"gorm.io/gorm"
	"time"
)

type AuditRepository struct {
	DB *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return AuditRepository{DB: db}
}

type AuditEntry struct {
	ActorUserID *int64
	Action      string
	EntityType  string
	EntityID    int64
	Metadata    []byte
	IPAddress   string
	UserAgent   string
}

func (r AuditRepository) Log(tx *gorm.DB, entry AuditEntry) error {
	record := models.AuditLog{
		ID:          0,
		ActorUserID: entry.ActorUserID,
		Action:      entry.Action,
		EntityType:  entry.EntityType,
		EntityID:    entry.EntityID,
		CreatedAt:   time.Now().UTC(),
	}
	if len(entry.Metadata) > 0 {
		metadata := string(entry.Metadata)
		record.Metadata = &metadata
	}
	if entry.IPAddress != "" {
		ip := entry.IPAddress
		record.IPAddress = &ip
	}
	if entry.UserAgent != "" {
		ua := entry.UserAgent
		record.UserAgent = &ua
	}

	db := r.DB
	if tx != nil {
		db = tx
	}
	return db.Create(&record).Error
}

type CategoryRepository struct {
	DB *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return CategoryRepository{DB: db}
}

func (r CategoryRepository) HasActiveExpenses(tx *gorm.DB, categoryID int64) (bool, error) {
	db := r.DB
	if tx != nil {
		db = tx
	}
	var count int64
	err := db.Model(&models.Expense{}).
		Where("category_id = ?", categoryID).
		Count(&count).Error
	return count > 0, err
}
