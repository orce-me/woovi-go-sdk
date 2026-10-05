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

func TestBoletoValidateAndPay(t *testing.T) {
	t.Parallel()
	barcode := "34195148200000003001095517077320772982609000"
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/boleto/validate", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["barcode"] != barcode {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"boleto":{"barcode":"` + barcode + `","expiresDate":"2026-06-27T02:59:59.999Z","totalValue":300,"issuingEntity":{"code":"341","name":"ITAU"},"finalBeneficiary":{"name":"WOOVI","taxID":"44720743000101"}}}`))
	})
	mux.HandleFunc("POST /api/v1/payment", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["type"] != "BOLETO" || in["boletoBarcode"] != barcode {
			t.Errorf("payment body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"payment":{"value":300,"status":"CREATED","type":"BOLETO","boletoBarcode":"` + barcode + `","correlationID":"boleto-1"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	boleto, err := client.Boletos.Validate(context.Background(), barcode)
	if err != nil {
		t.Fatal(err)
	}
	if boleto.TotalValue != 300 || boleto.FinalBeneficiary.Name != "WOOVI" {
		t.Fatalf("boleto=%+v", boleto)
	}

	payment, err := client.Payments.Create(context.Background(), &woovi.PaymentCreateParams{
		Type:          woovi.PaymentTypeBoleto,
		BoletoBarcode: barcode,
		CorrelationID: "boleto-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if payment.Status != woovi.PaymentStatusCreated || payment.Value != 300 {
		t.Fatalf("payment=%+v", payment)
	}
}

func TestBoletoBarcodeValidation(t *testing.T) {
	t.Parallel()
	client, err := woovi.NewClient("app", woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Boletos.Validate(context.Background(), "123")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
