package tests

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/config"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/routes"
	"github.com/ali/hesab-keepnet/backend/migrations"
)

func TestFullStackHealthFlow(t *testing.T) {
	t.Setenv("APP_ENV", config.EnvTest)
	t.Setenv("APP_PORT", "8081")
	t.Setenv("DATABASE_PATH", filepath.Join(t.TempDir(), "smoke", "accounting.db"))
	t.Setenv("CORS_ORIGIN", "http://localhost:3000")
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load(): %v", err)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		t.Fatalf("database.Open(): %v", err)
	}
	defer func() { _ = db.Close() }()

	applied, err := database.Migrate(migrations.FS, db)
	if err != nil {
		t.Fatalf("Migrate(): %v", err)
	}
	if len(applied) < 1 || applied[0].Version != 1 {
		t.Fatalf("applied migrations unexpected: %+v", applied)
	}

	engine := routes.NewRouter(cfg, db, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Status      string `json:"status"`
			Database    string `json:"database"`
			Environment string `json:"environment"`
		} `json:"data"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	if rec.Code != http.StatusOK || !body.Success || body.Data.Status != "ok" || body.Data.Environment != config.EnvTest {
		t.Fatalf("unexpected health response: status=%d body=%+v", rec.Code, body)
	}
}
