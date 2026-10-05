package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CashbackType is IMMEDIATE or LOYALTY on charge cashback.
type CashbackType string

const (
	CashbackTypeImmediate CashbackType = "IMMEDIATE"
	CashbackTypeLoyalty   CashbackType = "LOYALTY"
)

// ChargeCashback overrides account cashback settings on charge create.
type ChargeCashback struct {
	Value      *Money       `json:"value,omitempty"`
	Percentage *float64     `json:"percentage,omitempty"`
	Type       CashbackType `json:"type,omitempty"`
}

// CashbackFidelitiesService is /api/v1/cashback-fidelity.
type CashbackFidelitiesService struct {
	client *Client
}

// CashbackFidelityCreateParams credits loyalty cashback to a customer.
type CashbackFidelityCreateParams struct {
	TaxID string `json:"taxID"`
	Value Money  `json:"value"`
}

// CashbackFidelity is a loyalty cashback grant (value in cents).
type CashbackFidelity struct {
	Value Money `json:"value"`
}

// CashbackFidelityCreateResult is POST /cashback-fidelity.
type CashbackFidelityCreateResult struct {
	Cashback CashbackFidelity `json:"cashback"`
	Message  string           `json:"message,omitempty"`
}

// CashbackFidelityBalance is GET /cashback-fidelity/balance/{taxID} (cents).
type CashbackFidelityBalance struct {
	Balance Money  `json:"balance"`
	Status  string `json:"status,omitempty"`
}

// Create grants exclusive cashback. If a pending grant exists, the API returns it.
func (s *CashbackFidelitiesService) Create(ctx context.Context, params *CashbackFidelityCreateParams, opts ...RequestOption) (*CashbackFidelityCreateResult, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: CashbackFidelityCreateParams is required")
	}
	if params.TaxID == "" {
		return nil, fmt.Errorf("woovi: taxID is required")
	}
	if params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}
	var out CashbackFidelityCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/cashback-fidelity", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *CashbackFidelitiesService) Balance(ctx context.Context, taxID string, opts ...RequestOption) (*CashbackFidelityBalance, error) {
	if taxID == "" {
		return nil, fmt.Errorf("woovi: taxID is required")
	}
	var out CashbackFidelityBalance
	path := "/api/v1/cashback-fidelity/balance/" + url.PathEscape(taxID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
