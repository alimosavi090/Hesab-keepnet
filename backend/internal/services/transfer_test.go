package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"gorm.io/gorm"
)

func TestTransferHappyPath(t *testing.T) {
	env := openTestEnv(t)
	from := env.createAccount(t, "from", enums.CurrencyRIAL, 10_000_000)
	to := env.createAccount(t, "to", enums.CurrencyRIAL, 500_000)

	transfer, err := env.Services.Transfers.Create(env.Ctx, services.CreateTransferInput{
		FromAccountID: from.ID,
		ToAccountID:   to.ID,
		Amount:        1_000_000,
		Currency:      enums.CurrencyRIAL,
		TransferredAt: baseTime,
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	fromBalance, _ := env.Services.BankAccounts.Balance(env.Ctx, from.ID)
	toBalance, _ := env.Services.BankAccounts.Balance(env.Ctx, to.ID)

	if fromBalance.Balance != 9_000_000 {
		t.Errorf("from balance = %d, want 9000000", fromBalance.Balance)
	}
	if toBalance.Balance != 1_500_000 {
		t.Errorf("to balance = %d, want 1500000", toBalance.Balance)
	}
	if fromBalance.Outgoing != 1_000_000 || toBalance.Incoming != 1_000_000 {
		t.Errorf("flows wrong: out=%d in=%d", fromBalance.Outgoing, toBalance.Incoming)
	}

	rows := countRows(t, env, "SELECT COUNT(*) FROM transactions WHERE transfer_id = ?", transfer.ID)
	if rows != 2 {
		t.Errorf("transfer must produce exactly 2 ledger rows, got %d", rows)
	}

	profit, err := env.Services.Reporting.NetProfit(env.Ctx, baseTime.Add(-24*time.Hour), baseTime.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range profit {
		if row.Sales != 0 || row.BusinessExpense != 0 || row.NetProfit != 0 {
			t.Errorf("transfer leaked into profit report: %+v", row)
		}
	}
}

func TestTransferDeleteReversesBothSides(t *testing.T) {
	env := openTestEnv(t)
	from := env.createAccount(t, "from", enums.CurrencyRIAL, 100)
	to := env.createAccount(t, "to", enums.CurrencyRIAL, 0)

	transfer, err := env.Services.Transfers.Create(env.Ctx, services.CreateTransferInput{
		FromAccountID: from.ID, ToAccountID: to.ID,
		Amount: 60, Currency: enums.CurrencyRIAL, TransferredAt: baseTime,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := env.Services.Transfers.Delete(env.Ctx, transfer.ID); err != nil {
		t.Fatalf("delete transfer: %v", err)
	}

	fromBalance, _ := env.Services.BankAccounts.Balance(env.Ctx, from.ID)
	toBalance, _ := env.Services.BankAccounts.Balance(env.Ctx, to.ID)
	if fromBalance.Balance != 100 || toBalance.Balance != 0 {
		t.Errorf("after delete: from=%d to=%d, want 100/0", fromBalance.Balance, toBalance.Balance)
	}
}

func TestTransferValidationFailures(t *testing.T) {
	env := openTestEnv(t)
	tomanA := env.createAccount(t, "toman-a", enums.CurrencyRIAL, 0)
	tomanB := env.createAccount(t, "toman-b", enums.CurrencyRIAL, 0)
	usd := env.createAccount(t, "usd", enums.CurrencyUSD, 0)

	cases := []struct {
		name string
		in   services.CreateTransferInput
	}{
		{"same account", services.CreateTransferInput{
			FromAccountID: tomanA.ID, ToAccountID: tomanA.ID,
			Amount: 100, Currency: enums.CurrencyRIAL, TransferredAt: baseTime,
		}},
		{"cross currency accounts", services.CreateTransferInput{
			FromAccountID: tomanA.ID, ToAccountID: usd.ID,
			Amount: 100, Currency: enums.CurrencyRIAL, TransferredAt: baseTime,
		}},
		{"input currency mismatch", services.CreateTransferInput{
			FromAccountID: tomanA.ID, ToAccountID: tomanB.ID,
			Amount: 100, Currency: enums.CurrencyUSD, TransferredAt: baseTime,
		}},
		{"zero amount", services.CreateTransferInput{
			FromAccountID: tomanA.ID, ToAccountID: tomanB.ID,
			Amount: 0, Currency: enums.CurrencyRIAL, TransferredAt: baseTime,
		}},
		{"missing destination", services.CreateTransferInput{
			FromAccountID: tomanA.ID, ToAccountID: 99999,
			Amount: 100, Currency: enums.CurrencyRIAL, TransferredAt: baseTime,
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := env.Services.Transfers.Create(env.Ctx, tc.in); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	total := countRows(t, env, "SELECT COUNT(*) FROM transfers")
	if total != 0 {
		t.Errorf("failed transfers left %d rows behind", total)
	}
}

func TestTransferAtomicityRollsBackEverything(t *testing.T) {
	env := openTestEnv(t)
	from := env.createAccount(t, "from", enums.CurrencyRIAL, 1_000)
	to := env.createAccount(t, "to", enums.CurrencyRIAL, 0)

	services.AfterOutLedgerWrite = func(tx *gorm.DB) error {
		return errors.New("injected failure between ledger writes")
	}
	t.Cleanup(func() { services.AfterOutLedgerWrite = nil })

	_, err := env.Services.Transfers.Create(env.Ctx, services.CreateTransferInput{
		FromAccountID: from.ID, ToAccountID: to.ID,
		Amount: 400, Currency: enums.CurrencyRIAL, TransferredAt: baseTime,
	})
	if err == nil {
		t.Fatal("expected injected failure")
	}

	transfers := countRows(t, env, "SELECT COUNT(*) FROM transfers")
	ledgerOut := countRows(t, env, "SELECT COUNT(*) FROM transactions WHERE type = 'TRANSFER_OUT'")
	ledgerIn := countRows(t, env, "SELECT COUNT(*) FROM transactions WHERE type = 'TRANSFER_IN'")
	auditRows := countRows(t, env, "SELECT COUNT(*) FROM audit_logs")

	if transfers != 0 || ledgerOut != 0 || ledgerIn != 0 || auditRows != 0 {
		t.Fatalf("rollback incomplete: transfers=%d out=%d in=%d audit=%d",
			transfers, ledgerOut, ledgerIn, auditRows)
	}

	fromBalance, _ := env.Services.BankAccounts.Balance(env.Ctx, from.ID)
	if fromBalance.Balance != 1_000 {
		t.Errorf("from balance after rollback = %d, want 1000", fromBalance.Balance)
	}
}
