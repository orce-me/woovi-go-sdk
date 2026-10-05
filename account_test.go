package woovi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccountsCreateGetList(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/account", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"account":{"accountId":"acc-1","taxId":"12.345.678/0001-90","isDefault":true,"balance":{"total":0,"blocked":0,"available":0}}}`))
	})
	mux.HandleFunc("GET /api/v1/account/acc-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"account":{"accountId":"acc-1","isDefault":true,"pixKeyWithdraw":{"pixKey":"saque@woovi.com","type":"EMAIL"},"balance":{"total":100,"blocked":0,"available":100}}}`))
	})
	mux.HandleFunc("GET /api/v1/account", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"account":{"accountId":"acc-1","isDefault":true}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	created, err := client.Accounts.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if created.AccountID != "acc-1" {
		t.Fatalf("accountId=%q", created.AccountID)
	}

	got, err := client.Accounts.Get(context.Background(), "acc-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.PixKeyWithdraw == nil || got.PixKeyWithdraw.PixKey != "saque@woovi.com" {
		t.Fatalf("withdraw=%+v", got.PixKeyWithdraw)
	}
	if got.Balance == nil || got.Balance.Available != 100 {
		t.Fatalf("balance=%+v", got.Balance)
	}

	list, err := client.Accounts.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].AccountID != "acc-1" {
		t.Fatalf("list=%+v", list)
	}
}
