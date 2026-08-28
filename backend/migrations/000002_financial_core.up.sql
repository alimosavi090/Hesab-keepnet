CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL,
    password_hash TEXT    NOT NULL,
    display_name  TEXT    NOT NULL DEFAULT '',
    role          TEXT    NOT NULL DEFAULT 'ADMIN' CHECK (role IN ('ADMIN')),
    is_active     INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    deleted_at    DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_active_username ON users(username) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS sessions (
    token_hash TEXT    PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE RESTRICT,
    expires_at DATETIME NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS bank_accounts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL,
    bank_name       TEXT    NOT NULL,
    card_number     TEXT,
    currency        TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    initial_balance INTEGER NOT NULL DEFAULT 0,
    description     TEXT,
    is_active       INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL,
    deleted_at      DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_bank_accounts_active_name ON bank_accounts(name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS categories (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    type       TEXT    NOT NULL CHECK (type IN ('BUSINESS', 'PERSONAL')),
    parent_id  INTEGER REFERENCES categories(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    is_active  INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_tree
    ON categories(IFNULL(parent_id, 0), name, type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_categories_type ON categories(type);

CREATE TABLE IF NOT EXISTS sales (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    total_amount  INTEGER NOT NULL CHECK (total_amount > 0),
    currency      TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    sold_at       DATETIME NOT NULL,
    customer_name TEXT,
    description   TEXT,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    deleted_at    DATETIME
);

CREATE INDEX IF NOT EXISTS idx_sales_sold_at ON sales(sold_at);
CREATE INDEX IF NOT EXISTS idx_sales_currency ON sales(currency);

CREATE TABLE IF NOT EXISTS sale_payments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    sale_id         INTEGER NOT NULL REFERENCES sales(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    bank_account_id INTEGER NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    gateway         TEXT    NOT NULL CHECK (gateway IN ('ZARINPAL', 'CARD_TO_CARD', 'SUPPORT')),
    amount          INTEGER NOT NULL CHECK (amount > 0),
    currency        TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    paid_at         DATETIME NOT NULL,
    gateway_ref     TEXT,
    description     TEXT,
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL,
    deleted_at      DATETIME
);

CREATE INDEX IF NOT EXISTS idx_sale_payments_sale_id        ON sale_payments(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_payments_bank_account   ON sale_payments(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_sale_payments_paid_at        ON sale_payments(paid_at);

CREATE TABLE IF NOT EXISTS expenses (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id     INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    bank_account_id INTEGER REFERENCES bank_accounts(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    amount          INTEGER NOT NULL CHECK (amount > 0),
    currency        TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    occurred_at     DATETIME NOT NULL,
    description     TEXT,
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL,
    deleted_at      DATETIME
);

CREATE INDEX IF NOT EXISTS idx_expenses_category_id      ON expenses(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_bank_account     ON expenses(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_expenses_occurred_at      ON expenses(occurred_at);
CREATE INDEX IF NOT EXISTS idx_expenses_category_date    ON expenses(category_id, occurred_at);

CREATE TABLE IF NOT EXISTS transfers (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    from_account_id  INTEGER NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    to_account_id    INTEGER NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    amount           INTEGER NOT NULL CHECK (amount > 0),
    currency         TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    transferred_at   DATETIME NOT NULL,
    description      TEXT,
    created_at       DATETIME NOT NULL,
    updated_at       DATETIME NOT NULL,
    deleted_at       DATETIME,
    CHECK (from_account_id <> to_account_id)
);

CREATE INDEX IF NOT EXISTS idx_transfers_transferred_at ON transfers(transferred_at);
CREATE INDEX IF NOT EXISTS idx_transfers_from           ON transfers(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transfers_to             ON transfers(to_account_id);

CREATE TABLE IF NOT EXISTS transactions (
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

CREATE TABLE IF NOT EXISTS representatives (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name              TEXT    NOT NULL,
    phone                  TEXT    NOT NULL,
    email                  TEXT,
    address                TEXT,
    national_code          TEXT,
    currency               TEXT    NOT NULL DEFAULT 'RIAL' CHECK (currency IN ('RIAL', 'USD')),
    notes                  TEXT,
    start_date             DATETIME NOT NULL,
    is_active              INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at             DATETIME NOT NULL,
    updated_at             DATETIME NOT NULL,
    deleted_at             DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_representatives_national_code
    ON representatives(national_code)
    WHERE deleted_at IS NULL AND national_code IS NOT NULL;

CREATE TABLE IF NOT EXISTS representative_transactions (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    representative_id INTEGER NOT NULL REFERENCES representatives(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    direction         TEXT    NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount            INTEGER NOT NULL CHECK (amount > 0),
    currency          TEXT    NOT NULL CHECK (currency IN ('RIAL', 'USD')),
    occurred_at       DATETIME NOT NULL,
    sale_id           INTEGER REFERENCES sales(id) ON DELETE SET NULL ON UPDATE RESTRICT,
    description       TEXT,
    created_at        DATETIME NOT NULL,
    updated_at        DATETIME NOT NULL,
    deleted_at        DATETIME
);

CREATE INDEX IF NOT EXISTS idx_rep_tx_representative_date ON representative_transactions(representative_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_rep_tx_occurred_at         ON representative_transactions(occurred_at);

CREATE TABLE IF NOT EXISTS reminders (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT    NOT NULL,
    description     TEXT,
    due_date        DATETIME NOT NULL,
    repeat_interval TEXT    NOT NULL DEFAULT 'NONE' CHECK (repeat_interval IN ('NONE', 'MONTHLY', 'YEARLY')),
    is_done         INTEGER NOT NULL DEFAULT 0 CHECK (is_done IN (0, 1)),
    completed_at    DATETIME,
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL,
    deleted_at      DATETIME
);

CREATE INDEX IF NOT EXISTS idx_reminders_due_date ON reminders(due_date);

CREATE TABLE IF NOT EXISTS audit_logs (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_user_id  INTEGER REFERENCES users(id) ON DELETE SET NULL ON UPDATE RESTRICT,
    action         TEXT    NOT NULL,
    entity_type    TEXT    NOT NULL,
    entity_id      INTEGER NOT NULL,
    metadata       TEXT,
    ip_address     TEXT,
    user_agent     TEXT,
    created_at     DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_entity     ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_actor      ON audit_logs(actor_user_id);
