package webhookverify

import (
	"io"
	"net/http"
)

// HandlerFunc handles a verified webhook after signature checks.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, event Event)

// Options configures webhook signature verification for Handler.
type Options struct {
	// HMACSecret enables X-OpenPix-Signature verification when set.
	HMACSecret string
	// RequireRSA requires a valid x-webhook-signature. Default true.
	RequireRSA *bool
	// RequireHMAC requires HMAC when HMACSecret is set. Default true if secret set.
	RequireHMAC *bool
	// PublicKeyPEM overrides the default Woovi public key.
	PublicKeyPEM string
}

// Handler verifies signatures, parses the event, and calls fn (fn writes the response).
func Handler(opts Options, fn HandlerFunc) http.Handler {
	requireRSA := true
	if opts.RequireRSA != nil {
		requireRSA = *opts.RequireRSA
	}
	requireHMAC := opts.HMACSecret != ""
	if opts.RequireHMAC != nil {
		requireHMAC = *opts.RequireHMAC
	}
	pub := opts.PublicKeyPEM
	if pub == "" {
		pub = DefaultPublicKeyPEM
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		if requireRSA {
			sig := r.Header.Get(HeaderSignature)
			if err := VerifyRSAWithKey(body, sig, pub); err != nil {
				http.Error(w, "invalid signature", http.StatusUnauthorized)
				return
			}
		}
		if requireHMAC {
			sig := r.Header.Get(HeaderHMACSignature)
			if sig == "" {
				sig = r.Header.Get("x-openpix-signature")
			}
			if err := VerifyHMAC(body, sig, opts.HMACSecret); err != nil {
				http.Error(w, "invalid hmac", http.StatusUnauthorized)
				return
			}
		}

		event, err := ParseEvent(body)
		if err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		fn(w, r, event)
	})
}
