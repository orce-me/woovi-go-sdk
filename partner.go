package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// PartnersService is /api/v1/partner.
type PartnersService struct {
	client *Client
}

// PartnerCompany is a company visible to a partner AppID.
type PartnerCompany struct {
	CompanyID   string `json:"company_id"`
	CompanyName string `json:"company_name"`
	TaxID       string `json:"tax_id,omitempty"`
}

// PartnerTaxID is CPF/CNPJ with type on partner registration.
type PartnerTaxID struct {
	TaxID string `json:"taxID"`
	Type  string `json:"type"`
}

// PartnerPreRegistration is company data for partner create.
type PartnerPreRegistration struct {
	Name    string       `json:"name"`
	TaxID   PartnerTaxID `json:"taxID"`
	Website string       `json:"website"`
}

// PartnerUser is the primary user on partner create.
type PartnerUser struct {
	FirstName string       `json:"firstName"`
	LastName  string       `json:"lastName"`
	Email     string       `json:"email"`
	Phone     string       `json:"phone"`
	TaxID     PartnerTaxID `json:"taxID"`
}

// PartnerCreateParams registers a partner company and user.
type PartnerCreateParams struct {
	PreRegistration PartnerPreRegistration `json:"preRegistration"`
	User            PartnerUser            `json:"user"`
}

// PartnerCreateResult is POST /partner.
type PartnerCreateResult struct {
	PreRegistration PartnerPreRegistration `json:"preRegistration"`
	User            PartnerUser            `json:"user"`
}

func (s *PartnersService) ListCompanies(ctx context.Context, opts ...RequestOption) ([]PartnerCompany, error) {
	var out struct {
		CompanyList []PartnerCompany `json:"company_list"`
		Companies   []PartnerCompany `json:"companies"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/partner/company", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	if len(out.CompanyList) > 0 {
		return out.CompanyList, nil
	}
	return out.Companies, nil
}

func (s *PartnersService) GetCompany(ctx context.Context, taxID string, opts ...RequestOption) (*PartnerCompany, error) {
	if taxID == "" {
		return nil, fmt.Errorf("woovi: taxID is required")
	}
	var out struct {
		CompanyDetails *PartnerCompany `json:"company_details"`
		Company        *PartnerCompany `json:"company"`
	}
	path := "/api/v1/partner/company/" + url.PathEscape(taxID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	switch {
	case out.CompanyDetails != nil:
		return out.CompanyDetails, nil
	case out.Company != nil:
		return out.Company, nil
	default:
		return nil, fmt.Errorf("woovi: empty partner company response")
	}
}

// PartnerAffiliateAccount is bank/App credentials for an affiliate.
type PartnerAffiliateAccount struct {
	ClientID  string `json:"clientId,omitempty"`
	Name      string `json:"name,omitempty"`
	AccountID string `json:"accountId,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Account   string `json:"account,omitempty"`
}

// PartnerAffiliate is one affiliated company under a partner.
type PartnerAffiliate struct {
	Company Company                  `json:"company"`
	Account *PartnerAffiliateAccount `json:"account,omitempty"`
}

// PartnerAffiliateList is GET /partner/affiliate.
type PartnerAffiliateList struct {
	Affiliates []PartnerAffiliate `json:"affiliates"`
	PageInfo   PageInfo           `json:"pageInfo"`
}

func (s *PartnersService) ListAffiliates(ctx context.Context, opts ...RequestOption) (*PartnerAffiliateList, error) {
	var out PartnerAffiliateList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/partner/affiliate", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PartnersService) CreateApplication(ctx context.Context, params *PartnerCreateParams, opts ...RequestOption) (*PartnerCreateResult, error) {
	if err := validatePartnerCreate(params); err != nil {
		return nil, err
	}

	var out PartnerCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/partner/application", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PartnersService) CreateCompany(ctx context.Context, params *PartnerCreateParams, opts ...RequestOption) (*PartnerCreateResult, error) {
	if err := validatePartnerCreate(params); err != nil {
		return nil, err
	}
	var out PartnerCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/partner/company", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func validatePartnerCreate(params *PartnerCreateParams) error {
	if params == nil {
		return fmt.Errorf("woovi: PartnerCreateParams is required")
	}
	if params.PreRegistration.Name == "" || params.PreRegistration.TaxID.TaxID == "" {
		return fmt.Errorf("woovi: preRegistration name and taxID are required")
	}
	if params.User.Email == "" || params.User.FirstName == "" {
		return fmt.Errorf("woovi: user firstName and email are required")
	}
	return nil
}
