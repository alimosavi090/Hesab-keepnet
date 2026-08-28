package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidDevelopmentDefaults(t *testing.T) {
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != EnvDevelopment {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, EnvDevelopment)
	}
	if cfg.AppPort != 8080 {
		t.Errorf("AppPort = %d, want 8080", cfg.AppPort)
	}
	if cfg.DatabasePath != "./data/accounting.db" {
		t.Errorf("DatabasePath = %q", cfg.DatabasePath)
	}
	if len(cfg.CorsOrigins) != 1 || cfg.CorsOrigins[0] != "http://localhost:3000" {
		t.Errorf("CorsOrigins = %v, want [http://localhost:3000]", cfg.CorsOrigins)
	}
	if cfg.CookieSecure {
		t.Error("CookieSecure should default to false in development")
	}
	if cfg.SessionSecretEphemeral {
		t.Error("SessionSecretEphemeral must be false when SESSION_SECRET is provided")
	}
}

func TestLoadValidFullProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DATABASE_PATH", "/tmp/prod.db")
	t.Setenv("CORS_ORIGIN", " https://a.example.com , https://b.example.com/ ")
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("COOKIE_SECURE", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppPort != 9090 {
		t.Errorf("AppPort = %d, want 9090", cfg.AppPort)
	}
	want := []string{"https://a.example.com", "https://b.example.com"}
	if len(cfg.CorsOrigins) != len(want) {
		t.Fatalf("CorsOrigins = %v, want %v", cfg.CorsOrigins, want)
	}
	for i := range want {
		if cfg.CorsOrigins[i] != want[i] {
			t.Errorf("CorsOrigins[%d] = %q, want %q", i, cfg.CorsOrigins[i], want[i])
		}
	}
	if !cfg.CookieSecure {
		t.Error("CookieSecure = false, want true")
	}
	if !cfg.IsProd() || cfg.Addr() != ":9090" {
		t.Errorf("IsProd/Addr unexpected: %v / %q", cfg.IsProd(), cfg.Addr())
	}
}

func TestLoadProductionDefaultsCookieSecure(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if !cfg.CookieSecure {
		t.Error("CookieSecure must default to true in production")
	}
}

func TestLoadMissingSessionSecretInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	os.Unsetenv("SESSION_SECRET")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "SESSION_SECRET") {
		t.Fatalf("expected SESSION_SECRET error, got %v", err)
	}
}

func TestLoadShortSessionSecret(t *testing.T) {
	t.Setenv("SESSION_SECRET", "short")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "too short") {
		t.Fatalf("expected too-short error, got %v", err)
	}
}

func TestLoadInvalidPorts(t *testing.T) {
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")

	for _, port := range []string{"abc", "0", "-1", "70000"} {
		t.Setenv("APP_PORT", port)
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "APP_PORT") {
			t.Fatalf("APP_PORT=%q: expected invalid-port error, got %v", port, err)
		}
	}
}

func TestLoadInvalidAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("expected APP_ENV error, got %v", err)
	}
}

func TestLoadEphemeralSecretInDevelopment(t *testing.T) {
	os.Unsetenv("SESSION_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if !cfg.SessionSecretEphemeral || len(cfg.SessionSecret) < minSecretLength {
		t.Fatalf("expected ephemeral secret, got ephemeral=%v len=%d", cfg.SessionSecretEphemeral, len(cfg.SessionSecret))
	}
}

func TestLoadDotenv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\nFOO_BAR=hello\nQUOTED=\"wrapped value\"\nEMPTY=\nBROKEN LINE\nexport EXPORTED=yes\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EXISTING", "keep-me")

	if err := loadDotenv(path); err != nil {
		t.Fatalf("loadDotenv() error: %v", err)
	}
	if os.Getenv("FOO_BAR") != "hello" {
		t.Errorf("FOO_BAR = %q", os.Getenv("FOO_BAR"))
	}
	if os.Getenv("QUOTED") != "wrapped value" {
		t.Errorf("QUOTED = %q", os.Getenv("QUOTED"))
	}
	if os.Getenv("EXPORTED") != "yes" {
		t.Errorf("EXPORTED = %q", os.Getenv("EXPORTED"))
	}
	if _, ok := os.LookupEnv("EMPTY"); ok {
		t.Error("empty values must not be set")
	}
}

func TestLoadDotenvMissingFile(t *testing.T) {
	if err := loadDotenv(filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("missing .env must be ignored, got %v", err)
	}
}
