package tests

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/migrations"
)

func openCoreDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "constraints.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := database.Migrate(migrations.FS, db); err != nil {
		t.Fatal(err)
	}
	return db
}

func mustFail(t *testing.T, db *database.DB, reason string, query string, args ...any) {
	t.Helper()
	if err := db.DB.Exec(query, args...).Error; err == nil {
		t.Errorf("%s: query must fail but succeeded", reason)
	}
}

func insertAccount(t *testing.T, db *database.DB, name, currency string) int64 {
	t.Helper()
	res := db.DB.Exec(
		"INSERT INTO bank_accounts (name, bank_name, currency, initial_balance, is_active, created_at, updated_at) VALUES (?, 'b', ?, 0, 1, datetime('now'), datetime('now'))",
		name, currency,
	)
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	var id int64
	db.DB.Raw("SELECT id FROM bank_accounts WHERE name = ?", name).Scan(&id)
	return id
}

func TestForeignKeysEnforced(t *testing.T) {
	db := openCoreDB(t)

	mustFail(t, db, "expense with unknown category",
		"INSERT INTO expenses (category_id, amount, currency, occurred_at, created_at, updated_at) VALUES (99999, 100, 'RIAL', datetime('now'), datetime('now'), datetime('now'))")

	mustFail(t, db, "payment with unknown sale/account",
		"INSERT INTO sale_payments (sale_id, bank_account_id, gateway, amount, currency, paid_at, created_at, updated_at) VALUES (99999, 99999, 'ZARINPAL', 100, 'RIAL', datetime('now'), datetime('now'), datetime('now'))")

	mustFail(t, db, "rep transaction with unknown representative",
		"INSERT INTO representative_transactions (representative_id, direction, amount, currency, occurred_at, created_at, updated_at) VALUES (99999, 'DEBIT', 100, 'RIAL', datetime('now'), datetime('now'), datetime('now'))")
}

func TestCheckConstraints(t *testing.T) {
	db := openCoreDB(t)

	db.DB.Exec("INSERT INTO categories (name, type, parent_id, is_active, created_at, updated_at) VALUES ('c', 'BUSINESS', NULL, 1, datetime('now'), datetime('now'))")
	var categoryID int64
	db.DB.Raw("SELECT id FROM categories LIMIT 1").Scan(&categoryID)

	mustFail(t, db, "negative expense amount",
		"INSERT INTO expenses (category_id, amount, currency, occurred_at, created_at, updated_at) VALUES (?, -5, 'RIAL', datetime('now'), datetime('now'), datetime('now'))", categoryID)

	mustFail(t, db, "null expense category",
		"INSERT INTO expenses (category_id, amount, currency, occurred_at, created_at, updated_at) VALUES (NULL, 100, 'RIAL', datetime('now'), datetime('now'), datetime('now'))")

	accountID := insertAccount(t, db, "acc", "RIAL")

	mustFail(t, db, "transfer from == to",
		"INSERT INTO transfers (from_account_id, to_account_id, amount, currency, transferred_at, created_at, updated_at) VALUES (?, ?, 100, 'RIAL', datetime('now'), datetime('now'), datetime('now'))",
		accountID, accountID)

	mustFail(t, db, "invalid repeat interval",
		"INSERT INTO reminders (title, due_date, repeat_interval, is_done, created_at, updated_at) VALUES ('r', datetime('now'), 'WEEKLY', 0, datetime('now'), datetime('now'))")

	mustFail(t, db, "invalid account currency",
		"INSERT INTO bank_accounts (name, bank_name, currency, initial_balance, is_active, created_at, updated_at) VALUES ('bad', 'b', 'EUR', 0, 1, datetime('now'), datetime('now'))")

	mustFail(t, db, "invalid user role",
		"INSERT INTO users (username, password_hash, display_name, role, is_active, created_at, updated_at) VALUES ('u', 'h', '', 'SUPERUSER', 1, datetime('now'), datetime('now'))")
}

func TestLedgerSourceConstraintAndCurrencyTrigger(t *testing.T) {
	db := openCoreDB(t)
	usdID := insertAccount(t, db, "usd", "USD")
	tomanID := insertAccount(t, db, "toman", "RIAL")

	mustFail(t, db, "INCOME row without source document",
		"INSERT INTO transactions (bank_account_id, type, amount, currency, occurred_at, description, created_at, updated_at) VALUES (?, 'INCOME', 100, 'USD', datetime('now'), NULL, datetime('now'), datetime('now'))", usdID)

	mustFail(t, db, "EXPENSE row without source document",
		"INSERT INTO transactions (bank_account_id, type, amount, currency, occurred_at, description, created_at, updated_at) VALUES (?, 'EXPENSE', 100, 'USD', datetime('now'), NULL, datetime('now'), datetime('now'))", usdID)

	mustFail(t, db, "TRANSFER_IN row without transfer document",
		"INSERT INTO transactions (bank_account_id, type, amount, currency, occurred_at, description, created_at, updated_at) VALUES (?, 'TRANSFER_IN', 100, 'USD', datetime('now'), NULL, datetime('now'), datetime('now'))", usdID)

	err := db.DB.Exec(
		"INSERT INTO transactions (bank_account_id, type, amount, currency, occurred_at, description, created_at, updated_at) VALUES (?, 'INCOME', 100, 'USD', datetime('now'), NULL, datetime('now'), datetime('now'))",
		tomanID,
	).Error
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "currency") {
		t.Errorf("currency-mismatch trigger must abort insert, got %v", err)
	}
}

func TestPartialUniqueIndexes(t *testing.T) {
	db := openCoreDB(t)
	insertAccount(t, db, "melli", "RIAL")

	mustFail(t, db, "duplicate active account name",
		"INSERT INTO bank_accounts (name, bank_name, currency, initial_balance, is_active, created_at, updated_at) VALUES ('melli', 'x', 'RIAL', 0, 1, datetime('now'), datetime('now'))")

	insertCategory := func() error {
		return db.DB.Exec("INSERT INTO categories (name, type, parent_id, is_active, created_at, updated_at) VALUES ('کافه', 'PERSONAL', NULL, 1, datetime('now'), datetime('now'))").Error
	}
	if err := insertCategory(); err != nil {
		t.Fatal(err)
	}
	if err := insertCategory(); err == nil {
		t.Error("duplicate root category (name,type) must fail")
	}

	if err := db.DB.Exec("UPDATE categories SET deleted_at = datetime('now') WHERE name = 'کافه'").Error; err != nil {
		t.Fatal(err)
	}
	if err := insertCategory(); err != nil {
		t.Errorf("recreating after soft delete must be allowed: %v", err)
	}
}

func TestNoStoredBalanceColumn(t *testing.T) {
	db := openCoreDB(t)

	var count int64
	db.DB.Raw("SELECT COUNT(*) FROM pragma_table_info('bank_accounts') WHERE name LIKE '%balance%' AND name <> 'initial_balance'").Scan(&count)
	if count != 0 {
		t.Errorf("bank_accounts must not store derived balance columns, found %d", count)
	}

	var repCount int64
	db.DB.Raw("SELECT COUNT(*) FROM pragma_table_info('representatives') WHERE name LIKE '%balance%'").Scan(&repCount)
	if repCount != 0 {
		t.Errorf("representatives must not store wallet balance columns, found %d", repCount)
	}
}

func TestRequiredIndexesExist(t *testing.T) {
	db := openCoreDB(t)

	required := []string{
		"idx_transactions_account_date",
		"idx_transactions_type",
		"idx_expenses_category_id",
		"idx_expenses_bank_account",
		"idx_expenses_occurred_at",
		"idx_sales_sold_at",
		"idx_sale_payments_sale_id",
		"idx_sale_payments_bank_account",
		"idx_rep_tx_representative_date",
		"idx_transfers_transferred_at",
		"idx_reminders_due_date",
		"uq_users_active_username",
		"uq_bank_accounts_active_name",
		"uq_categories_tree",
		"uq_transactions_sale_payment",
		"uq_transactions_expense",
		"uq_representatives_national_code",
		"trg_transactions_currency_match",
	}

	for _, name := range required {
		var count int64
		db.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE name = ?", name).Scan(&count)
		if count != 1 {
			t.Errorf("required index/trigger %q missing", name)
		}
	}
}

func TestMigrationDownRemovesFinancialSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cycle.db")

	db, err := database.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if _, err := database.Migrate(migrations.FS, db); err != nil {
		t.Fatal(err)
	}
	if err := database.Seed(context.Background(), db.DB, "", ""); err != nil {
		t.Fatal(err)
	}

	coreTables := []string{"users", "bank_accounts", "categories", "sales", "sale_payments", "expenses", "transfers", "transactions", "representatives", "representative_transactions", "reminders", "audit_logs"}

	for _, table := range coreTables {
		var count int64
		db.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if count != 1 {
			t.Errorf("table %q missing after up migration", table)
		}
	}

	if _, err := database.RollbackAll(migrations.FS, db); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	for _, table := range coreTables {
		var count int64
		db.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if count != 0 {
			t.Errorf("table %q still exists after down migration", table)
		}
	}

	if _, err := database.Migrate(migrations.FS, db); err != nil {
		t.Fatalf("second migrate cycle after rollback failed: %v", err)
	}
}
