package woovi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/orce-me/woovi-go-sdk"
)

func TestHelpersAndMoneyInt64(t *testing.T) {
	t.Parallel()
	if *woovi.String("x") != "x" || *woovi.Int(3) != 3 || *woovi.Int64(9) != 9 || !*woovi.Bool(true) {
		t.Fatal("helpers")
	}
	if woovi.Cents(1990).Int64() != 1990 {
		t.Fatal("Money.Int64")
	}
}

func TestAPIErrorErrorString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  *woovi.APIError
		want string
	}{
		{"message+code+rid", &woovi.APIError{StatusCode: 400, Message: "bad", Code: "E1", RequestID: "r1"}, "woovi: bad (400/E1) [request_id=r1]"},
		{"message+code", &woovi.APIError{StatusCode: 400, Message: "bad", Code: "E1"}, "woovi: bad (400/E1)"},
		{"message+rid", &woovi.APIError{StatusCode: 500, Message: "oops", RequestID: "r2"}, "woovi: oops (500) [request_id=r2]"},
		{"from errors", &woovi.APIError{StatusCode: 422, Errors: []woovi.APIErrorItem{{Message: "field"}}}, "woovi: field (422)"},
		{"status text", &woovi.APIError{StatusCode: 418}, "woovi: I'm a teapot (418)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestClientDoAndOptions(t *testing.T) {
	t.Parallel()
	var gotUA, gotHdr, gotIdem string
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/custom", func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotHdr = r.Header.Get("X-Test")
		gotIdem = r.Header.Get("Idempotency-Key")
		if r.URL.Query().Get("q") != "1" {
			t.Errorf("query=%s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}),
		woovi.WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		woovi.WithUserAgent("cov-test"),
		woovi.WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	err = client.Do(context.Background(), http.MethodPost, "/api/v1/custom", map[string]string{"a": "b"}, &out,
		woovi.WithHeader("X-Test", "1"),
		woovi.WithQuery("q", "1"),
		woovi.WithIdempotencyKey("idem-1"),
		woovi.WithRequestTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	if out["ok"] != true {
		t.Fatalf("out=%v", out)
	}
	if gotUA != "cov-test" {
		t.Fatalf("ua=%q", gotUA)
	}
	if gotHdr != "1" || gotIdem != "idem-1" {
		t.Fatalf("hdr=%q idem=%q", gotHdr, gotIdem)
	}
}

func TestChargeDeleteAndQRCodeBase64(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/charge/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/image/qrcode/base64/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("size") != "800" {
			t.Errorf("size=%q", r.URL.Query().Get("size"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"imageBase64":"data:image/png;base64,abc"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	if err := client.Charges.Delete(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
	if err := client.Charges.Delete(context.Background(), ""); err == nil {
		t.Fatal("expected empty id error")
	}
	b64, err := client.Charges.QRCodeBase64(context.Background(), "pay-1", 800)
	if err != nil || b64 == "" {
		t.Fatalf("b64=%q err=%v", b64, err)
	}
}

func TestCustomersListAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/customer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"customers":[{"name":"A","correlationID":"c1"},{"name":"B","correlationID":"c2"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false,"hasPreviousPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	list, err := client.Customers.List(context.Background(), nil)
	if err != nil || len(list.Customers) != 2 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	n := 0
	for c, err := range client.Customers.ListAll(context.Background(), &woovi.CustomerListParams{Limit: 100}) {
		if err != nil {
			t.Fatal(err)
		}
		if c == nil {
			t.Fatal("nil customer")
		}
		n++
	}
	if n != 2 {
		t.Fatalf("n=%d", n)
	}
}

func TestPaymentsListAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/payment", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payments":[{"correlationID":"p1","value":100,"status":"CREATED"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	list, err := client.Payments.List(context.Background(), nil)
	if err != nil || len(list.Payments) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	n := 0
	for p, err := range client.Payments.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if p.CorrelationID != "p1" {
			t.Fatalf("p=%+v", p)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestRefundsListAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/refund", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"refunds":[{"correlationID":"r1","value":50,"status":"COMPLETED"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	list, err := client.Refunds.List(context.Background(), nil)
	if err != nil || len(list.Refunds) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	n := 0
	for r, err := range client.Refunds.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if r.CorrelationID != "r1" {
			t.Fatalf("r=%+v", r)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestPixQRCodesListAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/qrcode-static", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixQrCodes":[{"name":"q1","correlationID":"q1","value":100}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	list, err := client.PixQRCodes.List(context.Background(), nil)
	if err != nil || len(list.PixQRCodes) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	n := 0
	for q, err := range client.PixQRCodes.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if q.CorrelationID != "q1" {
			t.Fatalf("q=%+v", q)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestStatementsAndSubaccountsListAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/statement", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statements":[{"id":"E1","value":10}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	mux.HandleFunc("GET /api/v1/subaccount", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subAccounts":[{"name":"s1","pixKey":"k1"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	sn, sa := 0, 0
	for e, err := range client.Statements.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if e.ID != "E1" {
			t.Fatalf("e=%+v", e)
		}
		sn++
	}
	for page, err := range client.Statements.ListPages(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Statements) != 1 {
			t.Fatalf("page=%+v", page)
		}
	}
	for s, err := range client.Subaccounts.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if s.PixKey != "k1" {
			t.Fatalf("s=%+v", s)
		}
		sa++
	}
	for page, err := range client.Subaccounts.ListPages(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if len(page.SubAccounts) != 1 {
			t.Fatalf("page=%+v", page)
		}
	}
	if sn != 1 || sa != 1 {
		t.Fatalf("sn=%d sa=%d", sn, sa)
	}
}

func TestSubscriptionsListAllAndPaymentBook(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subscriptions":[{"correlationID":"s1","value":100}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	mux.HandleFunc("GET /api/v1/subscriptions/{id}/payment-book", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"https://woovi.com/book/s1"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	n := 0
	for s, err := range client.Subscriptions.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if s.CorrelationID != "s1" {
			t.Fatalf("s=%+v", s)
		}
		n++
	}
	book, err := client.Subscriptions.PaymentBook(context.Background(), "s1")
	if err != nil || book.URL == "" {
		t.Fatalf("book=%+v err=%v", book, err)
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestAccountDeleteAndAnticipationDeactivate(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/account/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/v1/anticipation/beneficiary/{taxID}/deactivate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"beneficiary":{"taxID":{"taxID":"12345678909"},"isActive":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	if err := client.Accounts.Delete(context.Background(), "acc-1"); err != nil {
		t.Fatal(err)
	}
	ben, err := client.Anticipations.DeactivateBeneficiary(context.Background(), "12345678909")
	if err != nil || ben.IsActive {
		t.Fatalf("ben=%+v err=%v", ben, err)
	}
}

func TestFilesUploadGuessContentType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want string
	}{
		{"doc.pdf", "application/pdf"},
		{"a.png", "image/png"},
		{"a.jpg", "image/jpeg"},
		{"a.jpeg", "image/jpeg"},
		{"a.webp", "image/webp"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var gotCT string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil {
					t.Fatal(err)
				}
				mr := multipart.NewReader(r.Body, params["boundary"])
				for {
					part, err := mr.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if part.FormName() == "file" {
						gotCT = part.Header.Get("Content-Type")
					}
					_, _ = io.Copy(io.Discard, part)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"file":{"id":"f1","purpose":"DISPUTE_EVIDENCE","fileName":"x","contentType":"x","size":1}}`))
			}))
			t.Cleanup(srv.Close)
			client := mustClient(t, srv)
			_, err := client.Files.Upload(context.Background(), &woovi.FileUploadParams{
				File: bytes.NewReader([]byte("x")), FileName: tc.name, Purpose: woovi.FilePurposeDisputeEvidence,
			})
			if err != nil {
				t.Fatal(err)
			}
			if gotCT != tc.want {
				t.Fatalf("got %q want %q", gotCT, tc.want)
			}
		})
	}
}

func TestFundsRecoveryNestedLists(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/funds-recovery/{id}/infraction-reports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"infractionReports":[{"bacenInfractionReportId":"ir1","status":"OPENED"}]}`))
	})
	mux.HandleFunc("GET /api/v1/funds-recovery/{id}/refund-solicitations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"bacenRefundId":"rs1","status":"OPENED","refundAmount":1000}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	irs, err := client.FundsRecoveries.ListInfractionReports(context.Background(), "fr-1")
	if err != nil || len(irs) != 1 || irs[0].BacenInfractionReportID != "ir1" {
		t.Fatalf("irs=%+v err=%v", irs, err)
	}
	rss, err := client.FundsRecoveries.ListRefundSolicitations(context.Background(), "fr-1")
	if err != nil || len(rss) != 1 || rss[0].BacenRefundID != "rs1" {
		t.Fatalf("rss=%+v err=%v", rss, err)
	}
}

func TestInvoiceDocumentsAndIntegration(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/invoice", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invoices":[{"correlationID":"INV-1","value":500,"status":"CONFIRMED"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	mux.HandleFunc("GET /api/v1/invoice/{id}/pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF"))
	})
	mux.HandleFunc("GET /api/v1/invoice/{id}/xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte("<xml/>"))
	})
	mux.HandleFunc("GET /api/v1/invoice/integration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"isActive":true}`))
	})
	mux.HandleFunc("POST /api/v1/invoice/integration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"isActive":true,"provider":"nfe"}`))
	})
	mux.HandleFunc("PUT /api/v1/invoice/integration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"isActive":false}`))
	})
	mux.HandleFunc("POST /api/v1/invoice/integration/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	n := 0
	for inv, err := range client.Invoices.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if inv.CorrelationID != "INV-1" {
			t.Fatalf("inv=%+v", inv)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
	pdf, err := client.Invoices.PDF(context.Background(), "INV-1")
	if err != nil || string(pdf.Data) != "%PDF" || pdf.ContentType != "application/pdf" {
		t.Fatalf("pdf=%+v err=%v", pdf, err)
	}
	xmlDoc, err := client.Invoices.XML(context.Background(), "INV-1")
	if err != nil || string(xmlDoc.Data) != "<xml/>" || xmlDoc.ContentType != "application/xml" {
		t.Fatalf("xml=%+v err=%v", xmlDoc, err)
	}
	got, err := client.Invoices.GetIntegration(context.Background())
	if err != nil || got["isActive"] != true {
		t.Fatalf("got=%v err=%v", got, err)
	}
	up, err := client.Invoices.UpsertIntegration(context.Background(), map[string]any{"provider": "nfe"})
	if err != nil || up["provider"] != "nfe" {
		t.Fatalf("up=%v err=%v", up, err)
	}
	put, err := client.Invoices.UpdateIntegration(context.Background(), map[string]any{"isActive": false})
	if err != nil || put["isActive"] != false {
		t.Fatalf("put=%v err=%v", put, err)
	}
	testOut, err := client.Invoices.TestIntegration(context.Background())
	if err != nil || testOut["ok"] != true {
		t.Fatalf("test=%v err=%v", testOut, err)
	}
}

func TestKYCRemainingMethods(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/kyc/representatives/documents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"representative":{"id":"r1","name":"Maria","active":true}}`))
	})
	mux.HandleFunc("GET /api/v1/kyc/documents", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("correlationID") != "m1" {
			t.Errorf("corr=%q", r.URL.Query().Get("correlationID"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"type":"CNH","url":"https://x"}]`))
	})
	mux.HandleFunc("POST /api/v1/kyc/documents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("GET /api/v1/kyc/rfi", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"OPEN","items":[]}`))
	})
	mux.HandleFunc("POST /api/v1/kyc/rfi", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ANSWERED"}`))
	})
	mux.HandleFunc("POST /api/v1/kyc/pix-authentication", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"pa1","status":"PENDING"}`))
	})
	mux.HandleFunc("GET /api/v1/kyc/pix-authentication/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"pa1","status":"COMPLETED"}`))
	})
	mux.HandleFunc("GET /api/v1/kyc/bc-protection", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"enabled":true}`))
	})
	mux.HandleFunc("POST /api/v1/kyc/bc-protection/resend", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resent":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	rep, err := client.KYC.AttachRepresentativeDocuments(context.Background(), &woovi.KYCRepresentativeDocumentsParams{
		CorrelationID:    "m1",
		RepresentativeID: "r1",
		Documents:        []woovi.KYCDocumentInput{{Type: "CNH", FileID: "f1"}},
	})
	if err != nil || rep.Representative.ID != "r1" {
		t.Fatalf("rep=%+v err=%v", rep, err)
	}
	docs, err := client.KYC.ListDocuments(context.Background(), "m1")
	if err != nil || len(docs) == 0 {
		t.Fatalf("docs=%s err=%v", docs, err)
	}
	att, err := client.KYC.AttachDocuments(context.Background(), &woovi.KYCCompanyDocumentsParams{
		CorrelationID: "m1",
		Documents:     []woovi.KYCDocumentInput{{Type: "CNPJ_CARD", FileID: "f2"}},
	})
	if err != nil || len(att) == 0 {
		t.Fatalf("att=%s err=%v", att, err)
	}
	rfi, err := client.KYC.GetRFI(context.Background(), "m1")
	if err != nil || len(rfi) == 0 {
		t.Fatalf("rfi=%s err=%v", rfi, err)
	}
	ans, err := client.KYC.AnswerRFI(context.Background(), &woovi.KYCRFIAnswerParams{
		CorrelationID: "m1",
		Documents:     []woovi.KYCDocumentInput{{Type: "PROOF", FileID: "f3"}},
	})
	if err != nil || len(ans) == 0 {
		t.Fatalf("ans=%s err=%v", ans, err)
	}
	pa, err := client.KYC.CreatePixAuthentication(context.Background(), &woovi.KYCPixAuthenticationParams{CorrelationID: "m1"})
	if err != nil || !bytes.Contains(pa, []byte(`"pa1"`)) {
		t.Fatalf("pa=%s err=%v", pa, err)
	}
	gotPA, err := client.KYC.GetPixAuthentication(context.Background(), "pa1")
	if err != nil || !bytes.Contains(gotPA, []byte(`COMPLETED`)) {
		t.Fatalf("gotPA=%s err=%v", gotPA, err)
	}
	bc, err := client.KYC.GetBCProtection(context.Background(), "m1")
	if err != nil || !bytes.Contains(bc, []byte(`true`)) {
		t.Fatalf("bc=%s err=%v", bc, err)
	}
	resend, err := client.KYC.ResendBCProtection(context.Background(), "m1")
	if err != nil || !bytes.Contains(resend, []byte(`true`)) {
		t.Fatalf("resend=%s err=%v", resend, err)
	}
}

func TestStablecoinRemainingMethods(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/stablecoin/quote", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rate":5.1}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/payout/approve", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"PROCESSING","payoutId":"p1","correlationId":"pay-1"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/payout/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"COMPLETED","payoutId":"p1"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/payout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payouts":[]}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/payout/quote", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("value") != "10000" || r.URL.Query().Get("currency") != "USDT" {
			t.Errorf("q=%s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"outputAmount":10000}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/swap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"swapId":"sw1"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/swap/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"swapId":"sw1","status":"DONE"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/swap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"swaps":[]}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/swap/quote", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rate":1}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/wallets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"wallets":[{"id":"w1"}]}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/subaccount/{id}/balances", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balances":[]}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/subaccount/{id}/wallets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"wallets":[]}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/subaccount/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subAccountId":"sub_1","status":"ACTIVE"}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/subaccount/kyb/usd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"PENDING"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/subaccount/kyb/usd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"APPROVED"}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/subaccount/kyb/usd/document", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uploadUrl":"https://s3"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/limit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"daily":100000}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/limit/document", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uploadUrl":"https://s3/doc"}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/limit/request", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"lr1"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/limit/request/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"lr1","status":"OPEN"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/limit/request", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requests":[]}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/limit/request/{id}/document", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"attached":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	mustRaw := func(raw json.RawMessage, err error) {
		t.Helper()
		if err != nil || len(raw) == 0 {
			t.Fatalf("raw=%s err=%v", raw, err)
		}
	}

	mustRaw(client.Stablecoins.QuoteDeposit(context.Background(), url.Values{"currency": {"USDT"}}))
	payout, err := client.Stablecoins.ApprovePayout(context.Background(), "pay-1")
	if err != nil || payout.PayoutID != "p1" {
		t.Fatalf("payout=%+v err=%v", payout, err)
	}
	got, err := client.Stablecoins.GetPayout(context.Background(), "p1")
	if err != nil || got.Status != "COMPLETED" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	mustRaw(client.Stablecoins.ListPayouts(context.Background(), nil))
	mustRaw(client.Stablecoins.QuotePayout(context.Background(), woovi.Cents(10000), woovi.StablecoinUSDT))
	mustRaw(client.Stablecoins.CreateSwap(context.Background(), woovi.StablecoinSwapCreateParams{"from": "USDT", "to": "USDC"}))
	mustRaw(client.Stablecoins.GetSwap(context.Background(), "sw1"))
	mustRaw(client.Stablecoins.ListSwaps(context.Background(), nil))
	mustRaw(client.Stablecoins.QuoteSwap(context.Background(), url.Values{"from": {"USDT"}}))
	mustRaw(client.Stablecoins.ListWallets(context.Background()))
	sa, err := client.Stablecoins.GetSubaccount(context.Background(), "sub_1")
	if err != nil || sa.SubAccountID != "sub_1" {
		t.Fatalf("sa=%+v err=%v", sa, err)
	}
	mustRaw(client.Stablecoins.GetSubaccountBalances(context.Background(), "sub_1"))
	mustRaw(client.Stablecoins.ListSubaccountWallets(context.Background(), "sub_1"))
	mustRaw(client.Stablecoins.CreateUSDKYB(context.Background(), map[string]any{"subAccountId": "sub_1"}))
	mustRaw(client.Stablecoins.GetUSDKYB(context.Background(), url.Values{"subAccountId": {"sub_1"}}))
	mustRaw(client.Stablecoins.CreateUSDKYBDocumentUpload(context.Background(), map[string]any{"type": "PASSPORT"}))
	mustRaw(client.Stablecoins.GetLimits(context.Background(), nil))
	mustRaw(client.Stablecoins.CreateLimitDocumentUpload(context.Background(), &woovi.StablecoinLimitDocumentUploadParams{
		Type: woovi.StablecoinLimitProofFinancialCapacity, FileName: "a.pdf", MimeType: "application/pdf",
	}))
	mustRaw(client.Stablecoins.CreateLimitRequest(context.Background(), map[string]any{"amount": 100000}))
	mustRaw(client.Stablecoins.GetLimitRequest(context.Background(), "lr1"))
	mustRaw(client.Stablecoins.ListLimitRequests(context.Background(), nil))
	mustRaw(client.Stablecoins.AttachLimitRequestDocument(context.Background(), "lr1", map[string]any{"documentId": "d1"}))
}

func TestLimitsListRequests(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/limits/request", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"limitRequests":[{"id":"lr1","status":"PENDING"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	list, err := client.Limits.ListRequests(context.Background(), nil)
	if err != nil || len(list.LimitRequests) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	n := 0
	for lr, err := range client.Limits.ListAllRequests(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if lr.ID != "lr1" {
			t.Fatalf("lr=%+v", lr)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestPixKeyCheckPathAndTokenLogsAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/pix-keys/{key}/check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixKey":"a@b.com","type":"EMAIL","owner":{"name":"A","taxID":"12345678909"}}`))
	})
	mux.HandleFunc("GET /api/v1/pix-keys/tokens/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"logs":[{"operation":"REMOVE","reason":"DICT_LOOKUP","tokens":1}],"pageInfo":{"skip":0,"limit":50,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	got, err := client.PixKeys.CheckPath(context.Background(), "a@b.com")
	if err != nil || got.PixKey != "a@b.com" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	n := 0
	for log, err := range client.PixKeys.ListAllTokenLogs(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if log.Tokens != 1 {
			t.Fatalf("log=%+v", log)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestAPIErrorRateLimitedUnwrap(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"slow down"}`))
	}))
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	_, err := client.Charges.Get(context.Background(), "x")
	if !errors.Is(err, woovi.ErrRateLimited) {
		t.Fatalf("err=%v", err)
	}
}

func TestWebhooksCRUDAndEvents(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"webhook":{"id":"wh-1","name":"charges","url":"https://example.com/hook","isActive":true,"event":"OPENPIX:CHARGE_COMPLETED"}}`))
	})
	mux.HandleFunc("GET /api/v1/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"webhooks":[{"id":"wh-1","name":"charges","url":"https://example.com/hook","isActive":true,"event":"OPENPIX:CHARGE_COMPLETED"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	mux.HandleFunc("DELETE /api/v1/webhook/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/v1/webhook/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"events":["OPENPIX:CHARGE_CREATED","OPENPIX:CHARGE_COMPLETED"]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	created, err := client.Webhooks.Create(context.Background(), &woovi.WebhookCreateParams{
		Name: "charges", Event: woovi.WebhookEventChargeCompleted, URL: "https://example.com/hook", IsActive: true,
	})
	if err != nil || created.ID != "wh-1" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	list, err := client.Webhooks.List(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	if err := client.Webhooks.Delete(context.Background(), "wh-1"); err != nil {
		t.Fatal(err)
	}
	events, err := client.Webhooks.ListEvents(context.Background())
	if err != nil || len(events) != 2 {
		t.Fatalf("events=%+v err=%v", events, err)
	}
}

func TestTEDListAll(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/ted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"teds":[{"correlationID":"t1","value":1000,"status":"COMPLETED"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	n := 0
	for ted, err := range client.TEDs.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if ted.CorrelationID != "t1" {
			t.Fatalf("ted=%+v", ted)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("n=%d", n)
	}
}

func TestAuthTokenValidationOKVariants(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/validate-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true,"application":{"name":"app"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	got, err := client.Auth.ValidateToken(context.Background())
	if err != nil || !got.OK() {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestRetryAfterHeader(t *testing.T) {
	t.Parallel()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"charge":{"correlationID":"c1","value":1,"status":"ACTIVE"}}`))
	}))
	t.Cleanup(srv.Close)
	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 2, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Charges.Get(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
	if n.Load() < 2 {
		t.Fatalf("attempts=%d", n.Load())
	}
}
