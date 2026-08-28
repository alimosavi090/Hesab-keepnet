package models

import (
	"time"

	"gorm.io/gorm"
)

type NoteEntityType string

const (
	NoteEntityRepresentative NoteEntityType = "REPRESENTATIVE"
	NoteEntitySale           NoteEntityType = "SALE"
	NoteEntityBankAccount    NoteEntityType = "BANK_ACCOUNT"
	NoteEntityJournal        NoteEntityType = "JOURNAL"
)

type Note struct {
	ID         int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	EntityType NoteEntityType `json:"entity_type" gorm:"column:entity_type;not null"`
	EntityID   *int64         `json:"entity_id" gorm:"column:entity_id"`
	Body       string         `json:"body" gorm:"column:body;not null"`
	Tags       string         `json:"tags" gorm:"column:tags;not null;default:''"`
	Pinned     bool           `json:"pinned" gorm:"column:pinned;not null;default:false"`
	CreatedAt  time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}

func (Note) TableName() string { return "notes" }
