package woovi_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestNewClientRequiresAppID(t *testing.T) {
	t.Parallel()
	if _, err := woovi.NewClient(""); err == nil {
		t.Fatal("expected error for empty appID")
	}
}

func TestClientSetsAuthorizationWithoutBearer(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"charge":{"correlationID":"c1","value":100,"status":"ACTIVE"}}`))
	}))
	t.Cleanup(srv.Close)

	client, err := woovi.NewClient("test-app-id",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Charges.Create(context.Background(), &woovi.ChargeCreateParams{
		CorrelationID: "c1",
		Value:         woovi.Cents(100),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "test-app-id" {
		t.Fatalf("Authorization=%q want raw appID", gotAuth)
	}
	if strings.HasPrefix(strings.ToLower(gotAuth), "bearer ") {
		t.Fatal("must not send Bearer prefix")
	}
}

func TestAPIErrorUnauthorized(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"appID inválido"}]}`))
	}))
	t.Cleanup(srv.Close)

	client, err := woovi.NewClient("bad",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Charges.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, woovi.ErrUnauthorized) {
		t.Fatalf("errors.Is Unauthorized: got %v", err)
	}
	var apiErr *woovi.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As APIError: got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status=%d", apiErr.StatusCode)
	}
	if apiErr.Message != "appID inválido" {
		t.Fatalf("message=%q", apiErr.Message)
	}
}

func TestChargesCreateAndGet(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/charge", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		if err := json.Unmarshal(body, &in); err != nil {
			t.Errorf("decode: %v", err)
		}
		if in["correlationID"] != "pedido-1" {
			t.Errorf("correlationID=%v", in["correlationID"])
		}
		if in["value"].(float64) != 1990 {
			t.Errorf("value=%v", in["value"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"charge":{
				"correlationID":"pedido-1",
				"value":1990,
				"status":"ACTIVE",
				"brCode":"000201",
				"paymentLinkUrl":"https://woovi.com/pay/x",
				"qrCodeImage":"https://api.woovi.com/img.png"
			},
			"correlationID":"pedido-1",
			"brCode":"000201"
		}`))
	})
	mux.HandleFunc("GET /api/v1/charge/pedido-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"charge":{"correlationID":"pedido-1","value":1990,"status":"ACTIVE"}}`))
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

	created, err := client.Charges.Create(context.Background(), &woovi.ChargeCreateParams{
		CorrelationID: "pedido-1",
		Value:         woovi.Cents(1990),
		Comment:       woovi.String("Pedido #1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Charge.BrCode != "000201" {
		t.Fatalf("brCode=%q", created.Charge.BrCode)
	}

	got, err := client.Charges.Get(context.Background(), "pedido-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.CorrelationID != "pedido-1" || got.Value != 1990 {
		t.Fatalf("got=%+v", got)
	}
}

func TestChargesListAll(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("skip") {
		case "0":
			_, _ = w.Write([]byte(`{"charges":[{"correlationID":"a","value":1},{"correlationID":"b","value":2}],"pageInfo":{"skip":0,"limit":2,"totalCount":3,"hasNextPage":true}}`))
		case "2":
			_, _ = w.Write([]byte(`{"charges":[{"correlationID":"c","value":3}],"pageInfo":{"skip":2,"limit":2,"totalCount":3,"hasNextPage":false}}`))
		default:
			t.Errorf("unexpected skip=%s call=%d", r.URL.Query().Get("skip"), n)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	for c, err := range client.Charges.ListAll(context.Background(), &woovi.ChargeListParams{Limit: 2}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, c.CorrelationID)
	}
	if len(ids) != 3 || ids[0] != "a" || ids[2] != "c" {
		t.Fatalf("ids=%v", ids)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls=%d", calls.Load())
	}
}

func TestRetryOn503ForGET(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"charge":{"correlationID":"x","value":1,"status":"ACTIVE"}}`))
	}))
	t.Cleanup(srv.Close)

	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 2, InitialBackoff: 1}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Charges.Get(context.Background(), "x")
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected retry, calls=%d", calls.Load())
	}
}

func TestNoRetryPOSTWithoutIdempotency(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 3, InitialBackoff: 1}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Charges.Create(context.Background(), &woovi.ChargeCreateParams{
		CorrelationID: "c1",
		Value:         woovi.Cents(100),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 1 {
		t.Fatalf("POST without idempotency must not retry, calls=%d", calls.Load())
	}
}
