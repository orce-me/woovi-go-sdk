package woovi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestSubaccountsCRUDWithdrawTransfer(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/subaccount", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["name"] != "seller-1" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"seller-1","pixKey":"seller@woovi.com"}`))
	})
	mux.HandleFunc("GET /api/v1/subaccount/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "seller@woovi.com" {
			t.Errorf("id=%q", r.PathValue("id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"SubAccount":{"name":"seller-1","pixKey":"seller@woovi.com","balance":7000}}`))
	})
	mux.HandleFunc("POST /api/v1/subaccount/{id}/withdraw", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["value"].(float64) != 7000 {
			t.Errorf("withdraw body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"transaction":{"status":"CREATED","value":7000,"correlationID":"w1","destinationAlias":"seller@woovi.com"}}`))
	})
	mux.HandleFunc("POST /api/v1/subaccount/transfer", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["fromPixKey"] != "a@x.com" || in["toPixKey"] != "b@x.com" {
			t.Errorf("transfer body=%s", body)
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("DELETE /api/v1/subaccount/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/v1/subaccount", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subAccounts":[{"name":"seller-1","pixKey":"seller@woovi.com","balance":0}],"pageInfo":{"skip":0,"limit":100,"totalCount":1,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)

	created, err := client.Subaccounts.Create(context.Background(), &woovi.SubaccountCreateParams{
		Name:   "seller-1",
		PixKey: "seller@woovi.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.PixKey != "seller@woovi.com" {
		t.Fatalf("created=%+v", created)
	}

	got, err := client.Subaccounts.Get(context.Background(), "seller@woovi.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.Balance != 7000 {
		t.Fatalf("balance=%d", got.Balance)
	}

	value := woovi.Cents(7000)
	withdraw, err := client.Subaccounts.Withdraw(context.Background(), "seller@woovi.com", &woovi.SubaccountWithdrawParams{
		Value: &value,
	})
	if err != nil {
		t.Fatal(err)
	}
	if withdraw.Transaction.Value != 7000 {
		t.Fatalf("withdraw=%+v", withdraw)
	}

	if err := client.Subaccounts.Transfer(context.Background(), &woovi.SubaccountTransferParams{
		Value:      woovi.Cents(65),
		FromPixKey: "a@x.com",
		ToPixKey:   "b@x.com",
	}); err != nil {
		t.Fatal(err)
	}

	list, err := client.Subaccounts.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.SubAccounts) != 1 {
		t.Fatalf("list=%+v", list)
	}

	if err := client.Subaccounts.Delete(context.Background(), "seller@woovi.com"); err != nil {
		t.Fatal(err)
	}
}
