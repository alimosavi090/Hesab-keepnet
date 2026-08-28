package database

import (
	"context"
	"fmt"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/passwordhash"
	"gorm.io/gorm"
)

func Seed(ctx context.Context, gdb *gorm.DB, adminUsername, adminPassword string) error {
	return WithImmediateTx(ctx, gdb, func(tx *gorm.DB) error {
		if err := seedDefaultCategories(tx); err != nil {
			return fmt.Errorf("seed default categories: %w", err)
		}
		if adminUsername != "" && adminPassword != "" {
			if err := seedAdminUser(tx, adminUsername, adminPassword); err != nil {
				return fmt.Errorf("seed admin user: %w", err)
			}
		}
		return nil
	})
}

func seedDefaultCategories(tx *gorm.DB) error {
	for _, categoryType := range []enums.CategoryType{enums.CategoryBusiness, enums.CategoryPersonal} {
		names := append([]string(nil), enums.DefaultCategories[categoryType]...)

		for _, name := range names {
			var count int64
			if err := tx.Model(&models.Category{}).
				Where("name = ? AND type = ? AND parent_id IS NULL AND deleted_at IS NULL", name, categoryType).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			category := models.Category{Name: name, Type: categoryType, IsActive: true}
			if err := tx.Create(&category).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedAdminUser(tx *gorm.DB, username, password string) error {
	var count int64
	if err := tx.Unscoped().Model(&models.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := passwordhash.Hash(password)
	if err != nil {
		return err
	}

	admin := models.User{
		Username:     username,
		PasswordHash: hash,
		DisplayName:  "Administrator",
		Role:         "ADMIN",
		IsActive:     true,
	}
	if err := tx.Create(&admin).Error; err != nil {
		return err
	}
	return nil
}
