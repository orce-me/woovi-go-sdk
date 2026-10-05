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

func TestPaymentsCreateGetApprove(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/payment", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["correlationID"] != "pay-1" || in["destinationAlias"] != "38763885700" {
			t.Errorf("unexpected body %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payment":{"value":100,"status":"CREATED","destinationAlias":"38763885700","destinationAliasType":"CPF","correlationID":"pay-1"}}`))
	})
	mux.HandleFunc("GET /api/v1/payment/pay-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payment":{"value":100,"status":"CREATED","correlationID":"pay-1"}}`))
	})
	mux.HandleFunc("POST /api/v1/payment/approve", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["correlationID"] != "pay-1" {
			t.Errorf("approve body %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payment":{"value":100,"status":"APPROVED","correlationID":"pay-1"},"transaction":{"value":100,"endToEndId":"E123","time":"2023-03-20T13:14:17.000Z"},"destination":{"name":"Dan","pixKey":"38763885700"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)

	created, err := client.Payments.Create(context.Background(), &woovi.PaymentCreateParams{
		Value:                woovi.Cents(100),
		DestinationAlias:     "38763885700",
		DestinationAliasType: woovi.PixKeyTypeCPF,
		CorrelationID:        "pay-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != woovi.PaymentStatusCreated {
		t.Fatalf("status=%s", created.Status)
	}

	got, err := client.Payments.Get(context.Background(), "pay-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.CorrelationID != "pay-1" {
		t.Fatalf("got=%+v", got)
	}

	approved, err := client.Payments.Approve(context.Background(), "pay-1")
	if err != nil {
		t.Fatal(err)
	}
	if approved.Payment.Status != woovi.PaymentStatusApproved {
		t.Fatalf("status=%s", approved.Payment.Status)
	}
	if approved.Transaction == nil || approved.Transaction.EndToEndID != "E123" {
		t.Fatalf("transaction=%+v", approved.Transaction)
	}
}

func TestPaymentsCreateByQRCode(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["qrCode"] == "" {
			t.Errorf("missing qrCode: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payment":{"value":100,"status":"CREATED","correlationID":"qr-1","qrCode":"000201"}}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	p, err := client.Payments.Create(context.Background(), &woovi.PaymentCreateParams{
		QRCode:        "000201",
		CorrelationID: "qr-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.QRCode != "000201" {
		t.Fatalf("qr=%q", p.QRCode)
	}
}

func mustClient(t *testing.T, srv *httptest.Server) *woovi.Client {
	t.Helper()
	client, err := woovi.NewClient("app",
		woovi.WithBaseURL(srv.URL),
		woovi.WithHTTPClient(srv.Client()),
		woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
