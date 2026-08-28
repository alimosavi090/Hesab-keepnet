package models

import (
	"time"

	"gorm.io/gorm"
)

const TableUsers = "users"

type User struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string         `json:"username" gorm:"column:username;not null"`
	PasswordHash string         `json:"password_hash" gorm:"column:password_hash;not null"`
	DisplayName  string         `json:"display_name" gorm:"column:display_name;not null;default:''"`
	Role         string         `json:"role" gorm:"column:role;not null;default:'ADMIN'"`
	IsActive     bool           `json:"is_active" gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at;not null"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;index"`
}

func (User) TableName() string { return TableUsers }
