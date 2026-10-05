package woovi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestFilesUpload(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/files" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Fatalf("content-type=%q err=%v", r.Header.Get("Content-Type"), err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		var purpose, correlationID, fileName string
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			switch part.FormName() {
			case "purpose":
				b, _ := io.ReadAll(part)
				purpose = string(b)
			case "correlationID":
				b, _ := io.ReadAll(part)
				correlationID = string(b)
			case "file":
				fileName = part.FileName()
				_, _ = io.Copy(io.Discard, part)
			}
		}
		if purpose != "DISPUTE_EVIDENCE" || correlationID != "ev-1" || fileName != "evidence.png" {
			t.Fatalf("purpose=%q corr=%q file=%q", purpose, correlationID, fileName)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"file":{"id":"file-1","purpose":"DISPUTE_EVIDENCE","fileName":"evidence.png","contentType":"image/png","size":4}}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	f, err := client.Files.Upload(context.Background(), &woovi.FileUploadParams{
		File:          bytes.NewReader([]byte{0x89, 0x50, 0x4e, 0x47}),
		FileName:      "evidence.png",
		ContentType:   "image/png",
		Purpose:       woovi.FilePurposeDisputeEvidence,
		CorrelationID: "ev-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if f.ID != "file-1" || f.Purpose != woovi.FilePurposeDisputeEvidence {
		t.Fatalf("file=%+v", f)
	}
}

func TestDisputesGetListEvidence(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/dispute/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dispute":{"id":"d1","status":"IN_REVIEW","endToEndId":"E1","type":"MED","value":10000}}`))
	})
	mux.HandleFunc("GET /api/v1/dispute", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start") == "" {
			t.Error("missing start")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"disputes":[{"status":"IN_REVIEW","value":100}],"pageInfo":{"skip":0,"limit":10,"hasNextPage":false}}`))
	})
	mux.HandleFunc("POST /api/v1/dispute/{id}/evidence", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		docs := in["documents"].([]any)
		doc := docs[0].(map[string]any)
		if doc["fileId"] != "file-1" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"documents":[{"fileId":"file-1","description":"nf"}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	d, err := client.Disputes.Get(context.Background(), "E1")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != woovi.DisputeStatusInReview || d.Type != woovi.DisputeTypeMED {
		t.Fatalf("dispute=%+v", d)
	}
	list, err := client.Disputes.List(context.Background(), &woovi.DisputeListParams{
		Start: "2020-01-01T00:00:00Z",
		End:   "2020-12-01T17:00:00Z",
	})
	if err != nil || len(list.Disputes) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	ev, err := client.Disputes.AddEvidence(context.Background(), "d1", &woovi.DisputeEvidenceParams{
		Documents: []woovi.DisputeEvidenceDocument{{
			FileID:      "file-1",
			Description: "nf",
		}},
	})
	if err != nil || len(ev.Documents) != 1 {
		t.Fatalf("evidence=%+v err=%v", ev, err)
	}
}

func TestLimitsAndCashbackAndPSP(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/limits/{accountId}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"limits":{"pixDayLimit":4000000,"pixNightLimit":100000},"usage":{"window":"DAY","aggregate":{"pixOut":{"period":"DAY_WINDOW","totalLimit":4000000,"usedValue":1250000,"availableValue":2750000}}}}`))
	})
	mux.HandleFunc("POST /api/v1/limits/request", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"limitRequest":{"id":"lr-1","status":"IN_REVIEW","companyBankAccountId":"acc-1"}}`))
	})
	mux.HandleFunc("GET /api/v1/limits/request/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"limitRequest":{"id":"lr-1","status":"APPROVED","approvedLimits":{"pixDayLimit":4500000}}}`))
	})
	mux.HandleFunc("POST /api/v1/cashback-fidelity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"cashback":{"value":1500}}`))
	})
	mux.HandleFunc("GET /api/v1/cashback-fidelity/balance/{taxID}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balance":1500,"status":"ACTIVE"}`))
	})
	mux.HandleFunc("GET /api/v1/psp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"psps":[{"name":"BCO DO BRASIL S.A.","ispb":"00000000","compe":"001"}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	limits, err := client.Limits.Get(context.Background(), "acc-1")
	if err != nil || limits.Limits.PixDayLimit != 4000000 {
		t.Fatalf("limits=%+v err=%v", limits, err)
	}
	req, err := client.Limits.CreateRequest(context.Background(), &woovi.LimitRequestCreateParams{
		CompanyBankAccountID: "acc-1",
		PixDayLimit:          woovi.Cents(5000000),
		PixNightLimit:        woovi.Cents(200000),
		Documents:            []woovi.LimitRequestDocumentInput{{FileID: "f1"}},
	})
	if err != nil || req.ID != "lr-1" {
		t.Fatalf("req=%+v err=%v", req, err)
	}
	got, err := client.Limits.GetRequest(context.Background(), "lr-1")
	if err != nil || got.Status != woovi.LimitRequestStatusApproved {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	cb, err := client.CashbackFidelities.Create(context.Background(), &woovi.CashbackFidelityCreateParams{
		TaxID: "11111111111",
		Value: woovi.Cents(1500),
	})
	if err != nil || cb.Cashback.Value != 1500 {
		t.Fatalf("cb=%+v err=%v", cb, err)
	}
	bal, err := client.CashbackFidelities.Balance(context.Background(), "11111111111")
	if err != nil || bal.Balance != 1500 {
		t.Fatalf("bal=%+v err=%v", bal, err)
	}
	psps, err := client.PSPs.List(context.Background(), &woovi.PSPListParams{Compe: "001"})
	if err != nil || len(psps) != 1 || psps[0].ISPB != "00000000" {
		t.Fatalf("psps=%+v err=%v", psps, err)
	}
}

func TestApplicationDeleteRotateScopes(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/v1/application", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	})
	mux.HandleFunc("POST /api/v1/application/rotate-secret", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["clientId"] != "client_123" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"application":{"name":"api","clientId":"client_123","clientSecret":"new-secret","appID":"app-new"}}`))
	})
	mux.HandleFunc("GET /api/v1/application/scopes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"scopes":["CHARGE_POST","CHARGE_GET"]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	if err := client.Applications.Delete(context.Background()); err != nil {
		t.Fatal(err)
	}
	app, err := client.Applications.RotateSecret(context.Background(), &woovi.ApplicationRotateParams{
		ClientID: "client_123",
	})
	if err != nil || app.ClientSecret != "new-secret" || app.AppID != "app-new" {
		t.Fatalf("app=%+v err=%v", app, err)
	}
	scopes, err := client.Applications.ListScopes(context.Background())
	if err != nil || len(scopes.Scopes) != 2 {
		t.Fatalf("scopes=%+v err=%v", scopes, err)
	}
}

func TestSubaccountCreditDebitAndPixKeyTokens(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/subaccount/{id}/credit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixKey":"seller@woovi.com","value":100,"success":"ok"}`))
	})
	mux.HandleFunc("POST /api/v1/subaccount/{id}/debit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixKey":"seller@woovi.com","value":50,"success":"ok"}`))
	})
	mux.HandleFunc("GET /api/v1/pix-keys/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tokens":480,"maxTokens":500,"refreshRate":20,"tokensAfterRefresh":500,"nextRefresh":"2026-01-01T12:01:00.000Z"}`))
	})
	mux.HandleFunc("GET /api/v1/pix-keys/tokens/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"logs":[{"operation":"REMOVE","reason":"DICT_LOOKUP","tokens":1}],"pageInfo":{"skip":0,"limit":50,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	credit, err := client.Subaccounts.Credit(context.Background(), "seller@woovi.com", &woovi.SubaccountMovementParams{
		Value:       woovi.Cents(100),
		Description: "Monthly deposit",
	})
	if err != nil || credit.Value != 100 {
		t.Fatalf("credit=%+v err=%v", credit, err)
	}
	debit, err := client.Subaccounts.Debit(context.Background(), "seller@woovi.com", &woovi.SubaccountMovementParams{
		Value: woovi.Cents(50),
	})
	if err != nil || debit.Value != 50 {
		t.Fatalf("debit=%+v err=%v", debit, err)
	}
	tokens, err := client.PixKeys.Tokens(context.Background())
	if err != nil || tokens.Tokens != 480 || tokens.MaxTokens != 500 {
		t.Fatalf("tokens=%+v err=%v", tokens, err)
	}
	logs, err := client.PixKeys.ListTokenLogs(context.Background(), nil)
	if err != nil || len(logs.Logs) != 1 || logs.Logs[0].Operation != woovi.PixKeyTokenLogRemove {
		t.Fatalf("logs=%+v err=%v", logs, err)
	}
}

func TestChargeBRCodeImagePath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openpix/charge/brcode/image/fe7834b4060c488a9b0f89811be5f5cf.png" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.URL.Query().Get("size") != "768" {
			t.Fatalf("size=%q", r.URL.Query().Get("size"))
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47})
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	img, err := client.Charges.BRCodeImage(context.Background(), "fe7834b4060c488a9b0f89811be5f5cf", 768)
	if err != nil {
		t.Fatal(err)
	}
	if img.ContentType != "image/png" || len(img.Data) != 4 {
		t.Fatalf("img=%+v", img)
	}
}
