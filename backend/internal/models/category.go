package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
)

type Category struct {
	ID        int64              `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string             `json:"name" gorm:"column:name;not null"`
	Type      enums.CategoryType `json:"type" gorm:"column:type;not null"`
	ParentID  *int64             `json:"parent_id" gorm:"column:parent_id"`
	IsActive  bool               `json:"is_active" gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time          `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt time.Time          `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt gorm.DeletedAt     `json:"deleted_at" gorm:"column:deleted_at;index"`

	Parent *Category `gorm:"foreignKey:ParentID" json:"-"`
}

func (Category) TableName() string { return "categories" }
