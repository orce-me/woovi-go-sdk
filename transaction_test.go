package woovi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestTransactionsGetAndListAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/transaction/E123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"transaction":{"value":100,"type":"PAYMENT","endToEndId":"E123","transactionID":"tx-1"}}`))
	})
	mux.HandleFunc("GET /api/v1/transaction", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("charge") != "c1" {
			t.Errorf("charge=%q", r.URL.Query().Get("charge"))
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("skip") {
		case "0":
			_, _ = w.Write([]byte(`{"transactions":[{"value":1,"endToEndId":"E1"},{"value":2,"endToEndId":"E2"}],"pageInfo":{"skip":0,"limit":2,"totalCount":3,"hasNextPage":true}}`))
		case "2":
			_, _ = w.Write([]byte(`{"transactions":[{"value":3,"endToEndId":"E3"}],"pageInfo":{"skip":2,"limit":2,"totalCount":3,"hasNextPage":false}}`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)

	tx, err := client.Transactions.Get(context.Background(), "E123")
	if err != nil {
		t.Fatal(err)
	}
	if tx.EndToEndID != "E123" || tx.Type != woovi.TransactionTypePayment {
		t.Fatalf("tx=%+v", tx)
	}

	var ids []string
	for item, err := range client.Transactions.ListAll(context.Background(), &woovi.TransactionListParams{
		Limit:  2,
		Charge: "c1",
	}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, item.EndToEndID)
	}
	if len(ids) != 3 || ids[2] != "E3" {
		t.Fatalf("ids=%v", ids)
	}
}
