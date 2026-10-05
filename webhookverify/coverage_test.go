package webhookverify_test

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orce-me/woovi-go-sdk"
	"github.com/orce-me/woovi-go-sdk/webhookverify"
)

func TestParseEventVariants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		body string
		typ  any
	}{
		{"expired", `{"event":"OPENPIX:CHARGE_EXPIRED","charge":{"correlationID":"e1"}}`, (*webhookverify.ChargeExpired)(nil)},
		{"completed_not_same", `{"event":"OPENPIX:CHARGE_COMPLETED_NOT_SAME_CUSTOMER_PAYER","charge":{"correlationID":"c1"}}`, (*webhookverify.ChargeCompleted)(nil)},
		{"unknown", `{"event":"OPENPIX:MOVEMENT_CONFIRMED"}`, (*webhookverify.UnknownEvent)(nil)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ev, err := webhookverify.ParseEvent([]byte(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			switch tc.typ.(type) {
			case *webhookverify.ChargeExpired:
				if _, ok := ev.(*webhookverify.ChargeExpired); !ok {
					t.Fatalf("got %T", ev)
				}
			case *webhookverify.ChargeCompleted:
				if _, ok := ev.(*webhookverify.ChargeCompleted); !ok {
					t.Fatalf("got %T", ev)
				}
			case *webhookverify.UnknownEvent:
				u, ok := ev.(*webhookverify.UnknownEvent)
				if !ok {
					t.Fatalf("got %T", ev)
				}
				if u.EventName() != woovi.WebhookEventMovementConfirmed {
					t.Fatalf("event=%s", u.EventName())
				}
			}
		})
	}
	if _, err := webhookverify.ParseEvent([]byte(`{`)); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestVerifyHMACMissingAndEmptySecret(t *testing.T) {
	t.Parallel()
	if err := webhookverify.VerifyHMAC([]byte(`{}`), "", "secret"); !errors.Is(err, webhookverify.ErrMissingSignature) {
		t.Fatalf("err=%v", err)
	}
	if err := webhookverify.VerifyHMAC([]byte(`{}`), "YWJj", ""); err == nil {
		t.Fatal("expected secret required")
	}
}

func TestVerifyRSAWithGeneratedKey(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	body := []byte(`{"event":"OPENPIX:CHARGE_CREATED"}`)
	sum := sha256.Sum256(body)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	sigB64 := base64.StdEncoding.EncodeToString(sig)
	if err := webhookverify.VerifyRSAWithKey(body, sigB64, string(pubPEM)); err != nil {
		t.Fatal(err)
	}
	// Wrong key via default VerifyRSA should fail with invalid signature.
	if err := webhookverify.VerifyRSA(body, sigB64); !errors.Is(err, webhookverify.ErrInvalidSignature) {
		t.Fatalf("VerifyRSA err=%v", err)
	}
	if err := webhookverify.VerifyRSAWithKey(body, "", string(pubPEM)); !errors.Is(err, webhookverify.ErrMissingSignature) {
		t.Fatalf("err=%v", err)
	}
	if err := webhookverify.VerifyRSAWithKey(body, "!!!", string(pubPEM)); !errors.Is(err, webhookverify.ErrInvalidSignature) {
		t.Fatalf("err=%v", err)
	}
	if err := webhookverify.VerifyRSAWithKey(body, sigB64, "not-pem"); err == nil {
		t.Fatal("expected pem error")
	}
}

func TestHandlerRSAAndRejects(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))

	body := []byte(`{"event":"OPENPIX:CHARGE_EXPIRED","charge":{"correlationID":"x"}}`)
	sum := sha256.Sum256(body)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	secret := "s3cret"
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(body)
	hmacSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	requireRSA := true
	h := webhookverify.Handler(webhookverify.Options{
		HMACSecret:   secret,
		RequireRSA:   &requireRSA,
		PublicKeyPEM: pubPEM,
	}, func(w http.ResponseWriter, r *http.Request, event webhookverify.Event) {
		if _, ok := event.(*webhookverify.ChargeExpired); !ok {
			t.Errorf("unexpected %T", event)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(webhookverify.HeaderSignature, sigB64)
	req.Header.Set(webhookverify.HeaderHMACSignature, hmacSig)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	bad := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	bad.Header.Set(webhookverify.HeaderSignature, "YmFk")
	badRec := httptest.NewRecorder()
	h.ServeHTTP(badRec, bad)
	if badRec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", badRec.Code)
	}

	noHMAC := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	noHMAC.Header.Set(webhookverify.HeaderSignature, sigB64)
	noHMACRec := httptest.NewRecorder()
	h.ServeHTTP(noHMACRec, noHMAC)
	if noHMACRec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", noHMACRec.Code)
	}
}

func TestHandlerInvalidPayloadAndHMACOptional(t *testing.T) {
	t.Parallel()
	requireRSA := false
	requireHMAC := false
	h := webhookverify.Handler(webhookverify.Options{
		RequireRSA:  &requireRSA,
		RequireHMAC: &requireHMAC,
	}, func(w http.ResponseWriter, r *http.Request, event webhookverify.Event) {
		w.WriteHeader(http.StatusAccepted)
	})

	bad := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`{`)))
	badRec := httptest.NewRecorder()
	h.ServeHTTP(badRec, bad)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", badRec.Code)
	}

	okReq := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`{"event":"OPENPIX:CHARGE_CREATED"}`)))
	okRec := httptest.NewRecorder()
	h.ServeHTTP(okRec, okReq)
	if okRec.Code != http.StatusAccepted {
		t.Fatalf("status=%d", okRec.Code)
	}
}
