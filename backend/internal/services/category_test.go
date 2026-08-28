package services_test

import (
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

func TestCategoryChildMustMatchParentType(t *testing.T) {
	env := openTestEnv(t)
	business := env.defaultCategory(t, "سرور ایران", enums.CategoryBusiness)

	_, err := env.Services.Categories.Create(env.Ctx, services.CreateCategoryInput{
		Name:     "بازی",
		Type:     enums.CategoryPersonal,
		ParentID: &business.ID,
	})
	var appErr *apperr.AppError
	if !asAppError(err, &appErr) || appErr.Code != apperr.CodeValidation {
		t.Fatalf("type mismatch must be validation error, got %v", err)
	}

	valid, err := env.Services.Categories.Create(env.Ctx, services.CreateCategoryInput{
		Name:     "سرور اروپا",
		Type:     enums.CategoryBusiness,
		ParentID: &business.ID,
	})
	if err != nil {
		t.Fatalf("same-type child must succeed: %v", err)
	}
	if valid.ParentID == nil || *valid.ParentID != business.ID {
		t.Errorf("child parent_id = %v", valid.ParentID)
	}
}

func TestSeedIdempotentAndComplete(t *testing.T) {
	env := openTestEnv(t)

	count := countRows(t, env, "SELECT COUNT(*) FROM categories WHERE parent_id IS NULL AND deleted_at IS NULL")
	want := len(enums.DefaultCategories[enums.CategoryBusiness]) + len(enums.DefaultCategories[enums.CategoryPersonal])
	if count != int64(want) {
		t.Fatalf("seeded categories = %d, want %d", count, want)
	}

	for _, categoryType := range []enums.CategoryType{enums.CategoryBusiness, enums.CategoryPersonal} {
		for _, name := range enums.DefaultCategories[categoryType] {
			found := false
			categories, _ := env.Services.Categories.List(env.Ctx, &categoryType, true)
			for _, c := range categories {
				if c.Name == name {
					found = true
				}
			}
			if !found {
				t.Errorf("missing seeded category %q (%s)", name, categoryType)
			}
		}
	}
}
