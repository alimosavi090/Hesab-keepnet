package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strconv"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version    INTEGER PRIMARY KEY,
	name       TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT (datetime('now'))
)`

var migrationFilePattern = regexp.MustCompile(`^(\d{6})_([a-z0-9_]+)\.(up|down)\.sql$`)

type MigrationFile struct {
	Version int
	Name    string
	Up      string
	Down    string
}

type MigrationRecord struct {
	Version int
	Name    string
}

func Migrate(root fs.FS, db *DB) ([]MigrationRecord, error) {
	files, err := LoadMigrationFiles(root)
	if err != nil {
		return nil, err
	}
	if err := db.DB.Exec(createMigrationsTable).Error; err != nil {
		return nil, fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	applied, err := AppliedVersions(db)
	if err != nil {
		return nil, err
	}
	done := make(map[int]bool, len(applied))
	for _, r := range applied {
		done[r.Version] = true
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}

	for _, m := range files {
		if done[m.Version] {
			continue
		}
		if err := applyMigration(context.Background(), sqlDB, m); err != nil {
			return applied, fmt.Errorf("migration %06d_%s failed: %w", m.Version, m.Name, err)
		}
		applied = append(applied, MigrationRecord{Version: m.Version, Name: m.Name})
	}
	return applied, nil
}

func RollbackAll(root fs.FS, db *DB) ([]MigrationRecord, error) {
	files, err := LoadMigrationFiles(root)
	if err != nil {
		return nil, err
	}
	applied, err := AppliedVersions(db)
	if err != nil {
		return nil, err
	}
	byVersion := make(map[int]MigrationFile, len(files))
	for _, m := range files {
		byVersion[m.Version] = m
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}

	var rolledBack []MigrationRecord
	for i := len(applied) - 1; i >= 0; i-- {
		record := applied[i]
		m, ok := byVersion[record.Version]
		if !ok {
			return rolledBack, fmt.Errorf("no local migration file for applied version %06d", record.Version)
		}
		if err := revertMigration(context.Background(), sqlDB, m); err != nil {
			return rolledBack, fmt.Errorf("rollback %06d_%s failed: %w", m.Version, m.Name, err)
		}
		rolledBack = append(rolledBack, record)
	}
	return rolledBack, nil
}

func LoadMigrationFiles(root fs.FS) ([]MigrationFile, error) {
	entries, err := fs.Glob(root, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("scan migrations: %w", err)
	}

	collected := make(map[int]*MigrationFile)
	for _, entry := range entries {
		match := migrationFilePattern.FindStringSubmatch(entry)
		if match == nil {
			return nil, fmt.Errorf("invalid migration file name %q: expected NNNNNN_name.up.sql / .down.sql", entry)
		}
		version, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", entry, err)
		}
		content, err := fs.ReadFile(root, entry)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry, err)
		}
		m := collected[version]
		if m == nil {
			m = &MigrationFile{Version: version, Name: match[2]}
			collected[version] = m
		}
		if match[3] == "up" {
			if m.Up != "" {
				return nil, fmt.Errorf("duplicate up migration for version %06d", version)
			}
			m.Up = string(content)
		} else {
			if m.Down != "" {
				return nil, fmt.Errorf("duplicate down migration for version %06d", version)
			}
			m.Down = string(content)
		}
	}

	files := make([]MigrationFile, 0, len(collected))
	for _, m := range collected {
		if m.Up == "" || m.Down == "" {
			missingSide := "up"
			if m.Up != "" {
				missingSide = "down"
			}
			return nil, fmt.Errorf("migration %06d_%s is missing its %s side", m.Version, m.Name, missingSide)
		}
		files = append(files, *m)
	}
	slices.SortFunc(files, func(a, b MigrationFile) int { return a.Version - b.Version })
	return files, nil
}

func AppliedVersions(db *DB) ([]MigrationRecord, error) {
	var records []MigrationRecord
	err := db.DB.Raw("SELECT version, name FROM schema_migrations ORDER BY version").Scan(&records).Error
	return records, err
}

func applyMigration(ctx context.Context, sq *sql.DB, m MigrationFile) error {
	return runScript(ctx, sq, m.Up, func(tx *sql.Conn) error {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.Version, m.Name)
		return err
	})
}

func revertMigration(ctx context.Context, sq *sql.DB, m MigrationFile) error {
	return runScript(ctx, sq, m.Down, func(tx *sql.Conn) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = ?", m.Version)
		return err
	})
}

func runScript(ctx context.Context, sq *sql.DB, script string, extra func(*sql.Conn) error) error {
	conn, err := sq.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
		}
	}()

	if _, err := conn.ExecContext(ctx, script); err != nil {
		return err
	}
	if extra != nil {
		if err := extra(conn); err != nil {
			return err
		}
	}
	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}
