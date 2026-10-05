package woovi

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// CustomersService is POST/GET/PATCH /api/v1/customer.
type CustomersService struct {
	client *Client
}

// CustomerCreateParams requires name and at least one of TaxID, Email, or Phone.
type CustomerCreateParams struct {
	Name          string   `json:"name"`
	TaxID         string   `json:"taxID,omitempty"`
	Email         string   `json:"email,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	CorrelationID string   `json:"correlationID,omitempty"`
	Address       *Address `json:"address,omitempty"`
}

// CustomerUpdateParams patches a customer. Omitted fields stay unchanged.
type CustomerUpdateParams struct {
	Name    *string  `json:"name,omitempty"`
	Email   *string  `json:"email,omitempty"`
	Phone   *string  `json:"phone,omitempty"`
	TaxID   *string  `json:"taxID,omitempty"`
	Address *Address `json:"address,omitempty"`
}

// CustomerListParams paginates GET /customer.
type CustomerListParams struct {
	Skip  int
	Limit int
}

// CustomerList is one page of GET /customer.
type CustomerList struct {
	Customers []Customer `json:"customers"`
	PageInfo  PageInfo   `json:"pageInfo"`
	Skip      int
	Limit     int
}

type customerEnvelope struct {
	Customer Customer `json:"customer"`
}

func (s *CustomersService) Create(ctx context.Context, params *CustomerCreateParams, opts ...RequestOption) (*Customer, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: CustomerCreateParams is required")
	}
	if params.Name == "" {
		return nil, fmt.Errorf("woovi: name is required")
	}
	if params.TaxID == "" && params.Email == "" && params.Phone == "" {
		return nil, fmt.Errorf("woovi: provide taxID, email, or phone")
	}

	var out customerEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/customer", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Customer, nil
}

// Get retrieves a customer by taxID or correlationID.
func (s *CustomersService) Get(ctx context.Context, id string, opts ...RequestOption) (*Customer, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out customerEnvelope
	path := "/api/v1/customer/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Customer, nil
}

// Update patches a customer by correlationID.
func (s *CustomersService) Update(ctx context.Context, correlationID string, params *CustomerUpdateParams, opts ...RequestOption) (*Customer, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	if params == nil {
		return nil, fmt.Errorf("woovi: CustomerUpdateParams is required")
	}

	var out customerEnvelope
	path := "/api/v1/customer/" + url.PathEscape(correlationID)
	if err := s.client.do(ctx, http.MethodPatch, path, nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Customer, nil
}

func (s *CustomersService) List(ctx context.Context, params *CustomerListParams, opts ...RequestOption) (*CustomerList, error) {
	if params == nil {
		params = &CustomerListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out CustomerList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/customer", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *CustomersService) ListAll(ctx context.Context, params *CustomerListParams, opts ...RequestOption) iter.Seq2[*Customer, error] {
	return func(yield func(*Customer, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Customers {
				if !yield(&page.Customers[i], nil) {
					return
				}
			}
		}
	}
}

func (s *CustomersService) ListPages(ctx context.Context, params *CustomerListParams, opts ...RequestOption) iter.Seq2[*CustomerList, error] {
	return func(yield func(*CustomerList, error) bool) {
		p := CustomerListParams{}
		if params != nil {
			p = *params
		}
		if p.Limit <= 0 {
			p.Limit = 100
		}
		for {
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}
			page, err := s.List(ctx, &p, opts...)
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(page, nil) {
				return
			}
			if !listPageHasMore(page.PageInfo, len(page.Customers), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}
