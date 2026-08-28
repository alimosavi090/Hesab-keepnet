package database

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func mapFS(files map[string]string) fs.FS {
	sys := fstest.MapFS{}
	for name, content := range files {
		sys[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return sys
}

func TestMigrateAppliesInOrder(t *testing.T) {
	db := openTestDB(t)
	files := mapFS(map[string]string{
		"000001_create_t.up.sql":   "CREATE TABLE t (id INTEGER PRIMARY KEY);",
		"000001_create_t.down.sql": "DROP TABLE IF EXISTS t;",
		"000002_create_u.up.sql":   "CREATE TABLE u (id INTEGER PRIMARY KEY);",
		"000002_create_u.down.sql": "DROP TABLE IF EXISTS u;",
	})

	applied, err := Migrate(files, db)
	if err != nil {
		t.Fatalf("Migrate() error: %v", err)
	}
	if len(applied) != 2 || applied[0].Version != 1 || applied[1].Version != 2 {
		t.Fatalf("applied = %+v", applied)
	}
	for _, table := range []string{"t", "u", "schema_migrations"} {
		var count int64
		if err := db.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", table).Scan(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("table %q missing", table)
		}
	}

	reApplied, err := Migrate(files, db)
	if err != nil {
		t.Fatalf("second Migrate() error: %v", err)
	}
	if len(reApplied) != 2 {
		t.Errorf("re-run must be idempotent, got %d records", len(reApplied))
	}
}

func TestRollbackAllReversesEverything(t *testing.T) {
	db := openTestDB(t)
	files := mapFS(map[string]string{
		"000001_create_t.up.sql":   "CREATE TABLE t (id INTEGER PRIMARY KEY);",
		"000001_create_t.down.sql": "DROP TABLE IF EXISTS t;",
	})

	if _, err := Migrate(files, db); err != nil {
		t.Fatal(err)
	}

	rolledBack, err := RollbackAll(files, db)
	if err != nil {
		t.Fatalf("RollbackAll() error: %v", err)
	}
	if len(rolledBack) != 1 || rolledBack[0].Version != 1 {
		t.Fatalf("rolledBack = %+v", rolledBack)
	}

	var tCount, migrationsCount int64
	if err := db.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='t'").Scan(&tCount).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.DB.Raw("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationsCount).Error; err != nil {
		t.Fatal(err)
	}
	if tCount != 0 || migrationsCount != 0 {
		t.Errorf("expected migrated table and records gone, got tables=%d records=%d", tCount, migrationsCount)
	}
}

func TestMigrateFailedMigrationLeavesCleanState(t *testing.T) {
	db := openTestDB(t)

	broken := mapFS(map[string]string{
		"000001_bad.up.sql":   "CREATE TABLE broken (id INTEGER PRIMARY KEY; SELECT 1;",
		"000001_bad.down.sql": "SELECT 1;",
	})
	if _, err := Migrate(broken, db); err == nil {
		t.Fatal("expected migration failure")
	}

	good := mapFS(map[string]string{
		"000001_good.up.sql":   "CREATE TABLE good (id INTEGER PRIMARY KEY);",
		"000001_good.down.sql": "DROP TABLE good;",
	})
	applied, err := Migrate(good, db)
	if err != nil {
		t.Fatalf("retry after failure failed: %v", err)
	}
	if len(applied) != 1 || applied[0].Name != "good" {
		t.Fatalf("applied = %+v", applied)
	}
}

func TestLoadMigrationFilesValidation(t *testing.T) {
	cases := []struct {
		name    string
		files   map[string]string
		wantErr string
	}{
		{"missing down", map[string]string{"000001_x.up.sql": "SELECT 1;"}, "missing its down"},
		{"missing up", map[string]string{"000001_x.down.sql": "SELECT 1;"}, "missing its up"},
		{"bad filename", map[string]string{"foo.sql": "SELECT 1;", "000001_x.up.sql": "SELECT 1;", "000001_x.down.sql": "SELECT 1;"}, "invalid migration file name"},
		{"duplicate version", map[string]string{"000001_a.up.sql": "SELECT 1;", "000001_a.down.sql": "SELECT 1;", "000001_b.up.sql": "SELECT 1;", "000001_b.down.sql": "SELECT 1;"}, "duplicate"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadMigrationFiles(mapFS(tc.files))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestAppliedVersionsEmptyOnFreshDB(t *testing.T) {
	db := openTestDB(t)
	if err := db.DB.Exec(createMigrationsTable).Error; err != nil {
		t.Fatal(err)
	}
	records, err := AppliedVersions(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("records = %+v, want empty", records)
	}
}
