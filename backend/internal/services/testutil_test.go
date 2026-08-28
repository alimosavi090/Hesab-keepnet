package services_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/ali/hesab-keepnet/backend/migrations"
)

type testEnv struct {
	DB       *database.DB
	Services *services.Services
	Ctx      context.Context
}

func openTestEnv(t *testing.T) *testEnv {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "core.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := database.Migrate(migrations.FS, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(context.Background(), db.DB, "", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}

	return &testEnv{
		DB:       db,
		Services: services.NewServices(db.DB),
		Ctx:      context.Background(),
	}
}

func (e *testEnv) createAccount(t *testing.T, name string, currency enums.Currency, initial int64) *models.BankAccount {
	t.Helper()
	account, err := e.Services.BankAccounts.Create(e.Ctx, services.CreateBankAccountInput{
		Name:           name,
		BankName:       "Test Bank",
		Currency:       currency,
		InitialBalance: initial,
	})
	if err != nil {
		t.Fatalf("create account %q: %v", name, err)
	}
	return account
}

func (e *testEnv) defaultCategory(t *testing.T, name string, categoryType enums.CategoryType) *models.Category {
	t.Helper()
	categories, err := e.Services.Categories.List(e.Ctx, &categoryType, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, category := range categories {
		if category.Name == name {
			return &category
		}
	}
	t.Fatalf("seeded category %q (%s) not found", name, categoryType)
	return nil
}
