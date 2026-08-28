package services_test

import (
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/enums"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/services"
)

func int64Ptr(v int64) *int64 { return &v }

func countLedgerRows(env *testEnv, accountID int64) (int64, error) {
	var count int64
	err := env.DB.DB.Model(&models.LedgerTransaction{}).
		Where("bank_account_id = ? AND deleted_at IS NULL", accountID).
		Count(&count).Error
	return count, err
}

func TestRepresentativeLedgerDeterministic(t *testing.T) {
	env := openTestEnv(t)
	acct := env.createAccount(t, "تسویه", enums.CurrencyRIAL, 0)
	rep, err := env.Services.Representatives.Create(env.Ctx, services.CreateRepresentativeInput{
		FullName: "علی نماینده", Phone: "09120000001",
		Currency: enums.CurrencyRIAL, StartDate: baseTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	svc := env.Services.Representatives

	entries := []struct {
		direction enums.RepDirection
		amount    int64
	}{
		{enums.RepCredit, 1_000_000},
		{enums.RepDebit, 300_000},
		{enums.RepCredit, 500_000},
	}
	for _, e := range entries {
		in := services.RecordRepTransactionInput{
			Direction: e.direction, Amount: e.amount, OccurredAt: baseTime,
		}
		if e.direction == enums.RepCredit {
			in.BankAccountID = &acct.ID
		}
		if _, err := svc.RecordTransaction(env.Ctx, rep.ID, in); err != nil {
			t.Fatalf("record %v %d: %v", e.direction, e.amount, err)
		}
	}

	balance, err := svc.Balance(env.Ctx, rep.ID)
	if err != nil {
		t.Fatal(err)
	}
	if balance.TotalCredit != 1_500_000 {
		t.Errorf("total credit = %d, want 1500000", balance.TotalCredit)
	}
	if balance.TotalDebit != 300_000 {
		t.Errorf("total debit = %d, want 300000", balance.TotalDebit)
	}
	if want := int64(-1_200_000); balance.Balance != want {
		t.Errorf("balance = %d, want %d (debit - credit)", balance.Balance, want)
	}
	if balance.OwedToBusiness() {
		t.Error("negative balance must not be flagged as owed-to-business")
	}

	// Settlements must land on the chosen account as INCOME (raises balance,
	// counts as revenue); debts must not touch balances at all.
	if bankBalance, err := env.Services.BankAccounts.Balance(env.Ctx, acct.ID); err != nil {
		t.Fatal(err)
	} else if bankBalance.Balance != 1_500_000 {
		t.Errorf("bank balance after settlements = %d, want 1500000", bankBalance.Balance)
	}

	// Overpaid representative must NOT appear as debtor on the dashboard.
	debts, err := svc.Debts(env.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(debts) != 0 {
		t.Errorf("overpaid representative must not be listed as debtor, got %+v", debts)
	}

	rep2, err := env.Services.Representatives.Create(env.Ctx, services.CreateRepresentativeInput{
		FullName: "نماینده بدهکار", Phone: "09120000002",
		Currency: enums.CurrencyRIAL, StartDate: baseTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RecordTransaction(env.Ctx, rep2.ID, services.RecordRepTransactionInput{
		Direction: enums.RepDebit, Amount: 800_000, OccurredAt: baseTime,
	}); err != nil {
		t.Fatal(err)
	}
	debts, err = svc.Debts(env.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(debts) != 1 || debts[0].RepresentativeID != rep2.ID ||
		debts[0].FullName != "نماینده بدهکار" || debts[0].Debt != 800_000 ||
		debts[0].Currency != enums.CurrencyRIAL {
		t.Errorf("debts = %+v, want single debt of 800000 for rep %d", debts, rep2.ID)
	}
}

func TestRepresentativeTransactionValidation(t *testing.T) {
	env := openTestEnv(t)
	rep, _ := env.Services.Representatives.Create(env.Ctx, services.CreateRepresentativeInput{
		FullName: "ر", Phone: "09123",
		Currency: enums.CurrencyUSD, StartDate: baseTime,
	})

	if _, err := env.Services.Representatives.RecordTransaction(env.Ctx, rep.ID, services.RecordRepTransactionInput{
		Direction: "CHARGE", Amount: 100, OccurredAt: baseTime,
	}); err == nil {
		t.Error("invalid direction must be rejected")
	}
	if _, err := env.Services.Representatives.RecordTransaction(env.Ctx, rep.ID, services.RecordRepTransactionInput{
		Direction: enums.RepCredit, Amount: -5, OccurredAt: baseTime,
	}); err == nil {
		t.Error("negative amount must be rejected")
	}
	if _, err := env.Services.Representatives.RecordTransaction(env.Ctx, rep.ID, services.RecordRepTransactionInput{
		Direction: enums.RepDebit, Amount: 100, OccurredAt: baseTime,
		BankAccountID: int64Ptr(1),
	}); err == nil {
		t.Error("debit must reject destination account")
	}

	// Wrong currency between rep ledger and destination account is invalid.
	rialAcct := env.createAccount(t, "ریالی", enums.CurrencyRIAL, 0)
	if _, err := env.Services.Representatives.RecordTransaction(env.Ctx, rep.ID, services.RecordRepTransactionInput{
		Direction: enums.RepCredit, Amount: 100, OccurredAt: baseTime,
		BankAccountID: &rialAcct.ID,
	}); err == nil {
		t.Error("credit into mismatched-currency account must be rejected")
	}

	balance, _ := env.Services.Representatives.Balance(env.Ctx, rep.ID)
	if balance.TotalCredit != 0 || balance.TotalDebit != 0 {
		t.Errorf("failed records changed balance: %+v", balance)
	}
}

func TestRepresentativeDeleteTransactionRestoresBalance(t *testing.T) {
	env := openTestEnv(t)
	acct := env.createAccount(t, "تسویه", enums.CurrencyRIAL, 1_000_000)
	rep, _ := env.Services.Representatives.Create(env.Ctx, services.CreateRepresentativeInput{
		FullName: "حذف‌شدنی", Phone: "09124445566",
		Currency: enums.CurrencyRIAL, StartDate: baseTime,
	})

	creditTx, err := env.Services.Representatives.RecordTransaction(env.Ctx, rep.ID, services.RecordRepTransactionInput{
		Direction: enums.RepCredit, Amount: 400_000, OccurredAt: baseTime,
		BankAccountID: &acct.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	balanceAfterCredit, _ := env.Services.BankAccounts.Balance(env.Ctx, acct.ID)
	if balanceAfterCredit.Balance != 1_400_000 {
		t.Fatalf("bank balance after credit = %d, want 1400000", balanceAfterCredit.Balance)
	}

	// Deleting the settlement must reverse its ledger row and restore balance.
	if err := env.Services.Representatives.DeleteTransaction(env.Ctx, creditTx.ID); err != nil {
		t.Fatal(err)
	}
	balanceAfterDelete, _ := env.Services.BankAccounts.Balance(env.Ctx, acct.ID)
	if balanceAfterDelete.Balance != 1_000_000 {
		t.Errorf("bank balance after delete = %d, want 1000000", balanceAfterDelete.Balance)
	}
	if ledgerCount, err := countLedgerRows(env, acct.ID); err != nil {
		t.Fatal(err)
	} else if ledgerCount != 0 {
		t.Errorf("ledger rows after delete = %d, want 0", ledgerCount)
	}

	tx, err := env.Services.Representatives.RecordTransaction(env.Ctx, rep.ID, services.RecordRepTransactionInput{
		Direction: enums.RepDebit, Amount: 250_000, OccurredAt: baseTime,
	})
	if err != nil {
		t.Fatal(err)
	}

	balanceAfterAdd, _ := env.Services.Representatives.Balance(env.Ctx, rep.ID)
	if balanceAfterAdd.Balance != 250_000 {
		t.Fatalf("rep balance after debit = %d, want 250000", balanceAfterAdd.Balance)
	}

	if err := env.Services.Representatives.DeleteTransaction(env.Ctx, tx.ID); err != nil {
		t.Fatal(err)
	}

	balanceAfterDebitDelete, _ := env.Services.Representatives.Balance(env.Ctx, rep.ID)
	if balanceAfterDebitDelete.Balance != 0 {
		t.Errorf("rep balance after delete = %d, want 0", balanceAfterDebitDelete.Balance)
	}

	var appErr *apperr.AppError
	_, err = env.Services.Representatives.RecordTransaction(env.Ctx, 99999, services.RecordRepTransactionInput{
		Direction: enums.RepDebit, Amount: 10, OccurredAt: baseTime,
	})
	if !asAppError(err, &appErr) || appErr.Code != apperr.CodeValidation {
		t.Errorf("unknown representative must be validation error, got %v", err)
	}
}

func TestNationalCodeUniqueAmongActives(t *testing.T) {
	env := openTestEnv(t)
	code := "0012345678"

	if _, err := env.Services.Representatives.Create(env.Ctx, services.CreateRepresentativeInput{
		FullName: "اولی", Phone: "09121", NationalCode: strPtr(code),
		Currency: enums.CurrencyRIAL, StartDate: baseTime,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := env.Services.Representatives.Create(env.Ctx, services.CreateRepresentativeInput{
		FullName: "دومی", Phone: "09122", NationalCode: strPtr(code),
		Currency: enums.CurrencyRIAL, StartDate: baseTime,
	}); err == nil {
		t.Error("duplicate national code must be rejected")
	}
}
