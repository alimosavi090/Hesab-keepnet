package services_test

import (
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

func TestSaleMultiGatewayPayments(t *testing.T) {
	env := openTestEnv(t)
	bankA := env.createAccount(t, "A", enums.CurrencyRIAL, 0)
	bankB := env.createAccount(t, "B", enums.CurrencyRIAL, 0)
	bankC := env.createAccount(t, "C", enums.CurrencyRIAL, 0)

	sale, err := env.Services.Sales.Create(env.Ctx, services.CreateSaleInput{
		TotalAmount: 800_000,
		Currency:    enums.CurrencyRIAL,
		SoldAt:      baseTime,
		Payments: []services.SalePaymentInput{
			{BankAccountID: bankA.ID, Gateway: enums.GatewayZarinpal, Amount: 500_000, PaidAt: baseTime},
			{BankAccountID: bankB.ID, Gateway: enums.GatewayCardToCard, Amount: 200_000, PaidAt: baseTime},
			{BankAccountID: bankC.ID, Gateway: enums.GatewaySupport, Amount: 100_000, PaidAt: baseTime},
		},
	})
	if err != nil {
		t.Fatalf("create sale: %v", err)
	}

	paid, err := env.Services.Sales.PaidAmount(env.Ctx, sale.ID)
	if err != nil || paid != 800_000 {
		t.Fatalf("paid = %d, %v; want 800000", paid, err)
	}
	if got := services.StatusOf(800_000, paid); got != models.SalePaid {
		t.Errorf("status = %s, want PAID", got)
	}

	for _, account := range []*models.BankAccount{bankA, bankB, bankC} {
		balance, err := env.Services.BankAccounts.Balance(env.Ctx, account.ID)
		if err != nil {
			t.Fatal(err)
		}
		if balance.Incoming == 0 {
			t.Errorf("account %d has no ledger income row", account.ID)
		}
	}
}

func TestSalePartialPaymentStatus(t *testing.T) {
	env := openTestEnv(t)
	account := env.createAccount(t, "A", enums.CurrencyRIAL, 0)

	sale, err := env.Services.Sales.Create(env.Ctx, services.CreateSaleInput{
		TotalAmount: 1_000_000,
		Currency:    enums.CurrencyRIAL,
		SoldAt:      baseTime,
		Payments: []services.SalePaymentInput{{
			BankAccountID: account.ID, Gateway: enums.GatewayZarinpal, Amount: 400_000, PaidAt: baseTime,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	paid, _ := env.Services.Sales.PaidAmount(env.Ctx, sale.ID)
	if got := services.StatusOf(sale.TotalAmount, paid); got != models.SalePartial {
		t.Errorf("status = %s, want PARTIAL", got)
	}
}

func TestSaleValidationFailures(t *testing.T) {
	env := openTestEnv(t)
	toman := env.createAccount(t, "toman", enums.CurrencyRIAL, 0)
	usd := env.createAccount(t, "usd", enums.CurrencyUSD, 0)
	categoryType := enums.CategoryBusiness
	_ = categoryType

	cases := []struct {
		name string
		in   services.CreateSaleInput
	}{
		{"zero total", services.CreateSaleInput{TotalAmount: 0, Currency: enums.CurrencyRIAL, SoldAt: baseTime}},
		{"invalid currency", services.CreateSaleInput{TotalAmount: 10, Currency: "EUR", SoldAt: baseTime}},
		{"bad gateway", services.CreateSaleInput{
			TotalAmount: 100, Currency: enums.CurrencyRIAL, SoldAt: baseTime,
			Payments: []services.SalePaymentInput{{BankAccountID: toman.ID, Gateway: "PAYPAL", Amount: 100, PaidAt: baseTime}},
		}},
		{"zero payment amount", services.CreateSaleInput{
			TotalAmount: 100, Currency: enums.CurrencyRIAL, SoldAt: baseTime,
			Payments: []services.SalePaymentInput{{BankAccountID: toman.ID, Gateway: enums.GatewaySupport, Amount: 0, PaidAt: baseTime}},
		}},
		{"payment on foreign-currency account", services.CreateSaleInput{
			TotalAmount: 100, Currency: enums.CurrencyRIAL, SoldAt: baseTime,
			Payments: []services.SalePaymentInput{{BankAccountID: usd.ID, Gateway: enums.GatewayZarinpal, Amount: 50, PaidAt: baseTime}},
		}},
		{"overpay beyond total", services.CreateSaleInput{
			TotalAmount: 100, Currency: enums.CurrencyRIAL, SoldAt: baseTime,
			Payments: []services.SalePaymentInput{
				{BankAccountID: toman.ID, Gateway: enums.GatewayZarinpal, Amount: 80, PaidAt: baseTime},
				{BankAccountID: toman.ID, Gateway: enums.GatewayZarinpal, Amount: 30, PaidAt: baseTime},
			},
		}},
		{"missing account", services.CreateSaleInput{
			TotalAmount: 100, Currency: enums.CurrencyRIAL, SoldAt: baseTime,
			Payments: []services.SalePaymentInput{{BankAccountID: 99999, Gateway: enums.GatewayZarinpal, Amount: 100, PaidAt: baseTime}},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := env.Services.Sales.Create(env.Ctx, tc.in); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}

	var saleCount, ledgerCount int64
	env.DB.DB.Raw("SELECT COUNT(*) FROM sales").Scan(&saleCount)
	env.DB.DB.Raw("SELECT COUNT(*) FROM transactions").Scan(&ledgerCount)
	if saleCount != 0 || ledgerCount != 0 {
		t.Errorf("failed creates left rows behind: sales=%d ledger=%d", saleCount, ledgerCount)
	}
}

func TestSaleDeleteRestoresBalanceAndKeepsRows(t *testing.T) {
	env := openTestEnv(t)
	account := env.createAccount(t, "A", enums.CurrencyRIAL, 1_000_000)

	sale, err := env.Services.Sales.Create(env.Ctx, services.CreateSaleInput{
		TotalAmount: 700_000,
		Currency:    enums.CurrencyRIAL,
		SoldAt:      baseTime,
		Payments: []services.SalePaymentInput{{
			BankAccountID: account.ID, Gateway: enums.GatewayCardToCard, Amount: 700_000, PaidAt: baseTime,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	before, _ := env.Services.BankAccounts.Balance(env.Ctx, account.ID)
	if before.Balance != 1_700_000 {
		t.Fatalf("balance before delete = %d", before.Balance)
	}

	if err := env.Services.Sales.Delete(env.Ctx, sale.ID); err != nil {
		t.Fatalf("delete sale: %v", err)
	}

	after, _ := env.Services.BankAccounts.Balance(env.Ctx, account.ID)
	if after.Balance != 1_000_000 {
		t.Errorf("balance after delete = %d, want initial 1000000", after.Balance)
	}

	var visibleSales, visibleLedger, allSales int64
	env.DB.DB.Raw("SELECT COUNT(*) FROM sales WHERE deleted_at IS NULL").Scan(&visibleSales)
	env.DB.DB.Raw("SELECT COUNT(*) FROM transactions WHERE deleted_at IS NULL").Scan(&visibleLedger)
	env.DB.DB.Unscoped().Model(&models.Sale{}).Count(&allSales)
	if visibleSales != 0 || visibleLedger != 0 {
		t.Errorf("soft-deleted rows still active: sales=%d ledger=%d", visibleSales, visibleLedger)
	}
	if allSales == 0 {
		t.Error("history rows must remain in database (no hard delete)")
	}

	paid, err := env.Services.Sales.PaidAmount(env.Ctx, sale.ID)
	if err != nil || paid != 0 {
		t.Errorf("paid amount after delete = %d, %v", paid, err)
	}
}
