package tests

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/passwordhash"
	"github.com/ali/hesab-keepnet/backend/migrations"
)

func openSeededDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "seed.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := database.Migrate(migrations.FS, db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSeedIdempotency(t *testing.T) {
	db := openSeededDB(t)
	ctx := context.Background()

	if err := database.Seed(ctx, db.DB, "", ""); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	var categoryCount int64
	db.DB.Raw("SELECT COUNT(*) FROM categories").Scan(&categoryCount)

	if err := database.Seed(ctx, db.DB, "", ""); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	if err := database.Seed(ctx, db.DB, "", ""); err != nil {
		t.Fatalf("third seed: %v", err)
	}

	var categoryCountAfter int64
	db.DB.Raw("SELECT COUNT(*) FROM categories").Scan(&categoryCountAfter)

	if categoryCount != 12 || categoryCountAfter != 12 {
		t.Errorf("categories first=%d after-reruns=%d, want 12/12", categoryCount, categoryCountAfter)
	}
}

func TestAdminUserSeed(t *testing.T) {
	db := openSeededDB(t)
	ctx := context.Background()

	if err := database.Seed(ctx, db.DB, "admin", "s3cure-passphrase"); err != nil {
		t.Fatal(err)
	}
	if err := database.Seed(ctx, db.DB, "admin", "other-password"); err != nil {
		t.Fatal(err)
	}

	var count int64
	db.DB.Raw("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if count != 1 {
		t.Errorf("admin rows after double seed = %d, want 1 (idempotent, password untouched)", count)
	}

	var hash string
	db.DB.Raw("SELECT password_hash FROM users WHERE username = 'admin'").Scan(&hash)

	ok, err := passwordhash.Verify("s3cure-passphrase", hash)
	if err != nil || !ok {
		t.Errorf("stored hash must verify against originally seeded password: ok=%v err=%v", ok, err)
	}
	if ok2, _ := passwordhash.Verify("other-password", hash); ok2 {
		t.Error("re-seed must not overwrite existing password")
	}
}

func TestPasswordHashFormat(t *testing.T) {
	hash, err := passwordhash.Hash("top-secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(hash, "$argon2id$") {
		t.Errorf("unexpected hash format: %.30s", hash)
	}

	hash2, _ := passwordhash.Hash("top-secret")
	if hash == hash2 {
		t.Error("two hashes of the same password must differ (random salt)")
	}

	if ok, _ := passwordhash.Verify("wrong-password", hash); ok {
		t.Error("wrong password must not verify")
	}
	if ok, err := passwordhash.Verify("top-secret", hash); err != nil || !ok {
		t.Errorf("correct password must verify: ok=%v err=%v", ok, err)
	}
}
