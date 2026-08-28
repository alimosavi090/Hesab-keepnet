package routes_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/config"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/routes"
	"github.com/gin-gonic/gin"
)

const (
	testOrigin = "http://localhost:3000"
	testSecret = "0123456789abcdef0123456789abcdef"
)

type envelope struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newConfig(env string) *config.Config {
	return &config.Config{
		AppEnv:        env,
		AppPort:       8080,
		DatabasePath:  filepath.Join("data", "unused.db"),
		CorsOrigins:   []string{testOrigin},
		SessionSecret: testSecret,
	}
}

func newServer(t *testing.T, cfg *config.Config) (*gin.Engine, *database.DB) {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "router-test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return routes.NewRouter(cfg, db, slog.Default()), db
}

func do(t *testing.T, engine http.Handler, method, target string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var body envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	return rec, body
}

func TestHealthOK(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))

	rec, body := do(t, server, http.MethodGet, "/health")
	if rec.Code != http.StatusOK || !body.Success || body.Error != nil {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	if body.Data["status"] != "ok" || body.Data["database"] != "up" {
		t.Errorf("unexpected data: %+v", body.Data)
	}
	for _, key := range []string{"environment", "version"} {
		if _, ok := body.Data[key]; !ok {
			t.Errorf("health data missing %q", key)
		}
	}
}

func TestHealthDegradedWhenDBClosed(t *testing.T) {
	cfg := newConfig(config.EnvProduction)
	server, db := newServer(t, cfg)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	rec, body := do(t, server, http.MethodGet, "/health")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if body.Success || body.Error == nil || body.Error.Code != httpx.CodeDatabase {
		t.Fatalf("unexpected error envelope: %+v", body)
	}
	if body.Data["database"] != "down" {
		t.Errorf("database field = %v, want down", body.Data["database"])
	}
}

func TestUnknownRouteReturnsEnvelope(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))

	rec, body := do(t, server, http.MethodGet, "/definitely/not/here")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if body.Success || body.Error == nil || body.Error.Code != httpx.CodeRouteNotFound {
		t.Fatalf("unexpected envelope: %+v", body)
	}
}

func TestMethodNotAllowedReturnsEnvelope(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))

	rec, body := do(t, server, http.MethodPost, "/health")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if body.Error == nil || body.Error.Code != httpx.CodeMethodNotAllowed {
		t.Fatalf("unexpected envelope: %+v", body)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "client-provided_id-123456")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Request-ID"); got != "client-provided_id-123456" {
		t.Errorf("valid incoming id not honored: %q", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
	req2.Header.Set("X-Request-ID", "<script>bad</script>")
	rec2 := httptest.NewRecorder()
	server.ServeHTTP(rec2, req2)
	got := rec2.Header().Get("X-Request-ID")
	if got == "<script>bad</script>" || len(got) < 32 {
		t.Errorf("unsafe id accepted or generated id too short: %q", got)
	}

	rec3, _ := do(t, server, http.MethodGet, "/health")
	if rec3.Header().Get("X-Request-ID") == got {
		t.Error("generated request ids must be unique per request")
	}
}

func TestCORSPreflightAllowedOrigin(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, testOrigin)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want true", got)
	}
}

func TestCORSPreflightRejectedOrigin(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin must be empty for unknown origin, got %q", got)
	}
}

func TestSecurityHeadersPresent(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))

	rec, _ := do(t, server, http.MethodGet, "/health")

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, value := range want {
		if got := rec.Header().Get(header); got != value {
			t.Errorf("%s = %q, want %q", header, got, value)
		}
	}
}

func TestPanicRecoveredWithDevDetail(t *testing.T) {
	httpx.SetProduction(false)
	server, _ := newServer(t, newConfig(config.EnvDevelopment))
	server.GET("/panic", func(c *gin.Context) { panic("boom") })

	rec, body := do(t, server, http.MethodGet, "/panic")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if body.Success || body.Error == nil || body.Error.Code != httpx.CodeInternal {
		t.Fatalf("unexpected envelope: %+v", body)
	}
	if body.Error.Message == "خطای داخلی سرور. لطفاً بعداً تلاش کنید." {
		t.Error("dev mode must include panic detail in message")
	}
}

func TestPanicSanitizedInProduction(t *testing.T) {
	httpx.SetProduction(true)
	server, _ := newServer(t, newConfig(config.EnvProduction))
	server.GET("/panic", func(c *gin.Context) { panic("boom with secrets") })

	rec, body := do(t, server, http.MethodGet, "/panic")
	if rec.Code != http.StatusInternalServerError || body.Error == nil {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	if body.Error.Message != "خطای داخلی سرور. لطفاً بعداً تلاش کنید." {
		t.Errorf("production must sanitize internal errors, got %q", body.Error.Message)
	}
}
