package tests

import (
	"net/http"
	"testing"
)

func TestDbgBind(t *testing.T) {
	client := newEnv(t)
	if s := client.login("admin", "test-passphrase-123"); s != 200 {
		t.Fatalf("login=%d", s)
	}
	status, body := client.do(http.MethodPost, "/api/v1/categories", map[string]any{
		"name": "dbg-cat",
		"type": "BUSINESS",
	})
	t.Logf("category: %d %v", status, body)

	status, body = client.do(http.MethodPost, "/api/v1/bank-accounts", map[string]any{
		"name": "dbg-acc", "bank_name": "b", "currency": "TOMAN",
	})
	t.Logf("account: %d %v", status, body)
}
