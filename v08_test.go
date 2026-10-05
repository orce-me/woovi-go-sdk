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

func TestTEDCreateGetListRefund(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/ted", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["correlationID"] != "payout-1" || in["value"].(float64) != 150050 {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ted":{"correlationID":"payout-1","status":"PROCESSING","type":"PAYMENT","direction":"OUT","value":150050}}`))
	})
	mux.HandleFunc("GET /api/v1/ted/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ted":{"correlationID":"payout-1","status":"COMPLETED","value":150050}}`))
	})
	mux.HandleFunc("GET /api/v1/ted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"teds":[{"correlationID":"payout-1","status":"COMPLETED"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	mux.HandleFunc("POST /api/v1/ted/{id}/refund", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ted":{"correlationID":"refund-1","type":"REFUND_SENT","direction":"OUT","status":"PROCESSING"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	created, err := client.TEDs.Create(context.Background(), &woovi.TEDCreateParams{
		CorrelationID: "payout-1",
		Value:         woovi.Cents(150050),
		AccountID:     "acc-1",
		Receiver: woovi.TEDParty{
			Name:        "Joao",
			Document:    "12345678901",
			ISPB:        "87654321",
			Agency:      4321,
			Account:     98765,
			AccountType: woovi.TEDAccountTypeCACC,
		},
	})
	if err != nil || created.Status != woovi.TEDStatusProcessing {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	got, err := client.TEDs.Get(context.Background(), "payout-1")
	if err != nil || got.Status != woovi.TEDStatusCompleted {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	list, err := client.TEDs.List(context.Background(), nil)
	if err != nil || len(list.TEDs) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	refund, err := client.TEDs.Refund(context.Background(), "in-ted-1")
	if err != nil || refund.Type != woovi.TEDTypeRefundSent {
		t.Fatalf("refund=%+v err=%v", refund, err)
	}
}

func TestReceiptFundsRecoveryInvoiceSubaccount(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/receipt/{type}/{e2e}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.4"))
	})
	mux.HandleFunc("POST /api/v1/funds-recovery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dictId":"fr-1","status":"CREATED","situationType":"SCAM","rootTransactionId":"E1"}`))
	})
	mux.HandleFunc("GET /api/v1/funds-recovery/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dictId":"fr-1","status":"CREATED"}`))
	})
	mux.HandleFunc("POST /api/v1/funds-recovery/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dictId":"fr-1","status":"CANCELLED"}`))
	})
	mux.HandleFunc("GET /api/v1/funds-recovery/{id}/disputes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"d1","type":"MED","status":"OPENED","value":50000}]`))
	})
	mux.HandleFunc("GET /api/v1/invoice", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoices":[{"correlationID":"INV-1","value":500,"status":"CONFIRMED"}]}`))
	})
	mux.HandleFunc("POST /api/v1/invoice/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	})
	mux.HandleFunc("GET /api/v1/subaccount/{id}/statement", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1","time":"2023-12-01T10:30:00.000Z","value":100,"balance":1500,"type":"CREDIT"}]`))
	})
	mux.HandleFunc("GET /api/v1/boleto-transaction/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"boletoTransaction":{"boletoTransactionID":"btx_1","type":"BOLETO_IN","value":245000}}`))
	})
	mux.HandleFunc("GET /api/image/qrcode/base64/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"imageBase64":"data:image/png;base64,abc"}`))
	})
	mux.HandleFunc("POST /api/v1/partner/company", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"preRegistration":{"name":"Example LLC","taxID":{"taxID":"11111111111111","type":"BR:CNPJ"}},"user":{"firstName":"John","email":"a@b.com"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	receipt, err := client.Receipts.Get(context.Background(), woovi.ReceiptTypePixIn, "E123")
	if err != nil || string(receipt.Data) != "%PDF-1.4" {
		t.Fatalf("receipt=%+v err=%v", receipt, err)
	}
	fr, err := client.FundsRecoveries.Create(context.Background(), &woovi.FundsRecoveryCreateParams{
		TransactionEndToEndID: "E1",
		SituationType:         woovi.FundsRecoverySituationScam,
		Details:               "fake seller",
	})
	if err != nil || fr.DictID != "fr-1" {
		t.Fatalf("fr=%+v err=%v", fr, err)
	}
	got, err := client.FundsRecoveries.Get(context.Background(), "fr-1")
	if err != nil || got.DictID != "fr-1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	canceled, err := client.FundsRecoveries.Cancel(context.Background(), "fr-1")
	if err != nil || canceled.Status != woovi.FundsRecoveryStatusCancelled {
		t.Fatalf("canceled=%+v err=%v", canceled, err)
	}
	disputes, err := client.FundsRecoveries.ListDisputes(context.Background(), "fr-1")
	if err != nil || len(disputes) != 1 {
		t.Fatalf("disputes=%+v err=%v", disputes, err)
	}
	invoices, err := client.Invoices.List(context.Background(), nil)
	if err != nil || len(invoices.Invoices) != 1 {
		t.Fatalf("invoices=%+v err=%v", invoices, err)
	}
	if err := client.Invoices.Cancel(context.Background(), "INV-1"); err != nil {
		t.Fatal(err)
	}
	stmt, err := client.Subaccounts.Statement(context.Background(), "seller@woovi.com", nil)
	if err != nil || len(stmt.Statements) != 1 {
		t.Fatalf("stmt=%+v err=%v", stmt, err)
	}
	btx, err := client.Boletos.GetTransaction(context.Background(), "btx_1")
	if err != nil || btx.Value != 245000 {
		t.Fatalf("btx=%+v err=%v", btx, err)
	}
	b64, err := client.Charges.QRCodeBase64(context.Background(), "fe78", 768)
	if err != nil || b64 == "" {
		t.Fatalf("b64=%q err=%v", b64, err)
	}
	partner, err := client.Partners.CreateCompany(context.Background(), &woovi.PartnerCreateParams{
		PreRegistration: woovi.PartnerPreRegistration{
			Name:  "Example LLC",
			TaxID: woovi.PartnerTaxID{TaxID: "11111111111111", Type: "BR:CNPJ"},
		},
		User: woovi.PartnerUser{FirstName: "John", Email: "a@b.com"},
	})
	if err != nil || partner.PreRegistration.Name != "Example LLC" {
		t.Fatalf("partner=%+v err=%v", partner, err)
	}
}
