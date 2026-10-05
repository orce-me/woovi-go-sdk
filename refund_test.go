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

func TestRefundsCreateGet(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/refund", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["transactionEndToEndId"] != "E123" || in["value"].(float64) != 100 {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"refund":{"status":"IN_PROCESSING","value":100,"correlationID":"r1","endToEndId":"D123"}}`))
	})
	mux.HandleFunc("GET /api/v1/refund/r1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"refund":{"status":"CONFIRMED","value":100,"correlationID":"r1"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	created, err := client.Refunds.Create(context.Background(), &woovi.RefundCreateParams{
		CorrelationID:         "r1",
		TransactionEndToEndID: "E123",
		Value:                 woovi.Cents(100),
		Comment:               "teste",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != woovi.RefundStatusInProcessing {
		t.Fatalf("status=%s", created.Status)
	}

	got, err := client.Refunds.Get(context.Background(), "r1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != woovi.RefundStatusConfirmed {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestChargeRefundAndList(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/charge/charge-1/refund", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["correlationID"] != "refund-1" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"refund":{"status":"IN_PROCESSING","value":50,"correlationID":"refund-1"}}`))
	})
	mux.HandleFunc("GET /api/v1/charge/charge-1/refund", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"refunds":[{"status":"CONFIRMED","value":50,"correlationID":"refund-1"}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	value := woovi.Cents(50)
	refund, err := client.Charges.Refund(context.Background(), "charge-1", &woovi.ChargeRefundCreateParams{
		CorrelationID: "refund-1",
		Value:         &value,
	})
	if err != nil {
		t.Fatal(err)
	}
	if refund.Value != 50 {
		t.Fatalf("value=%d", refund.Value)
	}

	list, err := client.Charges.ListRefunds(context.Background(), "charge-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].CorrelationID != "refund-1" {
		t.Fatalf("list=%+v", list)
	}
}
