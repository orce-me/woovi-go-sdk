package woovi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGiftbackBalance(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/giftback/balance/{taxID}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balance":2500,"status":"AVAILABLE"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	bal, err := client.Giftbacks.Balance(context.Background(), "31324227036")
	if err != nil || bal.Balance != 2500 || bal.Status != "AVAILABLE" {
		t.Fatalf("bal=%+v err=%v", bal, err)
	}
}

func TestAuthValidateToken(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/validate-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"isValid":true,"scopes":["CHARGE_GET","GIFTBACK_BALANCE_GET"]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	got, err := client.Auth.ValidateToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.OK() || len(got.Scopes) != 2 {
		t.Fatalf("got=%+v", got)
	}
}
