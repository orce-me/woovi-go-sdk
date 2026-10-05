package webhookverify_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
	"github.com/orce-me/woovi-go-sdk/webhookverify"
)

func TestVerifyHMAC(t *testing.T) {
	t.Parallel()
	body := []byte(`{"data_criacao":"2021-08-10T20:32:14.429Z","evento":"teste_webhook","event":"OPENPIX:CHARGE_COMPLETED"}`)
	secret := "hmac-secret-key"

	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if err := webhookverify.VerifyHMAC(body, sig, secret); err != nil {
		t.Fatal(err)
	}
	if err := webhookverify.VerifyHMAC(body, "bad", secret); err == nil {
		t.Fatal("expected invalid signature")
	}
}

func TestParseEventChargeCompleted(t *testing.T) {
	t.Parallel()
	body := []byte(`{"event":"OPENPIX:CHARGE_COMPLETED","charge":{"correlationID":"abc","value":1000,"status":"COMPLETED"}}`)
	ev, err := webhookverify.ParseEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	cc, ok := ev.(*webhookverify.ChargeCompleted)
	if !ok {
		t.Fatalf("type=%T", ev)
	}
	if cc.Event != woovi.WebhookEventChargeCompleted {
		t.Fatalf("event=%s", cc.Event)
	}
	if cc.Charge == nil || cc.Charge.CorrelationID != "abc" {
		t.Fatalf("charge=%+v", cc.Charge)
	}
}

func TestHandlerHMAC(t *testing.T) {
	t.Parallel()
	body := []byte(`{"event":"OPENPIX:CHARGE_CREATED","charge":{"correlationID":"x","value":1,"status":"ACTIVE"}}`)
	secret := "s3cret"
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	requireRSA := false
	h := webhookverify.Handler(webhookverify.Options{
		HMACSecret: secret,
		RequireRSA: &requireRSA,
	}, func(w http.ResponseWriter, r *http.Request, event webhookverify.Event) {
		if _, ok := event.(*webhookverify.ChargeCreated); !ok {
			t.Errorf("unexpected event %T", event)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(webhookverify.HeaderHMACSignature, sig)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
