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

func TestCustomersCreateGetUpdate(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/customer", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["name"] != "Ana" || in["taxID"] != "67200000051" {
			t.Errorf("unexpected body %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"customer":{"name":"Ana","correlationID":"cust-1","taxID":{"taxID":"67200000051","type":"BR:CPF"}}}`))
	})
	mux.HandleFunc("GET /api/v1/customer/67200000051", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"customer":{"name":"Ana","correlationID":"cust-1","taxID":{"taxID":"67200000051","type":"BR:CPF"}}}`))
	})
	mux.HandleFunc("PATCH /api/v1/customer/cust-1", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["email"] != "ana@example.com" {
			t.Errorf("unexpected patch %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"customer":{"name":"Ana","email":"ana@example.com","correlationID":"cust-1","taxID":{"taxID":"67200000051","type":"BR:CPF"}}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatal(err)
	}

	created, err := client.Customers.Create(context.Background(), &woovi.CustomerCreateParams{
		Name:  "Ana",
		TaxID: "67200000051",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.CorrelationID != "cust-1" {
		t.Fatalf("correlationID=%q", created.CorrelationID)
	}

	got, err := client.Customers.Get(context.Background(), "67200000051")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Ana" {
		t.Fatalf("name=%q", got.Name)
	}

	updated, err := client.Customers.Update(context.Background(), "cust-1", &woovi.CustomerUpdateParams{
		Email: woovi.String("ana@example.com"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Email != "ana@example.com" {
		t.Fatalf("email=%q", updated.Email)
	}
}

func TestCustomersCreateValidation(t *testing.T) {
	t.Parallel()
	client, err := woovi.NewClient("app", woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Customers.Create(context.Background(), &woovi.CustomerCreateParams{Name: "X"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestOnResponseHook(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req-1")
		_, _ = w.Write([]byte(`{"customer":{"name":"A","correlationID":"c1","taxID":{"taxID":"1","type":"BR:CPF"}}}`))
	}))
	t.Cleanup(srv.Close)

	var sawStatus int
	var sawAttempt int
	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}),
		woovi.WithOnRequest(func(info woovi.RequestInfo) {
			sawAttempt = info.Attempt
		}),
		woovi.WithOnResponse(func(req woovi.RequestInfo, resp woovi.ResponseInfo) {
			sawStatus = resp.StatusCode
			if resp.RequestID != "req-1" {
				t.Errorf("requestID=%q", resp.RequestID)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Customers.Get(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if sawStatus != http.StatusOK {
		t.Fatalf("status=%d", sawStatus)
	}
	if sawAttempt != 0 {
		t.Fatalf("attempt=%d", sawAttempt)
	}
}
