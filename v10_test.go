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

func TestPixAuthCreateGet(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/pix-auth", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["correlationID"] != "signup-1" || in["taxID"] != "12345678909" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixAuth":{"id":"pa1","correlationID":"signup-1","status":"ACTIVE","result":"UNVERIFIED","amount":1},"brCode":"000201","hostedUrl":"https://woovi.com/pix-auth/signup-1"}`))
	})
	mux.HandleFunc("GET /api/v1/pix-auth/signup-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixAuth":{"id":"pa1","correlationID":"signup-1","status":"COMPLETED","result":"MATCHED","amount":1}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	created, err := client.PixAuths.Create(context.Background(), &woovi.PixAuthCreateParams{
		CorrelationID: "signup-1",
		TaxID:         "12345678909",
	})
	if err != nil || created.PixAuth.Status != woovi.PixAuthStatusActive || created.BrCode == "" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	got, err := client.PixAuths.Get(context.Background(), "signup-1")
	if err != nil || got.PixAuth.Result != woovi.PixAuthResultMatched {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestAccountRegisterGetDelete(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/account-register/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"officialName":"Acme","tradeName":"Acme","correlationID":"reg-1","status":"APPROVED"}`))
	})
	mux.HandleFunc("DELETE /api/v1/account-register/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"deleted","accountRegisterId":"reg-1"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	got, err := client.AccountRegisters.Get(context.Background(), "reg-1")
	if err != nil || got.OfficialName != "Acme" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	del, err := client.AccountRegisters.Delete(context.Background(), "reg-1")
	if err != nil || del.AccountRegisterID != "reg-1" {
		t.Fatalf("del=%+v err=%v", del, err)
	}
}

func TestInvoiceCreateCertificatePatch(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/invoice", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["charge"] != "charge-1" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoice":{"id":"inv1","value":1000,"correlationID":"inv-1","status":"CONFIRMED"}}`))
	})
	mux.HandleFunc("POST /api/v1/invoice/integration/certificate", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["pcks12"] == nil || in["passphrase"] == nil {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"integration":{"status":"CONFIGURED"}}`))
	})
	mux.HandleFunc("PATCH /api/v1/invoice/integration", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["isActive"] != true {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"integration":{"id":"i1","status":"CONFIGURED","isActive":true}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	inv, err := client.Invoices.Create(context.Background(), &woovi.InvoiceCreateParams{
		Charge:        "charge-1",
		CorrelationID: "inv-1",
	})
	if err != nil || inv.ID != "inv1" {
		t.Fatalf("inv=%+v err=%v", inv, err)
	}
	cert, err := client.Invoices.UploadCertificate(context.Background(), &woovi.InvoiceCertificateParams{
		Pcks12:     "Y2VydA==",
		Passphrase: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cert["integration"] == nil {
		t.Fatalf("cert=%v", cert)
	}
	patched, err := client.Invoices.PatchIntegration(context.Background(), map[string]any{"isActive": true})
	if err != nil {
		t.Fatal(err)
	}
	if patched["integration"] == nil {
		t.Fatalf("patched=%v", patched)
	}
}

func TestBoletoListTransactions(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/boleto-transaction", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "BOLETO_IN" {
			t.Errorf("q=%s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"OK","pageInfo":{"skip":0,"limit":100,"hasNextPage":false},"boletoTransactions":[{"boletoTransactionID":"btx1","type":"BOLETO_IN","status":"CONFIRMED","value":245000}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	list, err := client.Boletos.ListTransactions(context.Background(), &woovi.BoletoTransactionListParams{
		Type: woovi.BoletoTransactionTypeIn,
	})
	if err != nil || len(list.BoletoTransactions) != 1 || list.BoletoTransactions[0].BoletoTransactionID != "btx1" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
}
