# Migrations

Convention: `NNNNNN_name.up.sql` + `NNNNNN_name.down.sql` (golang-migrate-compatible naming).

Rules:

- Versions are 6-digit, strictly increasing, unique.
- Every migration runs inside a single `BEGIN IMMEDIATE` transaction; on failure it rolls back completely.
- Applied versions are tracked in the `schema_migrations` table.
- Never edit an already-applied migration. Add a new one instead.
- Financial domain tables (users, bank_accounts, transactions, sales, ...) will be introduced here in later phases.

Phase 1 ships only the runner infrastructure with a no-op baseline migration (`000001_init`).
