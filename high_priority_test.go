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

func TestTransferCreate(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["fromPixKey"] != "a@woovi.com" || in["value"].(float64) != 5000 {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":5000,"time":"2026-08-24T15:33:27.165Z","correlationID":"repasse-1"}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	tr, err := client.Transfers.Create(context.Background(), &woovi.TransferCreateParams{
		Value:         woovi.Cents(5000),
		FromPixKey:    "a@woovi.com",
		ToPixKey:      "b@woovi.com",
		CorrelationID: "repasse-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tr.Value != 5000 || tr.CorrelationID != "repasse-1" {
		t.Fatalf("transfer=%+v", tr)
	}
}

func TestStatementListAndCompanyBankAccount(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("companyBankAccount") != "acc-other" {
			t.Errorf("companyBankAccount=%q", r.URL.Query().Get("companyBankAccount"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1","time":"2026-08-20T14:03:11.000Z","description":"Pix recebido","balance":129430,"value":1500,"type":"CREDIT"}]`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	page, err := client.Statements.List(context.Background(), &woovi.StatementListParams{
		Limit:              100,
		CompanyBankAccount: "acc-other",
		Start:              "2026-08-01T00:00:00Z",
		End:                "2026-08-24T23:59:59Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Statements) != 1 || page.Statements[0].Type != woovi.StatementEntryCredit {
		t.Fatalf("page=%+v", page)
	}
}

func TestAccountWithdraw(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/account/acc-1/withdraw" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"withdraw":{"account":{"accountId":"acc-1","balance":{"total":122430,"available":122430,"blocked":0}},"transaction":{"endToEndId":"E123","value":7000}}}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	result, err := client.Accounts.Withdraw(context.Background(), "acc-1", &woovi.AccountWithdrawParams{
		Value: woovi.Cents(7000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Withdraw.Transaction == nil || result.Withdraw.Transaction.Value != 7000 {
		t.Fatalf("result=%+v", result)
	}
}

func TestSubscriptionCancelUpdateInstallments(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/subscriptions/sub-1/cancel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subscription":{"globalID":"sub-1","value":100,"status":"INACTIVE"}}`))
	})
	mux.HandleFunc("PUT /api/v1/subscriptions/sub-1/value", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["value"].(float64) != 250 {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subscription":{"globalID":"sub-1","value":250,"status":"ACTIVE"}}`))
	})
	mux.HandleFunc("GET /api/v1/subscriptions/sub-1/installments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"installments":[{"globalID":"inst-1","value":100,"status":"SCHEDULED"}],"pageInfo":{"skip":0,"limit":100,"hasNextPage":false}}`))
	})
	mux.HandleFunc("GET /api/v1/installments/inst-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"globalID":"inst-1","value":100,"status":"SCHEDULED","cobr":{"identifierId":"id-1","value":100}}`))
	})
	mux.HandleFunc("POST /api/v1/installments/inst-1/cobr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"globalID":"inst-1","value":100,"status":"ACTIVE"}`))
	})
	mux.HandleFunc("POST /api/v1/installments/inst-1/cobr/retry", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"globalID":"inst-1","value":120,"status":"ACTIVE"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)

	updated, err := client.Subscriptions.UpdateValue(context.Background(), "sub-1", woovi.Cents(250))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Value != 250 {
		t.Fatalf("value=%d", updated.Value)
	}

	list, err := client.Subscriptions.ListInstallments(context.Background(), "sub-1", 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Installments) != 1 {
		t.Fatalf("list=%+v", list)
	}

	inst, err := client.Installments.Get(context.Background(), "inst-1")
	if err != nil {
		t.Fatal(err)
	}
	if inst.Cobr == nil {
		t.Fatal("expected cobr")
	}

	if _, err := client.Installments.CreateCobr(context.Background(), "inst-1", nil); err != nil {
		t.Fatal(err)
	}
	value := woovi.Cents(120)
	if _, err := client.Installments.RetryCobr(context.Background(), "inst-1", &woovi.InstallmentCobrParams{Value: &value}); err != nil {
		t.Fatal(err)
	}

	canceled, err := client.Subscriptions.Cancel(context.Background(), "sub-1")
	if err != nil {
		t.Fatal(err)
	}
	if canceled.Status != woovi.SubscriptionStatusInactive {
		t.Fatalf("status=%s", canceled.Status)
	}
}

func TestCompanyChargeUpdatePixDefaultWebhookExtras(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/company", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"company":{"id":"c1","name":"Demo","taxID":"123"}}`))
	})
	mux.HandleFunc("PATCH /api/v1/charge/charge-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"charge":{"correlationID":"charge-1","value":100,"comment":"updated","status":"ACTIVE"}}`))
	})
	mux.HandleFunc("PUT /api/v1/pix-keys/{pixKey}/default", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"main@woovi.com","type":"EMAIL"}`))
	})
	mux.HandleFunc("GET /api/v1/webhook/public-keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"public_keys":[{"key_identifier":"abc","is_current":true,"key":"-----BEGIN PUBLIC KEY-----\nMIIB\n-----END PUBLIC KEY-----"}]}`))
	})
	mux.HandleFunc("GET /api/v1/webhook/ips", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ips":["1.2.3.4","5.6.7.8"]}`))
	})
	mux.HandleFunc("GET /api/v1/webhook/wh-1/toggle-activation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"webhook":{"id":"wh-1","name":"n","url":"https://x","isActive":false,"event":"OPENPIX:CHARGE_COMPLETED"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)

	company, err := client.Companies.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if company.Name != "Demo" {
		t.Fatalf("company=%+v", company)
	}

	charge, err := client.Charges.Update(context.Background(), "charge-1", &woovi.ChargeUpdateParams{
		Comment: woovi.String("updated"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if charge.Comment != "updated" {
		t.Fatalf("charge=%+v", charge)
	}

	key, err := client.PixKeys.SetDefault(context.Background(), "main@woovi.com")
	if err != nil {
		t.Fatal(err)
	}
	if key.Key != "main@woovi.com" {
		t.Fatalf("key=%+v", key)
	}

	ips, err := client.Webhooks.ListIPs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ips) != 2 {
		t.Fatalf("ips=%v", ips)
	}

	keys, err := client.Webhooks.ListPublicKeys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || !keys[0].IsCurrent {
		t.Fatalf("keys=%+v", keys)
	}

	hook, err := client.Webhooks.ToggleActivation(context.Background(), "wh-1")
	if err != nil {
		t.Fatal(err)
	}
	if hook.IsActive {
		t.Fatal("expected inactive")
	}
}

func TestWithCompanyBankAccountQuery(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("companyBankAccount") != "acc-9" {
			t.Errorf("q=%s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"transaction":{"value":1,"endToEndId":"E1"}}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	_, err := client.Transactions.Get(context.Background(), "E1", woovi.WithCompanyBankAccount("acc-9"))
	if err != nil {
		t.Fatal(err)
	}
}
