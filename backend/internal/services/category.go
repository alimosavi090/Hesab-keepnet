package services

import (
	"context"
	"strings"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/repository"
	"gorm.io/gorm"
)

type CategoryService struct {
	db   *gorm.DB
	repo repository.CategoryRepository
}

type CreateCategoryInput struct {
	Name     string             `json:"name"`
	Type     enums.CategoryType `json:"type"`
	ParentID *int64             `json:"parent_id"`
}

func (s *CategoryService) Create(ctx context.Context, in CreateCategoryInput) (*models.Category, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, apperr.Validation("نام دسته‌بندی الزامی است.")
	}
	if !in.Type.Valid() {
		return nil, apperr.Validation("نوع دسته‌بندی نامعتبر است؛ فقط BUSINESS و PERSONAL مجاز است.")
	}

	if in.ParentID != nil {
		parent, err := s.Get(ctx, *in.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Type != in.Type {
			return nil, apperr.Validation(
				"نوع دسته فرزند باید با نوع دسته والد یکسان باشد (والد: %s، فرزند: %s).",
				parent.Type, in.Type,
			)
		}
		if !parent.IsActive {
			return nil, apperr.Validation("دسته والد غیرفعال است.")
		}
	}

	category := models.Category{Name: in.Name, Type: in.Type, ParentID: in.ParentID, IsActive: true}
	if err := s.db.WithContext(ctx).Create(&category).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return &category, nil
}

func (s *CategoryService) Get(ctx context.Context, id int64) (*models.Category, error) {
	var category models.Category
	if err := s.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return nil, apperr.Normalize(err)
	}
	return &category, nil
}

func (s *CategoryService) List(ctx context.Context, categoryType *enums.CategoryType, includeInactive bool) ([]models.Category, error) {
	query := s.db.WithContext(ctx)
	if categoryType != nil {
		query = query.Where("type = ?", *categoryType)
	}
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	var categories []models.Category
	if err := query.Order("type, id").Find(&categories).Error; err != nil {
		return nil, apperr.Database(err)
	}
	return categories, nil
}

func (s *CategoryService) SetActive(ctx context.Context, id int64, active bool) error {
	result := s.db.WithContext(ctx).Model(&models.Category{}).
		Where("id = ?", id).
		Update("is_active", active)
	if result.Error != nil {
		return apperr.Database(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("دسته‌بندی یافت نشد.")
	}
	return nil
}

func (s *CategoryService) HasExpenses(ctx context.Context, tx *gorm.DB, categoryID int64) (bool, error) {
	return s.repo.HasActiveExpenses(tx, categoryID)
}
