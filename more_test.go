package woovi_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestCashbackCreateValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	ctx := context.Background()
	if _, err := client.CashbackFidelities.Create(ctx, nil); err == nil {
		t.Fatal("nil params")
	}
	if _, err := client.CashbackFidelities.Create(ctx, &woovi.CashbackFidelityCreateParams{Value: 1}); err == nil {
		t.Fatal("empty taxID")
	}
	if _, err := client.CashbackFidelities.Create(ctx, &woovi.CashbackFidelityCreateParams{TaxID: "1"}); err == nil {
		t.Fatal("zero value")
	}
	if _, err := client.CashbackFidelities.Balance(ctx, ""); err == nil {
		t.Fatal("empty balance taxID")
	}
}

func TestChargeCreateValidationAndListAllPages(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	ctx := context.Background()
	if _, err := client.Charges.Create(ctx, nil); err == nil {
		t.Fatal("nil")
	}
	if _, err := client.Charges.Create(ctx, &woovi.ChargeCreateParams{Value: 1}); err == nil {
		t.Fatal("empty correlation")
	}
	if _, err := client.Charges.Create(ctx, &woovi.ChargeCreateParams{CorrelationID: "c"}); err == nil {
		t.Fatal("zero value")
	}
	if _, err := client.Charges.Get(ctx, ""); err == nil {
		t.Fatal("empty get")
	}

	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/charge", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if calls.Add(1) == 1 {
			_, _ = w.Write([]byte(`{"charges":[{"correlationID":"a","value":1}],"pageInfo":{"skip":0,"limit":1,"hasNextPage":true}}`))
			return
		}
		_, _ = w.Write([]byte(`{"charges":[{"correlationID":"b","value":2}],"pageInfo":{"skip":1,"limit":1,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := mustClient(t, srv)
	n := 0
	for ch, err := range c.Charges.ListAll(ctx, &woovi.ChargeListParams{Limit: 1, Status: woovi.ChargeStatusActive}) {
		if err != nil {
			t.Fatal(err)
		}
		if ch == nil {
			t.Fatal("nil charge")
		}
		n++
	}
	if n != 2 || calls.Load() != 2 {
		t.Fatalf("n=%d calls=%d", n, calls.Load())
	}
}

func TestReceiptValidationAndDefaultContentType(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	if _, err := client.Receipts.Get(context.Background(), "", "E1"); err == nil {
		t.Fatal("empty type")
	}
	if _, err := client.Receipts.Get(context.Background(), woovi.ReceiptTypePixOut, ""); err == nil {
		t.Fatal("empty e2e")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/receipt/{type}/{e2e}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("%PDF"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	got, err := mustClient(t, srv).Receipts.Get(context.Background(), woovi.ReceiptTypePixRefund, "E99")
	if err != nil || string(got.Data) != "%PDF" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.ContentType == "" {
		t.Fatal("empty content type")
	}
}

func TestDoBytesNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"missing"}`))
	}))
	t.Cleanup(srv.Close)
	_, err := mustClient(t, srv).Charges.BRCodeImage(context.Background(), "missing", 0)
	if !errors.Is(err, woovi.ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
	var api *woovi.APIError
	if !errors.As(err, &api) {
		t.Fatalf("api=%v", err)
	}
}

func TestListAllCancelledContext(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/customer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"customers":[{"name":"A","correlationID":"c1"}],"pageInfo":{"skip":0,"limit":1,"hasNextPage":true}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, err := range client.Customers.ListAll(ctx, &woovi.CustomerListParams{Limit: 1}) {
		if err == nil {
			t.Fatal("expected cancelled")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err=%v", err)
		}
		return
	}
	t.Fatal("expected at least one yield")
}

func TestDisputeEvidenceValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	ctx := context.Background()
	if _, err := client.Disputes.Get(ctx, ""); err == nil {
		t.Fatal("empty get")
	}
	if _, err := client.Disputes.AddEvidence(ctx, "", &woovi.DisputeEvidenceParams{
		Documents: []woovi.DisputeEvidenceDocument{{FileID: "f"}},
	}); err == nil {
		t.Fatal("empty id")
	}
	if _, err := client.Disputes.AddEvidence(ctx, "d1", nil); err == nil {
		t.Fatal("nil docs")
	}
	if _, err := client.Disputes.AddEvidence(ctx, "d1", &woovi.DisputeEvidenceParams{
		Documents: []woovi.DisputeEvidenceDocument{{FileID: "f", URL: "https://x"}},
	}); err == nil || !strings.Contains(err.Error(), "url or fileId") {
		t.Fatalf("both url and fileId: %v", err)
	}
	if _, err := client.Disputes.AddEvidence(ctx, "d1", &woovi.DisputeEvidenceParams{
		Documents: []woovi.DisputeEvidenceDocument{{}},
	}); err == nil {
		t.Fatal("neither url nor fileId")
	}
}

func TestFileUploadValidationAndUnknownExt(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	ctx := context.Background()
	if _, err := client.Files.Upload(ctx, &woovi.FileUploadParams{
		File: bytes.NewReader([]byte("x")), FileName: "  ", Purpose: woovi.FilePurposeDisputeEvidence,
	}); err == nil {
		t.Fatal("blank name")
	}
	if _, err := client.Files.Upload(ctx, &woovi.FileUploadParams{
		File: bytes.NewReader([]byte("x")), FileName: "a.bin", Purpose: woovi.FilePurposeDisputeEvidence,
	}); err == nil || !strings.Contains(err.Error(), "contentType") {
		t.Fatalf("unknown ext: %v", err)
	}
	if _, err := client.Files.Upload(ctx, &woovi.FileUploadParams{
		File: bytes.NewReader([]byte("x")), FileName: "a.pdf",
	}); err == nil {
		t.Fatal("missing purpose")
	}
}

func TestBoletoTransactionValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	if _, err := client.Boletos.GetTransaction(context.Background(), ""); err == nil {
		t.Fatal("empty id")
	}
}

func TestBoletoBarcodeNonDigit(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	_, err := client.Boletos.Validate(context.Background(), "1234567890123456789012345678901234567890123a")
	if err == nil || !strings.Contains(err.Error(), "digits") {
		t.Fatalf("err=%v", err)
	}
	_, err = client.Boletos.Validate(context.Background(), "")
	if err == nil {
		t.Fatal("empty barcode")
	}
}

func TestBoletoListTransactionsFilters(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/boleto-transaction", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for _, k := range []string{"type", "status", "start", "end", "settledStart", "settledEnd"} {
			if q.Get(k) == "" {
				t.Errorf("missing %s", k)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"boletoTransactions":[],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	list, err := mustClient(t, srv).Boletos.ListTransactions(context.Background(), &woovi.BoletoTransactionListParams{
		Type:         woovi.BoletoTransactionTypeIn,
		Status:       woovi.BoletoTransactionStatusConfirmed,
		Start:        "2026-01-01T00:00:00.000Z",
		End:          "2026-01-31T00:00:00.000Z",
		SettledStart: "2026-01-01T00:00:00.000Z",
		SettledEnd:   "2026-01-31T00:00:00.000Z",
	})
	if err != nil || list == nil {
		t.Fatalf("list=%+v err=%v", list, err)
	}
}

func TestDecodeEMVValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	if _, err := client.Decode.EMV(context.Background(), ""); err == nil {
		t.Fatal("empty emv")
	}
}

func TestInstallmentAndPixAuthValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	ctx := context.Background()
	if _, err := client.Installments.Get(ctx, ""); err == nil {
		t.Fatal("installment get")
	}
	if _, err := client.Installments.CreateCobr(ctx, "", nil); err == nil {
		t.Fatal("cobr empty id")
	}
	if _, err := client.PixAuths.Create(ctx, nil); err == nil {
		t.Fatal("pixauth nil")
	}
	if _, err := client.PixAuths.Create(ctx, &woovi.PixAuthCreateParams{TaxID: "1"}); err == nil {
		t.Fatal("pixauth missing corr")
	}
	if _, err := client.PixAuths.Get(ctx, ""); err == nil {
		t.Fatal("pixauth get")
	}
}

func TestAccountsListShapes(t *testing.T) {
	t.Parallel()
	t.Run("single", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/account", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"account":{"accountId":"a1","status":"OPENED"}}`))
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		list, err := mustClient(t, srv).Accounts.List(context.Background())
		if err != nil || len(list) != 1 || list[0].AccountID != "a1" {
			t.Fatalf("list=%+v err=%v", list, err)
		}
	})
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/account", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		list, err := mustClient(t, srv).Accounts.List(context.Background())
		if err != nil || len(list) != 0 {
			t.Fatalf("list=%+v err=%v", list, err)
		}
	})
	client := mustClient(t, unusedServer(t))
	if _, err := client.Accounts.Get(context.Background(), ""); err == nil {
		t.Fatal("empty get")
	}
}

func TestPSPListAllFilters(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/psp", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("ispb") != "00000000" || q.Get("name") != "BB" || q.Get("compe") != "001" {
			t.Errorf("q=%s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"psps":[{"name":"BB","ispb":"00000000","compe":"001"}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	psps, err := mustClient(t, srv).PSPs.List(context.Background(), &woovi.PSPListParams{
		ISPB: "00000000", Name: "BB", Compe: "001",
	})
	if err != nil || len(psps) != 1 {
		t.Fatalf("psps=%+v err=%v", psps, err)
	}
}

func TestLimitsCreateRequestValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	ctx := context.Background()
	if _, err := client.Limits.Get(ctx, ""); err == nil {
		t.Fatal("empty account")
	}
	if _, err := client.Limits.CreateRequest(ctx, nil); err == nil {
		t.Fatal("nil create")
	}
	if _, err := client.Limits.GetRequest(ctx, ""); err == nil {
		t.Fatal("empty get request")
	}
	if _, err := client.Limits.ListRequests(ctx, &woovi.LimitRequestListParams{Limit: 101}); err == nil {
		t.Fatal("limit too high")
	}
}

func TestInvoiceListAllCancelled(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/invoice", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoices":[{"correlationID":"I1","value":1}],"pageInfo":{"skip":0,"limit":1,"hasNextPage":true}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, err := range mustClient(t, srv).Invoices.ListAll(ctx, &woovi.InvoiceListParams{Limit: 1}) {
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err=%v", err)
		}
		return
	}
	t.Fatal("expected yield")
}

func TestPartnerGetCompanyValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	if _, err := client.Partners.GetCompany(context.Background(), ""); err == nil {
		t.Fatal("empty taxID")
	}
}

func TestTransfersCreateValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	ctx := context.Background()
	if _, err := client.Transfers.Create(ctx, nil); err == nil {
		t.Fatal("nil transfer")
	}
	if _, err := client.Transfers.Create(ctx, &woovi.TransferCreateParams{}); err == nil {
		t.Fatal("zero value")
	}
	if _, err := client.Transfers.Create(ctx, &woovi.TransferCreateParams{Value: 1}); err == nil {
		t.Fatal("missing keys")
	}
}

func TestAccountRegisterValidation(t *testing.T) {
	t.Parallel()
	client := mustClient(t, unusedServer(t))
	if _, err := client.AccountRegisters.Get(context.Background(), ""); err == nil {
		t.Fatal("empty get")
	}
	if _, err := client.AccountRegisters.Delete(context.Background(), ""); err == nil {
		t.Fatal("empty delete")
	}
}

func unusedServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(srv.Close)
	return srv
}
