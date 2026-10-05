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

func TestSubscriptionsCreateGetList(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["value"].(float64) != 100 {
			t.Errorf("value=%v", in["value"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subscription":{"globalID":"sub-1","value":100,"dayGenerateCharge":5,"status":"ACTIVE","customer":{"name":"Dan","taxID":{"taxID":"31324227036","type":"BR:CPF"}}}}`))
	})
	mux.HandleFunc("GET /api/v1/subscriptions/sub-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subscription":{"globalID":"sub-1","value":100,"dayGenerateCharge":5}}`))
	})
	mux.HandleFunc("GET /api/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subscriptions":[{"globalID":"sub-1","value":100}],"pageInfo":{"skip":0,"limit":100,"totalCount":1,"hasNextPage":false}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	day := 5
	created, err := client.Subscriptions.Create(context.Background(), &woovi.SubscriptionCreateParams{
		Value: woovi.Cents(100),
		Customer: woovi.SubscriptionCustomerInput{
			Name:  "Dan",
			TaxID: "31324227036",
			Email: "email0@example.com",
			Phone: "5511999999999",
		},
		DayGenerateCharge: &day,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.GlobalID != "sub-1" {
		t.Fatalf("globalID=%q", created.GlobalID)
	}

	got, err := client.Subscriptions.Get(context.Background(), "sub-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != 100 {
		t.Fatalf("value=%d", got.Value)
	}

	list, err := client.Subscriptions.List(context.Background(), &woovi.SubscriptionListParams{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Subscriptions) != 1 {
		t.Fatalf("len=%d", len(list.Subscriptions))
	}
}

func TestSubscriptionsDayValidation(t *testing.T) {
	t.Parallel()
	client, err := woovi.NewClient("app", woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}))
	if err != nil {
		t.Fatal(err)
	}
	bad := 30
	_, err = client.Subscriptions.Create(context.Background(), &woovi.SubscriptionCreateParams{
		Value:             woovi.Cents(100),
		Customer:          woovi.SubscriptionCustomerInput{Name: "A", TaxID: "1"},
		DayGenerateCharge: &bad,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
