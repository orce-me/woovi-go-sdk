package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// DecodeService is POST /api/v1/decode/emv.
type DecodeService struct {
	client *Client
}

// EMVMerchantAccountPix is EMV merchant account info (Pix key or URL).
type EMVMerchantAccountPix struct {
	GUI                   string `json:"gui,omitempty"`
	URL                   string `json:"url,omitempty"`
	PixKey                string `json:"pixKey,omitempty"`
	AdditionalInformation string `json:"additionalInformation,omitempty"`
}

// EMVAdditionalData is the EMV additional data field (reference label).
type EMVAdditionalData struct {
	ReferenceLabel string `json:"referenceLabel,omitempty"`
}

// EMVUnreservedTemplate holds unreserved EMV templates (e.g. Pix Automático URL).
type EMVUnreservedTemplate struct {
	GUI string `json:"gui,omitempty"`
	URL string `json:"url,omitempty"`
}

// EMVPayload is the parsed local EMV / BR Code payload.
type EMVPayload struct {
	PayloadFormatIndicator        string                 `json:"payloadFormatIndicator,omitempty"`
	PointOfInitiationMethod       string                 `json:"pointOfInitiationMethod,omitempty"`
	MerchantAccountInformationPix *EMVMerchantAccountPix `json:"merchantAccountInformationPix,omitempty"`
	MerchantCategoryCode          string                 `json:"merchantCategoryCode,omitempty"`
	TransactionCurrency           string                 `json:"transactionCurrency,omitempty"`
	TransactionAmount             string                 `json:"transactionAmount,omitempty"`
	CountryCode                   string                 `json:"countryCode,omitempty"`
	MerchantName                  string                 `json:"merchantName,omitempty"`
	MerchantCity                  string                 `json:"merchantCity,omitempty"`
	AdditionalDataFieldTemplate   *EMVAdditionalData     `json:"additionalDataFieldTemplate,omitempty"`
	UnreservedTemplates           *EMVUnreservedTemplate `json:"unreservedTemplates,omitempty"`
	CRC                           string                 `json:"crc,omitempty"`
}

// DecodeLocation is a resolved COB/COBV or REC location from the issuer PSP.
type DecodeLocation struct {
	IsValid        bool            `json:"isValid"`
	LocationErrors []string        `json:"locationErrors"`
	Payload        json.RawMessage `json:"payload"`
	URL            string          `json:"url,omitempty"`
}

// DecodeEMVResult is the response of POST /api/v1/decode/emv.
type DecodeEMVResult struct {
	EMV         EMVPayload      `json:"emv"`
	CobLocation *DecodeLocation `json:"cobLocation"`
	RecLocation *DecodeLocation `json:"recLocation"`
}

// EMV decodes a Pix QR Code / Copia e Cola string.
func (s *DecodeService) EMV(ctx context.Context, emv string, opts ...RequestOption) (*DecodeEMVResult, error) {
	if emv == "" {
		return nil, fmt.Errorf("woovi: emv is required")
	}
	body := map[string]string{"emv": emv}
	var out DecodeEMVResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/decode/emv", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
