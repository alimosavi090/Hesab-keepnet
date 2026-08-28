package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/ali/hesab-keepnet/backend/internal/auth"
	"github.com/ali/hesab-keepnet/backend/internal/config"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/routes"
	"github.com/ali/hesab-keepnet/backend/migrations"
)

type apiClient struct {
	t    *testing.T
	base *httptest.Server
	http *http.Client
}

func newEnv(t *testing.T) *apiClient {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := database.Migrate(migrations.FS, db); err != nil {
		t.Fatal(err)
	}
	if err := database.Seed(t.Context(), db.DB, "admin", "test-passphrase-123"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		AppEnv:        config.EnvTest,
		AppPort:       0,
		CorsOrigins:   []string{"http://localhost:3000"},
		SessionSecret: testSecret,
	}
	router := routes.NewRouter(cfg, db, discardLogger())

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	jar, _ := cookiejar.New(nil)
	return &apiClient{t: t, base: server, http: &http.Client{Jar: jar}}
}

func (c *apiClient) do(method, path string, body any) (int, map[string]any) {
	c.t.Helper()

	var reader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, c.base.URL+path, reader)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if method != http.MethodGet && method != http.MethodHead {
		req.Header.Set(auth.CSRFHeader, c.csrf())
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()

	var decoded map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&decoded)
	return resp.StatusCode, decoded
}

func (c *apiClient) csrf() string {
	u, _ := url.Parse(c.base.URL)
	for _, cookie := range c.http.Jar.Cookies(u) {
		if cookie.Name == auth.CSRFCookie {
			return cookie.Value
		}
	}

	resp, err := c.http.Get(c.base.URL + "/api/v1/auth/csrf")
	if err != nil {
		c.t.Fatal(err)
	}
	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()

	for _, cookie := range c.http.Jar.Cookies(u) {
		if cookie.Name == auth.CSRFCookie {
			return cookie.Value
		}
	}
	c.t.Fatal("csrf cookie was not set")
	return ""
}

func (c *apiClient) login(username, password string) int {
	status, _ := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": username,
		"password": password,
	})
	return status
}

func dataMap(body map[string]any) map[string]any {
	data, _ := body["data"].(map[string]any)
	return data
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAuthFlowOverHTTP(t *testing.T) {
	client := newEnv(t)

	if status, _ := client.do(http.MethodGet, "/api/v1/auth/me", nil); status != 401 {
		t.Fatalf("unauthenticated /me = %d, want 401", status)
	}

	if status := client.login("admin", "wrong-password"); status != 401 {
		t.Fatalf("wrong password login = %d, want 401", status)
	}

	if status := client.login("admin", "test-passphrase-123"); status != 200 {
		t.Fatalf("login = %d, want 200", status)
	}

	status, body := client.do(http.MethodGet, "/api/v1/auth/me", nil)
	user := dataMap(body)
	if status != 200 || user["username"] != "admin" {
		t.Fatalf("/me = %d %v", status, user)
	}

	logoutStatus, _ := client.do(http.MethodPost, "/api/v1/auth/logout", nil)
	if logoutStatus != 200 {
		t.Fatalf("logout = %d", logoutStatus)
	}
	if status, _ := client.do(http.MethodGet, "/api/v1/auth/me", nil); status != 401 {
		t.Fatalf("/me after logout = %d, want 401", status)
	}
}

func TestCSRFEnforcement(t *testing.T) {
	client := newEnv(t)
	if status := client.login("admin", "test-passphrase-123"); status != 200 {
		t.Fatal(status)
	}

	req, _ := http.NewRequest(http.MethodPost, client.base.URL+"/api/v1/categories", bytes.NewBufferString(`{"name":"تست","type":"BUSINESS"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.CSRFHeader, "mismatched-token")
	resp, err := client.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("POST with bad CSRF = %d, want 403", resp.StatusCode)
	}

	status, _ := client.do(http.MethodPost, "/api/v1/categories", map[string]any{
		"name": "تست دسته",
		"type": "BUSINESS",
	})
	if status != 201 {
		t.Fatalf("POST with valid CSRF = %d, want 201", status)
	}
}

func TestProtectedEndpointsRequireSession(t *testing.T) {
	client := newEnv(t)

	for _, methodPath := range [][2]string{
		{http.MethodGet, "/api/v1/bank-accounts"},
		{http.MethodGet, "/api/v1/expenses"},
		{http.MethodGet, "/api/v1/sales"},
		{http.MethodGet, "/api/v1/dashboard/summary"},
	} {
		if status, _ := client.do(methodPath[0], methodPath[1], nil); status != 401 {
			t.Errorf("%s %s = %d, want 401", methodPath[0], methodPath[1], status)
		}
	}
}

func TestFinancialFlowOverHTTP(t *testing.T) {
	client := newEnv(t)
	if status := client.login("admin", "test-passphrase-123"); status != 200 {
		t.Fatal(status)
	}

	status, body := client.do(http.MethodPost, "/api/v1/bank-accounts", map[string]any{
		"name": "بانک اصلی", "bank_name": "ملی", "currency": "RIAL", "initial_balance": 10_000_000,
	})
	if status != 201 {
		t.Fatalf("create account = %d %v", status, body)
	}
	accountID := int64(dataMap(body)["id"].(float64))

	balanceStatus, balanceRaw := client.do(http.MethodGet, fmt.Sprintf("/api/v1/bank-accounts/%d/balance", accountID), nil)
	if balanceStatus != 200 {
		t.Fatalf("balance endpoint = %d", balanceStatus)
	}
	balanceBody := dataMap(balanceRaw)
	if balanceBody["balance"].(float64) != 10_000_000 {
		t.Fatalf("initial balance = %v", balanceBody["balance"])
	}

	status, _ = client.do(http.MethodPost, "/api/v1/expenses", map[string]any{
		"category_id":     businessCategoryID(t, client),
		"bank_account_id": accountID,
		"amount":          250_000,
		"currency":        "RIAL",
		"occurred_at":     "2026-08-01T12:00:00Z",
	})
	if status != 201 {
		t.Fatalf("create expense = %d", status)
	}

	listStatus, listBody := client.do(http.MethodGet, "/api/v1/expenses?page=1&page_size=10&from=2026-07-01&to=2026-08-31", nil)
	items, _ := listBody["data"].(map[string]any)["items"].([]any)
	if listStatus != 200 || len(items) != 1 {
		t.Fatalf("expenses list = %d with %d items", listStatus, len(items))
	}

	chartStatus, chartBody := client.do(http.MethodGet, "/api/v1/dashboard/chart?days=7", nil)
	points, _ := chartBody["data"].([]any)
	if chartStatus != 200 || len(points) != 7 {
		t.Fatalf("chart = %d with %d points, want 7", chartStatus, len(points))
	}

	summaryStatus, summaryBody := client.do(http.MethodGet, "/api/v1/dashboard/summary", nil)
	summary := dataMap(summaryBody)
	banks, _ := summary["banks"].([]any)
	if summaryStatus != 200 || len(banks) != 1 {
		t.Fatalf("summary = %d banks=%d", summaryStatus, len(banks))
	}

	csvResp, err := client.http.Get(client.base.URL + "/api/v1/reports/export.csv?dataset=expenses&from=2026-07-01&to=2026-08-31")
	if err != nil {
		t.Fatal(err)
	}
	defer csvResp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(csvResp.Body)
	content := buf.String()
	if !bytes.HasPrefix([]byte(content), []byte("\uFEFF")) || len(content) < 50 {
		t.Fatalf("csv export invalid (len=%d)", len(content))
	}

	transferStatus, _ := client.do(http.MethodPost, "/api/v1/transfers", map[string]any{
		"from_account_id": accountID,
		"to_account_id":   createSecondAccount(t, client),
		"amount":          500_000,
		"currency":        "RIAL",
		"transferred_at":  "2026-08-02T12:00:00Z",
	})
	if transferStatus != 201 {
		t.Fatalf("create transfer = %d", transferStatus)
	}

	feedStatus, feedBody := client.do(http.MethodGet, "/api/v1/transactions?page_size=20", nil)
	feedItems, _ := feedBody["data"].(map[string]any)["items"].([]any)
	total := dataMap(feedBody)["meta"].(map[string]any)["total"].(float64)
	if feedStatus != 200 || total != 3 {
		t.Fatalf("ledger feed = %d total=%v items=%d, want 3 rows (expense+2 transfers)", feedStatus, total, len(feedItems))
	}
}

func createSecondAccount(t *testing.T, client *apiClient) float64 {
	t.Helper()
	status, body := client.do(http.MethodPost, "/api/v1/bank-accounts", map[string]any{
		"name": "حساب دوم", "bank_name": "صادرات", "currency": "RIAL",
	})
	if status != 201 {
		t.Fatalf("second account = %d %v", status, body)
	}
	return dataMap(body)["id"].(float64)
}

func businessCategoryID(t *testing.T, client *apiClient) float64 {
	t.Helper()
	status, body := client.do(http.MethodGet, "/api/v1/categories?type=BUSINESS", nil)
	if status != 200 {
		t.Fatal(status)
	}
	items, _ := body["data"].([]any)
	if len(items) == 0 {
		t.Fatal("no seeded business categories")
	}
	first := items[0].(map[string]any)
	return first["id"].(float64)
}

const testSecret = "0123456789abcdef0123456789abcdef"
