package woovi

import (
	"context"
	"net/http"
)

// CompaniesService is GET /api/v1/company.
type CompaniesService struct {
	client *Client
}

// Company is the authenticated merchant from GET /company.
type Company struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	TaxID string `json:"taxID,omitempty"`
}

type companyEnvelope struct {
	Company Company `json:"company"`
}

func (s *CompaniesService) Get(ctx context.Context, opts ...RequestOption) (*Company, error) {
	var out companyEnvelope
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/company", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Company, nil
}
