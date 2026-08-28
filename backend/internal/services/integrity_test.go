package services_test

import (
	"testing"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

func TestFullFinancialScenario(t *testing.T) {
	env := openTestEnv(t)

	bankA := env.createAccount(t, "Bank A", enums.CurrencyRIAL, 10_000_000)
	bankB := env.createAccount(t, "Bank B", enums.CurrencyRIAL, 500_000)
	business := env.defaultCategory(t, "سرور ایران", enums.CategoryBusiness)
	personal := env.defaultCategory(t, "غذا بیرون", enums.CategoryPersonal)
	from := baseTime.Add(-24 * time.Hour)
	to := baseTime.Add(24 * time.Hour)

	sale, err := env.Services.Sales.Create(env.Ctx, services.CreateSaleInput{
		TotalAmount: 2_000_000,
		Currency:    enums.CurrencyRIAL,
		SoldAt:      baseTime,
		Payments: []services.SalePaymentInput{{
			BankAccountID: bankA.ID, Gateway: enums.GatewayZarinpal, Amount: 2_000_000, PaidAt: baseTime,
		}},
	})
	if err != nil {
		t.Fatalf("sale: %v", err)
	}

	expense, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID: business.ID, BankAccountID: &bankA.ID,
		Amount: 500_000, Currency: enums.CurrencyRIAL, OccurredAt: baseTime,
	})
	if err != nil {
		t.Fatalf("expense: %v", err)
	}

	if _, err := env.Services.Transfers.Create(env.Ctx, services.CreateTransferInput{
		FromAccountID: bankA.ID, ToAccountID: bankB.ID,
		Amount: 1_000_000, Currency: enums.CurrencyRIAL, TransferredAt: baseTime,
	}); err != nil {
		t.Fatalf("transfer: %v", err)
	}

	assertBalances(t, env, map[int64]int64{bankA.ID: 10_500_000, bankB.ID: 1_500_000})

	assertProfit(t, env, from, to, enums.CurrencyRIAL, services.ProfitRow{
		Currency:        enums.CurrencyRIAL,
		Sales:           2_000_000,
		BusinessExpense: 500_000,
		NetProfit:       1_500_000,
	})

	if _, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID: personal.ID, Amount: 200_000, Currency: enums.CurrencyRIAL, OccurredAt: baseTime,
	}); err != nil {
		t.Fatalf("personal cash expense: %v", err)
	}
	assertProfit(t, env, from, to, enums.CurrencyRIAL, services.ProfitRow{
		Currency: enums.CurrencyRIAL, Sales: 2_000_000, BusinessExpense: 500_000, NetProfit: 1_500_000,
	})

	if _, err := env.Services.Expenses.Create(env.Ctx, services.CreateExpenseInput{
		CategoryID: personal.ID, BankAccountID: &bankA.ID,
		Amount: 300_000, Currency: enums.CurrencyRIAL, OccurredAt: baseTime,
	}); err != nil {
		t.Fatalf("personal bank expense: %v", err)
	}
	assertBalances(t, env, map[int64]int64{bankA.ID: 10_200_000, bankB.ID: 1_500_000})
	assertProfit(t, env, from, to, enums.CurrencyRIAL, services.ProfitRow{
		Currency: enums.CurrencyRIAL, Sales: 2_000_000, BusinessExpense: 500_000, NetProfit: 1_500_000,
	})

	if _, err := env.Services.Sales.Create(env.Ctx, services.CreateSaleInput{
		TotalAmount: 50, Currency: enums.CurrencyUSD, SoldAt: baseTime,
		Payments: []services.SalePaymentInput{{
			BankAccountID: bankA.ID, Gateway: enums.GatewaySupport, Amount: 50, PaidAt: baseTime,
		}},
	}); err == nil {
		t.Error("USD sale payment on TOMAN account must be rejected")
	}

	profitAfterAll, _ := env.Services.Reporting.NetProfit(env.Ctx, from, to)
	for _, row := range profitAfterAll {
		if row.Currency == enums.CurrencyUSD && row.NetProfit != 0 {
			t.Errorf("rejected sale leaked into USD profit: %+v", row)
		}
	}

	if err := env.Services.Expenses.Delete(env.Ctx, expense.ID); err != nil {
		t.Fatal(err)
	}
	if err := env.Services.Sales.Delete(env.Ctx, sale.ID); err != nil {
		t.Fatal(err)
	}

	assertBalances(t, env, map[int64]int64{bankA.ID: 8_700_000, bankB.ID: 1_500_000})
	assertProfit(t, env, from, to, enums.CurrencyRIAL, services.ProfitRow{
		Currency: enums.CurrencyRIAL, Sales: 0, BusinessExpense: 0, NetProfit: 0,
	})
}

func assertBalances(t *testing.T, env *testEnv, want map[int64]int64) {
	t.Helper()
	for accountID, expected := range want {
		balance, err := env.Services.BankAccounts.Balance(env.Ctx, accountID)
		if err != nil {
			t.Fatal(err)
		}
		if balance.Balance != expected {
			t.Errorf("account %d balance = %d, want %d", accountID, balance.Balance, expected)
		}
	}
}

func assertProfit(t *testing.T, env *testEnv, from, to time.Time, currency enums.Currency, want services.ProfitRow) {
	t.Helper()
	rows, err := env.Services.Reporting.NetProfit(env.Ctx, from, to)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, row := range rows {
		if row.Currency == currency {
			found = true
			if row != want {
				t.Errorf("profit row = %+v, want %+v", row, want)
			}
		}
	}
	zeroExpected := want.Sales == 0 && want.BusinessExpense == 0 && want.NetProfit == 0
	if !found && !zeroExpected {
		t.Errorf("no profit row for %s; rows=%+v", currency, rows)
	}
}
