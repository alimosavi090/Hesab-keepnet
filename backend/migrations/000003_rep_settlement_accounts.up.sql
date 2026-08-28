-- Settlements of representative debts land on real bank accounts:
-- 1) rebuild `transactions` so an INCOME row may point to a representative settlement,
-- 2) link representative_transactions to the destination account + ledger row.
--
-- SQLite cannot alter CHECK constraints, so the ledger table is rebuilt here.

CREATE TABLE IF NOT EXISTS transactions_new (
    id                           INTEGER PRIMARY KEY AUTOINCREMENT,
    bank_account_id              INTEGER NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    type                         TEXT    NOT NULL CHECK (type IN ('INCOME', 'EXPENSE', 'TRANSFER_IN', 'TRANSFER_OUT')),
    amount                       INTEGER NOT NULL CHECK (amount > 0),
    currency                     TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    occurred_at                  DATETIME NOT NULL,
    description                  TEXT,
    sale_payment_id              INTEGER REFERENCES sale_payments(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    expense_id                   INTEGER REFERENCES expenses(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    transfer_id                  INTEGER REFERENCES transfers(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    representative_transaction_id INTEGER REFERENCES representative_transactions(id) ON DELETE SET NULL ON UPDATE RESTRICT,
    created_at                   DATETIME NOT NULL,
    updated_at                   DATETIME NOT NULL,
    deleted_at                   DATETIME,
    CHECK (
        (type = 'INCOME'      AND ((sale_payment_id IS NOT NULL AND expense_id IS NULL AND transfer_id IS NULL AND representative_transaction_id IS NULL)
                                   OR (representative_transaction_id IS NOT NULL AND sale_payment_id IS NULL AND expense_id IS NULL AND transfer_id IS NULL)))
     OR (type = 'EXPENSE'      AND sale_payment_id IS NULL AND expense_id IS NOT NULL AND transfer_id IS NULL AND representative_transaction_id IS NULL)
     OR (type = 'TRANSFER_IN'  AND sale_payment_id IS NULL AND expense_id IS NULL AND transfer_id IS NOT NULL AND representative_transaction_id IS NULL)
     OR (type = 'TRANSFER_OUT' AND sale_payment_id IS NULL AND expense_id IS NULL AND transfer_id IS NOT NULL AND representative_transaction_id IS NULL)
    )
);

INSERT INTO transactions_new (
    id, bank_account_id, type, amount, currency, occurred_at, description,
    sale_payment_id, expense_id, transfer_id, representative_transaction_id,
    created_at, updated_at, deleted_at
)
SELECT
    id, bank_account_id, type, amount, currency, occurred_at, description,
    sale_payment_id, expense_id, transfer_id, NULL,
    created_at, updated_at, deleted_at
FROM transactions;

DROP TABLE transactions;
ALTER TABLE transactions_new RENAME TO transactions;

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

ALTER TABLE representative_transactions ADD COLUMN bank_account_id       INTEGER REFERENCES bank_accounts(id) ON DELETE SET NULL ON UPDATE RESTRICT;
ALTER TABLE representative_transactions ADD COLUMN ledger_transaction_id INTEGER REFERENCES transactions(id) ON DELETE SET NULL ON UPDATE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_rep_tx_bank_account        ON representative_transactions(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_rep_tx_ledger_transaction  ON representative_transactions(ledger_transaction_id);
