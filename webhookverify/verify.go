package webhookverify

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// HeaderSignature is the Woovi origin RSA signature header.
const HeaderSignature = "x-webhook-signature"

// HeaderHMACSignature is the per-webhook HMAC signature header.
const HeaderHMACSignature = "X-OpenPix-Signature"

// DefaultPublicKeyPEM is the Woovi webhook RSA public key from the docs.
const DefaultPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC/+NtIkjzevvqD+I3MMv3bLXDt
pvxBjY4BsRrSdca3rtAwMcRYYvxSnd7jagVLpctMiOxQO8ieUCKLSWHpsMAjO/zZ
WMKbqoG8MNpi/u3fp6zz0mcHCOSqYsPUUG19buW8bis5ZZ2IZgBObWSpTvJ0cnj6
HKBAA82Jln+lGwS1MwIDAQAB
-----END PUBLIC KEY-----
`

var (
	// ErrInvalidSignature means the signature does not match the body.
	ErrInvalidSignature = errors.New("webhookverify: invalid signature")
	// ErrMissingSignature means the required signature header is absent.
	ErrMissingSignature = errors.New("webhookverify: missing signature")
)

// VerifyRSA checks x-webhook-signature. payload must be raw body bytes.
func VerifyRSA(payload []byte, signatureBase64 string) error {
	return VerifyRSAWithKey(payload, signatureBase64, DefaultPublicKeyPEM)
}

func VerifyRSAWithKey(payload []byte, signatureBase64, publicKeyPEM string) error {
	if signatureBase64 == "" {
		return ErrMissingSignature
	}
	sig, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("%w: decode signature: %w", ErrInvalidSignature, err)
	}
	pub, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig); err != nil {
		return ErrInvalidSignature
	}
	return nil
}

// VerifyHMAC checks X-OpenPix-Signature (HMAC-SHA1). payload must be raw body bytes.
func VerifyHMAC(payload []byte, signatureBase64, secret string) error {
	if signatureBase64 == "" {
		return ErrMissingSignature
	}
	if secret == "" {
		return fmt.Errorf("webhookverify: hmac secret is required")
	}
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(payload)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signatureBase64)) {
		return ErrInvalidSignature
	}
	return nil
}

func parseRSAPublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("webhookverify: invalid public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("webhookverify: parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("webhookverify: public key is not RSA")
	}
	return rsaPub, nil
}
