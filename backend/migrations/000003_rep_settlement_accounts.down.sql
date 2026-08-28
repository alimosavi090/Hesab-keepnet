-- Reverse of 000003: unlink representative settlements from the ledger.

DROP INDEX IF EXISTS idx_rep_tx_ledger_transaction;
DROP INDEX IF EXISTS idx_rep_tx_bank_account;

ALTER TABLE representative_transactions DROP COLUMN ledger_transaction_id;
ALTER TABLE representative_transactions DROP COLUMN bank_account_id;

CREATE TABLE IF NOT EXISTS transactions_orig (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    bank_account_id  INTEGER NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    type             TEXT    NOT NULL CHECK (type IN ('INCOME', 'EXPENSE', 'TRANSFER_IN', 'TRANSFER_OUT')),
    amount           INTEGER NOT NULL CHECK (amount > 0),
    currency         TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    occurred_at      DATETIME NOT NULL,
    description      TEXT,
    sale_payment_id  INTEGER REFERENCES sale_payments(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    expense_id       INTEGER REFERENCES expenses(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    transfer_id      INTEGER REFERENCES transfers(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    created_at       DATETIME NOT NULL,
    updated_at       DATETIME NOT NULL,
    deleted_at       DATETIME,
    CHECK (
        (type = 'INCOME'       AND sale_payment_id IS NOT NULL AND expense_id IS NULL AND transfer_id IS NULL)
     OR (type = 'EXPENSE'      AND sale_payment_id IS NULL AND expense_id IS NOT NULL AND transfer_id IS NULL)
     OR (type = 'TRANSFER_IN'  AND sale_payment_id IS NULL AND expense_id IS NULL AND transfer_id IS NOT NULL)
     OR (type = 'TRANSFER_OUT' AND sale_payment_id IS NULL AND expense_id IS NULL AND transfer_id IS NOT NULL)
    )
);

INSERT INTO transactions_orig (
    id, bank_account_id, type, amount, currency, occurred_at, description,
    sale_payment_id, expense_id, transfer_id, created_at, updated_at, deleted_at
)
SELECT
    id, bank_account_id, type, amount, currency, occurred_at, description,
    sale_payment_id, expense_id, transfer_id, created_at, updated_at, deleted_at
FROM transactions;

DROP TABLE transactions;
ALTER TABLE transactions_orig RENAME TO transactions;

CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_sale_payment
    ON transactions(sale_payment_id) WHERE sale_payment_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_expense
    ON transactions(expense_id) WHERE expense_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(bank_account_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_transactions_type         ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_transfer     ON transactions(transfer_id);

CREATE TRIGGER IF NOT EXISTS trg_transactions_currency_match
BEFORE INSERT ON transactions
WHEN NEW.currency <> (SELECT currency FROM bank_accounts WHERE id = NEW.bank_account_id)
BEGIN
    SELECT RAISE(ABORT, 'transaction currency must match bank account currency');
END;
