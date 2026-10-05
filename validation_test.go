package woovi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/orce-me/woovi-go-sdk"
)

func TestPageInfoHasMore(t *testing.T) {
	t.Parallel()
	if !(woovi.PageInfo{HasNextPage: true}).HasMore() {
		t.Fatal("HasNextPage")
	}
	if !(woovi.PageInfo{Skip: 0, Limit: 10, TotalCount: 25}).HasMore() {
		t.Fatal("totalCount")
	}
	if (woovi.PageInfo{Skip: 20, Limit: 10, TotalCount: 25}).HasMore() {
		t.Fatal("last page")
	}
	if (woovi.PageInfo{}).HasMore() {
		t.Fatal("empty")
	}
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	ctx := context.Background()
	mustErr := func(err error) {
		t.Helper()
		if err == nil {
			t.Fatal("expected error")
		}
	}

	mustErr(client.Accounts.Delete(ctx, ""))
	_, err := client.Accounts.Withdraw(ctx, "", &woovi.AccountWithdrawParams{Value: 1})
	mustErr(err)
	_, err = client.Accounts.Withdraw(ctx, "a", &woovi.AccountWithdrawParams{})
	mustErr(err)

	mustErr(client.Charges.Delete(ctx, ""))
	_, err = client.Charges.Update(ctx, "", &woovi.ChargeUpdateParams{})
	mustErr(err)
	_, err = client.Charges.Update(ctx, "c", nil)
	mustErr(err)
	_, err = client.Charges.BRCodeImage(ctx, "", 0)
	mustErr(err)
	_, err = client.Charges.BRCodeImage(ctx, "c", 10)
	mustErr(err)
	_, err = client.Charges.QRCodeBase64(ctx, "", 0)
	mustErr(err)
	_, err = client.Charges.QRCodeBase64(ctx, "c", 10)
	mustErr(err)

	_, err = client.Customers.Create(ctx, nil)
	mustErr(err)
	_, err = client.Customers.Get(ctx, "")
	mustErr(err)
	_, err = client.Customers.Update(ctx, "", &woovi.CustomerUpdateParams{})
	mustErr(err)

	_, err = client.Payments.Create(ctx, nil)
	mustErr(err)
	_, err = client.Payments.Get(ctx, "")
	mustErr(err)
	_, err = client.Payments.Approve(ctx, "")
	mustErr(err)

	_, err = client.Refunds.Create(ctx, nil)
	mustErr(err)
	_, err = client.Refunds.Get(ctx, "")
	mustErr(err)

	_, err = client.Subscriptions.Create(ctx, nil)
	mustErr(err)
	_, err = client.Subscriptions.Get(ctx, "")
	mustErr(err)
	_, err = client.Subscriptions.Cancel(ctx, "")
	mustErr(err)
	_, err = client.Subscriptions.PaymentBook(ctx, "")
	mustErr(err)

	_, err = client.TEDs.Create(ctx, nil)
	mustErr(err)
	_, err = client.TEDs.Create(ctx, &woovi.TEDCreateParams{})
	mustErr(err)
	_, err = client.TEDs.Create(ctx, &woovi.TEDCreateParams{CorrelationID: "x", Value: 0, AccountID: "a"})
	mustErr(err)
	_, err = client.TEDs.Create(ctx, &woovi.TEDCreateParams{CorrelationID: "x", Value: 1})
	mustErr(err)
	_, err = client.TEDs.Create(ctx, &woovi.TEDCreateParams{CorrelationID: "x", Value: 1, AccountID: "a"})
	mustErr(err)
	_, err = client.TEDs.Get(ctx, "")
	mustErr(err)
	_, err = client.TEDs.Refund(ctx, "")
	mustErr(err)

	_, err = client.PixKeys.Create(ctx, nil)
	mustErr(err)
	_, err = client.PixKeys.Create(ctx, &woovi.PixKeyCreateParams{})
	mustErr(err)
	_, err = client.PixKeys.Check(ctx, "")
	mustErr(err)
	_, err = client.PixKeys.CheckPath(ctx, "")
	mustErr(err)
	mustErr(client.PixKeys.Delete(ctx, ""))
	_, err = client.PixKeys.SetDefault(ctx, "")
	mustErr(err)

	mustErr(client.Subaccounts.Transfer(ctx, nil))
	mustErr(client.Subaccounts.Transfer(ctx, &woovi.SubaccountTransferParams{Value: 0}))
	mustErr(client.Subaccounts.Transfer(ctx, &woovi.SubaccountTransferParams{Value: 1}))
	_, err = client.Subaccounts.Credit(ctx, "", &woovi.SubaccountMovementParams{Value: 1})
	mustErr(err)
	_, err = client.Subaccounts.Debit(ctx, "k", nil)
	mustErr(err)

	_, err = client.Applications.Create(ctx, nil)
	mustErr(err)
	_, err = client.Applications.RotateSecret(ctx, nil)
	mustErr(err)
	_, err = client.Applications.RotateSecret(ctx, &woovi.ApplicationRotateParams{})
	mustErr(err)

	_, err = client.FundsRecoveries.Create(ctx, nil)
	mustErr(err)
	_, err = client.FundsRecoveries.Create(ctx, &woovi.FundsRecoveryCreateParams{})
	mustErr(err)
	_, err = client.FundsRecoveries.Get(ctx, "")
	mustErr(err)
	_, err = client.FundsRecoveries.Cancel(ctx, "")
	mustErr(err)
	_, err = client.FundsRecoveries.ListDisputes(ctx, "")
	mustErr(err)

	_, err = client.Invoices.Create(ctx, nil)
	mustErr(err)
	mustErr(client.Invoices.Cancel(ctx, ""))
	_, err = client.Invoices.PDF(ctx, "")
	mustErr(err)
	_, err = client.Invoices.UpsertIntegration(ctx, nil)
	mustErr(err)
	_, err = client.Invoices.UpdateIntegration(ctx, nil)
	mustErr(err)
	_, err = client.Invoices.PatchIntegration(ctx, nil)
	mustErr(err)
	_, err = client.Invoices.UploadCertificate(ctx, nil)
	mustErr(err)

	_, err = client.KYC.CreateOnboarding(ctx, nil)
	mustErr(err)
	_, err = client.KYC.SubmitOnboarding(ctx, nil)
	mustErr(err)
	_, err = client.KYC.ListRepresentatives(ctx, "")
	mustErr(err)
	_, err = client.KYC.CreateRepresentative(ctx, nil)
	mustErr(err)
	_, err = client.KYC.AttachRepresentativeDocuments(ctx, nil)
	mustErr(err)
	_, err = client.KYC.ListDocuments(ctx, "")
	mustErr(err)
	_, err = client.KYC.AttachDocuments(ctx, nil)
	mustErr(err)
	_, err = client.KYC.GetRFI(ctx, "")
	mustErr(err)
	_, err = client.KYC.AnswerRFI(ctx, nil)
	mustErr(err)
	_, err = client.KYC.CreatePixAuthentication(ctx, nil)
	mustErr(err)
	_, err = client.KYC.GetPixAuthentication(ctx, "")
	mustErr(err)
	_, err = client.KYC.GetBCProtection(ctx, "")
	mustErr(err)
	_, err = client.KYC.ResendBCProtection(ctx, "")
	mustErr(err)
	_, err = client.KYC.CreateValidation(ctx, nil)
	mustErr(err)
	_, err = client.KYC.GetValidation(ctx, "")
	mustErr(err)

	_, err = client.Stablecoins.CreateDeposit(ctx, nil)
	mustErr(err)
	_, err = client.Stablecoins.ApproveDeposit(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.CreatePayout(ctx, nil)
	mustErr(err)
	_, err = client.Stablecoins.ApprovePayout(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.GetPayout(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.QuotePayout(ctx, 0, woovi.StablecoinUSDT)
	mustErr(err)
	_, err = client.Stablecoins.CreateSwap(ctx, nil)
	mustErr(err)
	_, err = client.Stablecoins.GetSwap(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.GetSubaccount(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.GetSubaccountBalances(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.ListSubaccountWallets(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.CreateUSDKYBDocumentUpload(ctx, nil)
	mustErr(err)
	_, err = client.Stablecoins.CreateLimitDocumentUpload(ctx, nil)
	mustErr(err)
	_, err = client.Stablecoins.CreateLimitRequest(ctx, nil)
	mustErr(err)
	_, err = client.Stablecoins.GetLimitRequest(ctx, "")
	mustErr(err)
	_, err = client.Stablecoins.AttachLimitRequestDocument(ctx, "", nil)
	mustErr(err)

	_, err = client.Anticipations.Approve(ctx, "")
	mustErr(err)
	_, err = client.Anticipations.Reject(ctx, "", "x")
	mustErr(err)
	_, err = client.Anticipations.CreateBeneficiary(ctx, nil)
	mustErr(err)
	_, err = client.Anticipations.ActivateBeneficiary(ctx, "")
	mustErr(err)
	_, err = client.Anticipations.DeactivateBeneficiary(ctx, "")
	mustErr(err)
	_, err = client.Anticipations.SyncBalances(ctx, nil)
	mustErr(err)

	_, err = client.Webhooks.Create(ctx, nil)
	mustErr(err)
	mustErr(client.Webhooks.Delete(ctx, ""))

	_, err = client.Partners.CreateCompany(ctx, nil)
	mustErr(err)
	_, err = client.Partners.CreateCompany(ctx, &woovi.PartnerCreateParams{})
	mustErr(err)
	_, err = client.Partners.CreateCompany(ctx, &woovi.PartnerCreateParams{
		PreRegistration: woovi.PartnerPreRegistration{Name: "A", TaxID: woovi.PartnerTaxID{TaxID: "1"}},
	})
	mustErr(err)

	_, err = client.Giftbacks.Balance(ctx, "")
	mustErr(err)
	_, err = client.Files.Upload(ctx, nil)
	mustErr(err)
	_, err = client.Files.Upload(ctx, &woovi.FileUploadParams{Purpose: woovi.FilePurposeDisputeEvidence})
	mustErr(err)
}

func TestApplicationListScopesGroups(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/application/scopes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"scopeGroups":[{"scopes":["A","B"]}],"groups":[{"scopes":["C"]}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	got, err := client.Applications.ListScopes(context.Background())
	if err != nil || len(got.Scopes) != 3 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestStablecoinListSubaccountsArray(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/stablecoin/subaccount", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"subAccountId":"s1","status":"ACTIVE"}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	list, err := client.Stablecoins.ListSubaccounts(context.Background())
	if err != nil || len(list) != 1 || list[0].SubAccountID != "s1" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
}

func TestFundsRecoveryDecodeShapes(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/funds-recovery/{id}/disputes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"d1","type":"MED","status":"OPENED","value":1}`))
	})
	mux.HandleFunc("GET /api/v1/funds-recovery/{id}/infraction-reports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	disputes, err := client.FundsRecoveries.ListDisputes(context.Background(), "fr-1")
	if err != nil || len(disputes) != 1 || disputes[0].ID != "d1" {
		t.Fatalf("disputes=%+v err=%v", disputes, err)
	}
	irs, err := client.FundsRecoveries.ListInfractionReports(context.Background(), "fr-1")
	if err != nil || len(irs) != 0 {
		t.Fatalf("irs=%+v err=%v", irs, err)
	}
}

func TestCustomersListAllMultiPage(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/customer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		n := calls.Add(1)
		if n == 1 {
			_, _ = w.Write([]byte(`{"customers":[{"name":"A","correlationID":"c1"}],"pageInfo":{"skip":0,"limit":1,"hasNextPage":true}}`))
			return
		}
		_, _ = w.Write([]byte(`{"customers":[{"name":"B","correlationID":"c2"}],"pageInfo":{"skip":1,"limit":1,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)

	n := 0
	for c, err := range client.Customers.ListAll(context.Background(), &woovi.CustomerListParams{Limit: 1}) {
		if err != nil {
			t.Fatal(err)
		}
		if c == nil {
			t.Fatal("nil")
		}
		n++
	}
	if n != 2 || calls.Load() != 2 {
		t.Fatalf("n=%d calls=%d", n, calls.Load())
	}
}

func TestRetryAfterHTTPDate(t *testing.T) {
	t.Parallel()
	var n atomic.Int32
	when := time.Now().UTC().Add(50 * time.Millisecond).Format(http.TimeFormat)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) == 1 {
			w.Header().Set("Retry-After", when)
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"charge":{"correlationID":"c1","value":1,"status":"ACTIVE"}}`))
	}))
	t.Cleanup(srv.Close)
	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 2, InitialBackoff: time.Millisecond, MaxBackoff: time.Second}),
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

func TestPixKeyCreateFlatResponse(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/pix-keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"a@b.com","type":"EMAIL"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	key, err := client.PixKeys.Create(context.Background(), &woovi.PixKeyCreateParams{
		PixKey: "a@b.com", Type: woovi.PixKeyTypeEmail,
	})
	if err != nil || key.Key != "a@b.com" {
		t.Fatalf("key=%+v err=%v", key, err)
	}
}

func TestKYCListRepresentativesArray(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/kyc/representatives", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"r1","name":"Maria","active":true}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	reps, err := client.KYC.ListRepresentatives(context.Background(), "m1")
	if err != nil || len(reps) != 1 || reps[0].ID != "r1" {
		t.Fatalf("reps=%+v err=%v", reps, err)
	}
}

func TestDoBytesNilContext(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	//nolint:staticcheck // intentional nil context for coverage
	_, err := client.Charges.BRCodeImage(nil, "id", 0)
	if err == nil {
		t.Fatal("expected nil context error")
	}
}

func TestInvoiceListAllStopsOnFullLastPage(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/invoice", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		// Full page with hasNextPage=false must not loop forever.
		_, _ = w.Write([]byte(`{"invoices":[{"correlationID":"I1","value":1}],"pageInfo":{"skip":0,"limit":1,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	n := 0
	for inv, err := range client.Invoices.ListAll(context.Background(), &woovi.InvoiceListParams{Limit: 1}) {
		if err != nil {
			t.Fatal(err)
		}
		_ = inv
		n++
	}
	if n != 1 || calls.Load() != 1 {
		t.Fatalf("n=%d calls=%d", n, calls.Load())
	}
}

func TestAuthOKFromScopesOnly(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/validate-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"scopes":["CHARGE_GET"]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := mustClient(t, srv)
	got, err := client.Auth.ValidateToken(context.Background())
	if err != nil || !got.OK() {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	_ = json.RawMessage(got.Raw)
}
