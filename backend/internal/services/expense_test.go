package services_test

import (
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

func countRows(t *testing.T, env *testEnv, query string, args ...any) int64 {
	t.Helper()
	var count int64
	if err := env.DB.DB.Raw(query, args...).Scan(&count).Error; err != nil {
		t.Fatal(err)
	}
	return count
}

func TestExpenseBusinessRequiresBankAccount(t *testing.T) {
	env := openTestEnv(t)
	business := env.defaultCategory(t, "سرور ایران", enums.CategoryBusiness)

	_, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID: business.ID,
		Amount:     100_000,
		Currency:   enums.CurrencyRIAL,
		OccurredAt: baseTime,
	})
	if err == nil {
		t.Fatal("business expense without bank account must be rejected")
	}
}

func TestExpensePersonalWithoutBankHasNoLedgerRow(t *testing.T) {
	env := openTestEnv(t)
	personal := env.defaultCategory(t, "کافه", enums.CategoryPersonal)

	expense, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID:  personal.ID,
		Amount:      80_000,
		Currency:    enums.CurrencyRIAL,
		OccurredAt:  baseTime,
		Description: strPtr("قهوه"),
	})
	if err != nil {
		t.Fatalf("personal cash expense must succeed: %v", err)
	}

	ledgerCount := countRows(t, env, "SELECT COUNT(*) FROM transactions WHERE expense_id = ?", expense.ID)
	if ledgerCount != 0 {
		t.Errorf("cash personal expense created %d ledger rows, want 0", ledgerCount)
	}
}

func TestExpenseCurrencyMismatchRejected(t *testing.T) {
	env := openTestEnv(t)
	usdAccount := env.createAccount(t, "usd", enums.CurrencyUSD, 0)
	personal := env.defaultCategory(t, "کافه", enums.CategoryPersonal)

	_, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID:    personal.ID,
		BankAccountID: &usdAccount.ID,
		Amount:        10,
		Currency:      enums.CurrencyRIAL,
		OccurredAt:    baseTime,
	})
	var appErr *apperr.AppError
	if !asAppError(err, &appErr) || appErr.Code != apperr.CodeValidation {
		t.Fatalf("expected validation error for USD account + TOMAN expense, got %v", err)
	}
}

func TestExpenseInactiveCategoryRejected(t *testing.T) {
	env := openTestEnv(t)
	personal := env.defaultCategory(t, "کافه", enums.CategoryPersonal)
	if err := env.Services.Categories.SetActive(env.Ctx, personal.ID, false); err != nil {
		t.Fatal(err)
	}

	_, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID: personal.ID,
		Amount:     50_000,
		Currency:   enums.CurrencyRIAL,
		OccurredAt: baseTime,
	})
	if err == nil {
		t.Fatal("expense on inactive category must be rejected")
	}
}

func TestExpenseDeleteRestoresBalance(t *testing.T) {
	env := openTestEnv(t)
	account := env.createAccount(t, "main", enums.CurrencyRIAL, 5_000_000)
	business := env.defaultCategory(t, "هاست و دامنه", enums.CategoryBusiness)

	expense, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID:    business.ID,
		BankAccountID: &account.ID,
		Amount:        1_200_000,
		Currency:      enums.CurrencyRIAL,
		OccurredAt:    baseTime,
	})
	if err != nil {
		t.Fatal(err)
	}

	mid, _ := env.Services.BankAccounts.Balance(env.Ctx, account.ID)
	if mid.Balance != 3_800_000 {
		t.Fatalf("balance with expense = %d, want 3800000", mid.Balance)
	}

	if err := env.Services.Expenses.Delete(env.Ctx, expense.ID); err != nil {
		t.Fatalf("delete expense: %v", err)
	}

	final, _ := env.Services.BankAccounts.Balance(env.Ctx, account.ID)
	if final.Balance != 5_000_000 {
		t.Errorf("balance after delete = %d, want 5000000", final.Balance)
	}

	ledgerLeft := countRows(t, env, "SELECT COUNT(*) FROM transactions WHERE deleted_at IS NULL")
	if ledgerLeft != 0 {
		t.Errorf("ledger rows still active after expense delete: %d", ledgerLeft)
	}
}

func asAppError(err error, target **apperr.AppError) bool {
	if e, ok := err.(*apperr.AppError); ok {
		*target = e
		return true
	}
	return false
}

func strPtr(s string) *string { return &s }
