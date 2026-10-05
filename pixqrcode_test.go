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

func TestPixQRCodesCreateGetDelete(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/qrcode-static", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		if in["identifier"] != "c782e0ac833d4a899e739b60b" {
			t.Errorf("body=%s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixQrCode":{"name":"QRCode Teste","identifier":"c782e0ac833d4a899e739b60b","correlationID":"c1","brCode":"000201","paymentLinkUrl":"https://woovi.com/pay/x"}}`))
	})
	mux.HandleFunc("GET /api/v1/qrcode-static/c1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pixQrCode":{"name":"QRCode Teste","identifier":"c782e0ac833d4a899e739b60b","correlationID":"c1"}}`))
	})
	mux.HandleFunc("DELETE /api/v1/qrcode-static/c1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := mustClient(t, srv)
	value := woovi.Cents(1000)
	created, err := client.PixQRCodes.Create(context.Background(), &woovi.PixQRCodeCreateParams{
		Name:       "QRcode Teste",
		Identifier: "c782e0ac833d4a899e739b60b",
		Value:      &value,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.BrCode == "" {
		t.Fatal("expected brCode")
	}

	got, err := client.PixQRCodes.Get(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if got.CorrelationID != "c1" {
		t.Fatalf("got=%+v", got)
	}

	if err := client.PixQRCodes.Delete(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
}

func TestPixQRCodeIdentifierValidation(t *testing.T) {
	t.Parallel()
	client, err := woovi.NewClient("app", woovi.WithRetry(woovi.RetryConfig{MaxRetries: 0}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.PixQRCodes.Create(context.Background(), &woovi.PixQRCodeCreateParams{
		Name:       "x",
		Identifier: "has spaces",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
