DROP TRIGGER IF EXISTS trg_transactions_currency_match;
DROP INDEX IF EXISTS idx_audit_actor;
DROP INDEX IF EXISTS idx_audit_created_at;
DROP INDEX IF EXISTS idx_audit_entity;
DROP TABLE IF EXISTS audit_logs;

DROP INDEX IF EXISTS idx_reminders_due_date;
DROP TABLE IF EXISTS reminders;

DROP INDEX IF EXISTS idx_rep_tx_occurred_at;
DROP INDEX IF EXISTS idx_rep_tx_representative_date;
DROP TABLE IF EXISTS representative_transactions;

DROP INDEX IF EXISTS uq_representatives_national_code;
DROP TABLE IF EXISTS representatives;

DROP INDEX IF EXISTS idx_transactions_transfer;
DROP INDEX IF EXISTS idx_transactions_type;
DROP INDEX IF EXISTS idx_transactions_account_date;
DROP INDEX IF EXISTS uq_transactions_expense;
DROP INDEX IF EXISTS uq_transactions_sale_payment;
DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS idx_transfers_from;
DROP INDEX IF EXISTS idx_transfers_to;
DROP INDEX IF EXISTS idx_transfers_transferred_at;
DROP TABLE IF EXISTS transfers;

DROP INDEX IF EXISTS idx_expenses_category_date;
DROP INDEX IF EXISTS idx_expenses_occurred_at;
DROP INDEX IF EXISTS idx_expenses_bank_account;
DROP INDEX IF EXISTS idx_expenses_category_id;
DROP TABLE IF EXISTS expenses;

DROP INDEX IF EXISTS idx_sale_payments_paid_at;
DROP INDEX IF EXISTS idx_sale_payments_bank_account;
DROP INDEX IF EXISTS idx_sale_payments_sale_id;
DROP TABLE IF EXISTS sale_payments;

DROP INDEX IF EXISTS idx_sales_currency;
DROP INDEX IF EXISTS idx_sales_sold_at;
DROP TABLE IF EXISTS sales;

DROP INDEX IF EXISTS idx_categories_type;
DROP INDEX IF EXISTS uq_categories_tree;
DROP TABLE IF EXISTS categories;

DROP TABLE IF EXISTS bank_accounts;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS sessions;

DROP INDEX IF EXISTS uq_users_active_username;
DROP TABLE IF EXISTS users;
