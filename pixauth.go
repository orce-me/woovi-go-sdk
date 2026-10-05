package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// PixAuthStatus is ACTIVE, COMPLETED, EXPIRED, or FAILED.
type PixAuthStatus string

const (
	PixAuthStatusActive    PixAuthStatus = "ACTIVE"
	PixAuthStatusCompleted PixAuthStatus = "COMPLETED"
	PixAuthStatusExpired   PixAuthStatus = "EXPIRED"
	PixAuthStatusFailed    PixAuthStatus = "FAILED"
)

// PixAuthResult is UNVERIFIED, MATCHED, or MISMATCH after the 1-cent Pix.
type PixAuthResult string

const (
	PixAuthResultUnverified PixAuthResult = "UNVERIFIED"
	PixAuthResultMatched    PixAuthResult = "MATCHED"
	PixAuthResultMismatch   PixAuthResult = "MISMATCH"
)

// PixAuthsService validates CPF/CNPJ via a 1-cent Pix.
type PixAuthsService struct {
	client *Client
}

// PixAuthCreateParams starts Pix auth for a taxID (correlationID required).
type PixAuthCreateParams struct {
	CorrelationID string `json:"correlationID"`
	TaxID         string `json:"taxID"`
	Name          string `json:"name,omitempty"`
}

// PixAuthTaxID is the CPF/CNPJ under authentication.
type PixAuthTaxID struct {
	TaxID string `json:"taxID,omitempty"`
	Type  string `json:"type,omitempty"`
}

// PixAuth is a 1-cent Pix authentication session.
type PixAuth struct {
	ID            string        `json:"id,omitempty"`
	CorrelationID string        `json:"correlationID,omitempty"`
	Status        PixAuthStatus `json:"status,omitempty"`
	Result        PixAuthResult `json:"result,omitempty"`
	TaxID         *PixAuthTaxID `json:"taxID,omitempty"`
	Amount        Money         `json:"amount,omitempty"`
	DueDate       time.Time     `json:"dueDate,omitempty"`
	CompletedAt   *time.Time    `json:"completedAt,omitempty"`
	CreatedAt     time.Time     `json:"createdAt,omitempty"`
}

// PixAuthCreateResult is POST/GET /pix-auth (pixAuth, brCode, hostedUrl).
type PixAuthCreateResult struct {
	PixAuth   PixAuth `json:"pixAuth"`
	BrCode    string  `json:"brCode,omitempty"`
	HostedURL string  `json:"hostedUrl,omitempty"`
}

func (s *PixAuthsService) Create(ctx context.Context, params *PixAuthCreateParams, opts ...RequestOption) (*PixAuthCreateResult, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: PixAuthCreateParams is required")
	}
	if params.CorrelationID == "" || params.TaxID == "" {
		return nil, fmt.Errorf("woovi: correlationID and taxID are required")
	}
	var out PixAuthCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/pix-auth", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a Pix authentication by id or correlationID.
func (s *PixAuthsService) Get(ctx context.Context, id string, opts ...RequestOption) (*PixAuthCreateResult, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out PixAuthCreateResult
	path := "/api/v1/pix-auth/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
