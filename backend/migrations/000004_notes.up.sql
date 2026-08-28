-- Data-attached notes (representatives, sales, bank accounts) + free daily journal
CREATE TABLE IF NOT EXISTS notes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('REPRESENTATIVE', 'SALE', 'BANK_ACCOUNT', 'JOURNAL')),
    entity_id   INTEGER, -- NULL for JOURNAL entries
    body        TEXT NOT NULL CHECK (length(trim(body)) > 0),
    tags        TEXT NOT NULL DEFAULT '', -- comma-separated lowercase tags
    pinned      INTEGER NOT NULL DEFAULT 0 CHECK (pinned IN (0, 1)),
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    deleted_at  DATETIME
);

CREATE INDEX IF NOT EXISTS idx_notes_entity      ON notes(entity_type, entity_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notes_pinned      ON notes(pinned) WHERE deleted_at IS NULL AND entity_type = 'JOURNAL';
CREATE INDEX IF NOT EXISTS idx_notes_journal     ON notes(created_at DESC) WHERE deleted_at IS NULL AND entity_type = 'JOURNAL';
