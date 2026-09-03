package services_test

import (
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

func TestUpdateSaleEditsFields(t *testing.T) {
	env := openTestEnv(t)
	acct := env.createAccount(t, "فروش", enums.CurrencyRIAL, 0)
	category := env.defaultCategory(t, "سرور ایران", enums.CategoryBusiness)

	sale, err := env.Services.Sales.Create(env.Ctx, services.CreateSaleInput{
		TotalAmount: 100_000, Currency: enums.CurrencyRIAL,
		SoldAt: baseTime,
		Payments: []services.SalePaymentInput{{
			BankAccountID: acct.ID, Gateway: enums.GatewayCardToCard, Amount: 50_000, PaidAt: baseTime,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = category

	newAmount := int64(250_000)
	updated, err := env.Services.Sales.Update(env.Ctx, sale.ID, services.UpdateSaleInput{
		TotalAmount:  &newAmount,
		CustomerName: strPtr("مشتری تست"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.TotalAmount != 250_000 || updated.CustomerName == nil || *updated.CustomerName != "مشتری تست" {
		t.Errorf("updated sale = %+v", updated)
	}
}

func TestUpdateExpenseSyncsLedgerRow(t *testing.T) {
	env := openTestEnv(t)
	acct := env.createAccount(t, "بانک", enums.CurrencyRIAL, 1_000_000)
	category := env.defaultCategory(t, "سرور ایران", enums.CategoryBusiness)

	expense, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID: category.ID, BankAccountID: &acct.ID,
		Amount: 300_000, Currency: enums.CurrencyRIAL, OccurredAt: baseTime,
	})
	if err != nil {
		t.Fatal(err)
	}

	newAmount := int64(450_000)
	if _, err := env.Services.Expenses.Update(env.Ctx, expense.ID, services.UpdateExpenseInput{
		Amount: &newAmount, Description: strPtr("بروزرسانی‌شده"),
	}); err != nil {
		t.Fatal(err)
	}

	// Ledger row must mirror the edit so the account balance stays truthful.
	if ledgerCount, err := countLedgerRows(env, acct.ID); err != nil {
		t.Fatal(err)
	} else if ledgerCount != 1 {
		t.Fatalf("ledger rows = %d, want 1", ledgerCount)
	}
	balance, _ := env.Services.BankAccounts.Balance(env.Ctx, acct.ID)
	if balance.Balance != 1_000_000-450_000 {
		t.Errorf("balance after expense update = %d, want %d", balance.Balance, 1_000_000-450_000)
	}
}

func TestUpdateBankAccountMetadata(t *testing.T) {
	env := openTestEnv(t)
	acct := env.createAccount(t, "ملت اصلی", enums.CurrencyRIAL, 500_000)
	env.createAccount(t, "دیگری", enums.CurrencyRIAL, 0)

	newName := "ملت تجاری"
	updated, err := env.Services.BankAccounts.Update(env.Ctx, acct.ID, services.UpdateBankAccountInput{
		Name: &newName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "ملت تجاری" {
		t.Errorf("name = %q", updated.Name)
	}

	// Duplicate name among actives must be rejected.
	if _, err := env.Services.BankAccounts.Update(env.Ctx, acct.ID, services.UpdateBankAccountInput{
		Name: strPtr("دیگری"),
	}); err == nil {
		t.Error("duplicate active name must be rejected")
	}
}