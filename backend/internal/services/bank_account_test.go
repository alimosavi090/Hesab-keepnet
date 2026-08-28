package services_test

import (
	"testing"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

var baseTime = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

func TestBankAccountCreateAndList(t *testing.T) {
	env := openTestEnv(t)
	svc := env.Services.BankAccounts

	if _, err := svc.Create(env.Ctx, services.CreateBankAccountInput{Name: "", BankName: "X", Currency: enums.CurrencyRIAL}); err == nil {
		t.Error("empty name must be rejected")
	}
	if _, err := svc.Create(env.Ctx, services.CreateBankAccountInput{Name: "A", BankName: "X", Currency: "EUR"}); err == nil {
		t.Error("invalid currency must be rejected")
	}
	if _, err := svc.Create(env.Ctx, services.CreateBankAccountInput{Name: "A", BankName: "X", Currency: enums.CurrencyRIAL}); err != nil {
		t.Errorf("valid create failed: %v", err)
	}
	if _, err := svc.Create(env.Ctx, services.CreateBankAccountInput{Name: "A", BankName: "X", Currency: enums.CurrencyUSD}); err == nil {
		t.Error("duplicate active name must be rejected")
	}

	list, err := svc.List(env.Ctx, false)
	if err != nil || len(list) != 1 {
		t.Fatalf("list = %v, %v", list, err)
	}
}

func TestBankBalanceFormula(t *testing.T) {
	env := openTestEnv(t)
	svc := env.Services.BankAccounts
	account := env.createAccount(t, "main", enums.CurrencyRIAL, 10_000_000)

	business := env.defaultCategory(t, "سرور ایران", enums.CategoryBusiness)
	personal := env.defaultCategory(t, "کافه", enums.CategoryPersonal)

	if _, err := env.Services.Sales.Create(env.Ctx, services.CreateSaleInput{
		TotalAmount: 2_000_000,
		Currency:    enums.CurrencyRIAL,
		SoldAt:      baseTime,
		Payments: []services.SalePaymentInput{{
			BankAccountID: account.ID,
			Gateway:       enums.GatewayZarinpal,
			Amount:        2_000_000,
			PaidAt:        baseTime,
		}},
	}); err != nil {
		t.Fatalf("sale: %v", err)
	}

	if _, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID:    business.ID,
		BankAccountID: &account.ID,
		Amount:        500_000,
		Currency:      enums.CurrencyRIAL,
		OccurredAt:    baseTime,
	}); err != nil {
		t.Fatalf("business expense: %v", err)
	}

	if _, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID: personal.ID,
		Amount:     200_000,
		Currency:   enums.CurrencyRIAL,
		OccurredAt: baseTime,
	}); err != nil {
		t.Fatalf("personal cash expense: %v", err)
	}

	balance, err := svc.Balance(env.Ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balance.Incoming != 2_000_000 || balance.Outgoing != 500_000 {
		t.Errorf("incoming=%d outgoing=%d", balance.Incoming, balance.Outgoing)
	}
	if want := int64(11_500_000); balance.Balance != want {
		t.Errorf("balance = %d, want %d (initial + in - out)", balance.Balance, want)
	}
}

func TestPersonalBankExpenseAffectsBalanceNotProfit(t *testing.T) {
	env := openTestEnv(t)
	account := env.createAccount(t, "main", enums.CurrencyRIAL, 1_000_000)
	personal := env.defaultCategory(t, "کافه", enums.CategoryPersonal)

	if _, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID:    personal.ID,
		BankAccountID: &account.ID,
		Amount:        300_000,
		Currency:      enums.CurrencyRIAL,
		OccurredAt:    baseTime,
	}); err != nil {
		t.Fatalf("personal bank expense: %v", err)
	}

	balance, _ := env.Services.BankAccounts.Balance(env.Ctx, account.ID)
	if balance.Balance != 700_000 {
		t.Errorf("balance = %d, want 700000 (personal bank expense hits the account)", balance.Balance)
	}

	rows, err := env.Services.Reporting.NetProfit(env.Ctx, baseTime.Add(-24*time.Hour), baseTime.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.NetProfit != 0 || row.BusinessExpense != 0 {
			t.Errorf("personal expense leaked into profit report: %+v", row)
		}
	}
}

func TestSetActiveNotFound(t *testing.T) {
	env := openTestEnv(t)
	err := env.Services.BankAccounts.SetActive(env.Ctx, 99999, false)
	if err == nil {
		t.Fatal("expected not-found error for unknown account")
	}
}
