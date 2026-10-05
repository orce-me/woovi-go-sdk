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

func TestAnticipationFlow(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/anticipation", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "PENDING" {
			t.Errorf("status=%q", r.URL.Query().Get("status"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"anticipations":[{"id":"a1","status":"PENDING","requestedAmount":100000}],"count":1}`))
	})
	mux.HandleFunc("POST /api/v1/anticipation/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"anticipation":{"id":"a1","status":"PROCESSING"}}`))
	})
	mux.HandleFunc("POST /api/v1/anticipation/{id}/reject", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["reason"] != "acima do limite" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"anticipation":{"id":"a1","status":"CANCELED"}}`))
	})
	mux.HandleFunc("POST /api/v1/anticipation/beneficiary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"beneficiary":{"name":"Joao","taxID":{"taxID":"12345678909","type":"BR:CPF"},"isActive":true},"correlationID":"erp-1"}`))
	})
	mux.HandleFunc("POST /api/v1/anticipation/beneficiary/{taxID}/activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"beneficiary":{"taxID":{"taxID":"12345678909"},"isActive":true}}`))
	})
	mux.HandleFunc("POST /api/v1/anticipation/balance/batch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"processed":1,"succeeded":1,"failed":0,"results":[{"taxID":"12345678909","ok":true}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	list, err := client.Anticipations.List(context.Background(), &woovi.AnticipationListParams{Status: woovi.AnticipationStatusPending})
	if err != nil || len(list.Anticipations) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	approved, err := client.Anticipations.Approve(context.Background(), "a1")
	if err != nil || approved.Status != woovi.AnticipationStatusProcessing {
		t.Fatalf("approved=%+v err=%v", approved, err)
	}
	rejected, err := client.Anticipations.Reject(context.Background(), "a1", "acima do limite")
	if err != nil || rejected.Status != woovi.AnticipationStatusCanceled {
		t.Fatalf("rejected=%+v err=%v", rejected, err)
	}
	ben, err := client.Anticipations.CreateBeneficiary(context.Background(), &woovi.AnticipationBeneficiaryCreateParams{
		Name:  "Joao",
		TaxID: "12345678909",
	})
	if err != nil || !ben.Beneficiary.IsActive {
		t.Fatalf("ben=%+v err=%v", ben, err)
	}
	act, err := client.Anticipations.ActivateBeneficiary(context.Background(), "12345678909")
	if err != nil || !act.IsActive {
		t.Fatalf("act=%+v err=%v", act, err)
	}
	batch, err := client.Anticipations.SyncBalances(context.Background(), &woovi.AnticipationBalanceBatchParams{
		Items: []woovi.AnticipationBalanceItem{{
			TaxID: "12345678909", AvailableAmount: 500000, MaxAdvanceableAmount: 350000,
		}},
	})
	if err != nil || batch.Succeeded != 1 {
		t.Fatalf("batch=%+v err=%v", batch, err)
	}
}

func TestKYCAndStablecoin(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/kyc/onboarding", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"linkOnboarding":"https://kyc.woovi.com/x","accountRegister":{"correlationID":"m1","status":"PENDING"}}`))
	})
	mux.HandleFunc("POST /api/v1/kyc/onboarding/submit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"correlationID":"m1","status":"IN_REVIEW"}`))
	})
	mux.HandleFunc("GET /api/v1/kyc/representatives", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"representatives":[{"id":"r1","name":"Maria","type":"ADMIN","active":true}]}`))
	})
	mux.HandleFunc("POST /api/v1/kyc/representatives", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"representative":{"id":"r2","name":"Maria","active":true}}`))
	})
	mux.HandleFunc("POST /api/v1/kyc-validation/taxid", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"correlationID":"v1","status":"PROCESSING","taxID":"12345678909"}`))
	})
	mux.HandleFunc("GET /api/v1/kyc-validation/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"correlationID":"v1","status":"COMPLETED"}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/deposit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"PENDING","depositId":"d1","correlationId":"dep-1","quote":{"inputAmount":10000,"outputCurrency":"USDT"}}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/deposit/approve", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"PROCESSING","depositId":"d1","correlationId":"dep-1"}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/payout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"PENDING","payoutId":"p1","correlationId":"pay-1"}`))
	})
	mux.HandleFunc("GET /api/v1/stablecoin/subaccount", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","subAccounts":[{"id":"sa1","subAccountId":"sub_1"}]}`))
	})
	mux.HandleFunc("POST /api/v1/stablecoin/subaccount", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"subAccountId":"sub_1","status":"IN_REVIEW"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	onb, err := client.KYC.CreateOnboarding(context.Background(), &woovi.KYCOnboardingParams{
		TaxID:         "12345678000199",
		CorrelationID: "m1",
	})
	if err != nil || onb.LinkOnboarding == "" {
		t.Fatalf("onb=%+v err=%v", onb, err)
	}
	sub, err := client.KYC.SubmitOnboarding(context.Background(), &woovi.KYCSubmitParams{CorrelationID: "m1"})
	if err != nil || sub.Status != "IN_REVIEW" {
		t.Fatalf("sub=%+v err=%v", sub, err)
	}
	reps, err := client.KYC.ListRepresentatives(context.Background(), "m1")
	if err != nil || len(reps) != 1 {
		t.Fatalf("reps=%+v err=%v", reps, err)
	}
	rep, err := client.KYC.CreateRepresentative(context.Background(), &woovi.KYCRepresentativeCreateParams{
		CorrelationID: "m1", Name: "Maria", TaxID: "52998224725", Type: "ADMIN",
	})
	if err != nil || rep.Representative.ID != "r2" {
		t.Fatalf("rep=%+v err=%v", rep, err)
	}
	val, err := client.KYC.CreateValidation(context.Background(), &woovi.KYCValidationCreateParams{
		TaxID: "12345678909", CorrelationID: "v1",
	})
	if err != nil || val.Status != "PROCESSING" {
		t.Fatalf("val=%+v err=%v", val, err)
	}
	gotVal, err := client.KYC.GetValidation(context.Background(), "v1")
	if err != nil || gotVal.Status != "COMPLETED" {
		t.Fatalf("gotVal=%+v err=%v", gotVal, err)
	}

	gross := woovi.Cents(10000)
	dep, err := client.Stablecoins.CreateDeposit(context.Background(), &woovi.StablecoinDepositCreateParams{
		GrossAmount:   &gross,
		Currency:      woovi.StablecoinUSDT,
		CorrelationID: "dep-1",
	})
	if err != nil || dep.DepositID != "d1" {
		t.Fatalf("dep=%+v err=%v", dep, err)
	}
	approved, err := client.Stablecoins.ApproveDeposit(context.Background(), "dep-1")
	if err != nil || approved.Status != "PROCESSING" {
		t.Fatalf("approved=%+v err=%v", approved, err)
	}
	payout, err := client.Stablecoins.CreatePayout(context.Background(), &woovi.StablecoinPayoutCreateParams{
		Value: woovi.Cents(10000), Currency: woovi.StablecoinUSDC, PixKey: "key", CorrelationID: "pay-1",
	})
	if err != nil || payout.PayoutID != "p1" {
		t.Fatalf("payout=%+v err=%v", payout, err)
	}
	sas, err := client.Stablecoins.ListSubaccounts(context.Background())
	if err != nil || len(sas) != 1 {
		t.Fatalf("sas=%+v err=%v", sas, err)
	}
	created, err := client.Stablecoins.CreateSubaccount(context.Background(), nil)
	if err != nil || created.Status != "IN_REVIEW" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
}
