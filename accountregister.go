package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// AccountRegistersService is GET/DELETE /api/v1/account-register. Prefer KYC onboarding for new integrations.
type AccountRegistersService struct {
	client *Client
}

// AccountRegister is a legacy merchant register keyed by correlationID.
type AccountRegister struct {
	OfficialName  string          `json:"officialName,omitempty"`
	TradeName     string          `json:"tradeName,omitempty"`
	Type          string          `json:"type,omitempty"`
	TaxID         *KYCTaxID       `json:"taxID,omitempty"`
	CorrelationID string          `json:"correlationID,omitempty"`
	Status        string          `json:"status,omitempty"`
	Raw           json.RawMessage `json:"-"`
}

// AccountRegisterDeleteResult is DELETE /account-register/{id}.
type AccountRegisterDeleteResult struct {
	Message           string `json:"message,omitempty"`
	AccountRegisterID string `json:"accountRegisterId,omitempty"`
}

func (s *AccountRegistersService) Get(ctx context.Context, correlationID string, opts ...RequestOption) (*AccountRegister, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/account-register/" + url.PathEscape(correlationID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	out := &AccountRegister{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

func (s *AccountRegistersService) Delete(ctx context.Context, correlationID string, opts ...RequestOption) (*AccountRegisterDeleteResult, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	var out AccountRegisterDeleteResult
	path := "/api/v1/account-register/" + url.PathEscape(correlationID)
	if err := s.client.do(ctx, http.MethodDelete, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
