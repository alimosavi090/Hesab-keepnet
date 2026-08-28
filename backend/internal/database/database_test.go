package database

import (
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestOpenCreatesDirectoryAndConnects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
}

func TestSQLitePragmas(t *testing.T) {
	db := openTestDB(t)

	fk, err := db.PragmaInt("foreign_keys")
	if err != nil {
		t.Fatalf("pragma foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1 (ON)", fk)
	}

	mode, err := db.PragmaText("journal_mode")
	if err != nil {
		t.Fatalf("pragma journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want wal", mode)
	}

	sync, err := db.PragmaInt("synchronous")
	if err != nil {
		t.Fatalf("pragma synchronous: %v", err)
	}
	if sync != 1 {
		t.Errorf("synchronous = %d, want 1 (NORMAL)", sync)
	}

	busy, err := db.PragmaInt("busy_timeout")
	if err != nil {
		t.Fatalf("pragma busy_timeout: %v", err)
	}
	if busy != busyTimeoutMs {
		t.Errorf("busy_timeout = %d, want %d", busy, busyTimeoutMs)
	}
}

func TestWALSidecarFilesCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.DB.Exec("CREATE TABLE probe (id INTEGER PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	mode, err := reopened.PragmaText("journal_mode")
	if err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Errorf("reopened db journal_mode = %q, want wal (persistent)", mode)
	}
}

func TestCloseThenPingFails(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	if err := db.Close(); err != nil && err.Error() == "sql: database is closed" {
		t.Logf("second Close: %v", err)
	}
	if err := db.Ping(); err == nil {
		t.Error("Ping on closed DB must fail")
	}
}
