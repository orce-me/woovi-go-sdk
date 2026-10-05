package woovi

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// InvoiceStatus is the invoice lifecycle (CONFIRMED).
type InvoiceStatus string

const (
	InvoiceStatusConfirmed InvoiceStatus = "CONFIRMED"
)

// InvoicesService is /api/v1/invoice.
type InvoicesService struct {
	client *Client
}

// Invoice is a billing invoice (value in cents; optional charge link).
type Invoice struct {
	ID            string        `json:"id,omitempty"`
	Value         Money         `json:"value,omitempty"`
	Date          time.Time     `json:"date,omitempty"`
	BillingDate   time.Time     `json:"billingDate,omitempty"`
	Status        InvoiceStatus `json:"status,omitempty"`
	StatusRaw     *string       `json:"statusRaw"`
	CorrelationID string        `json:"correlationID,omitempty"`
	Customer      *Customer     `json:"customer,omitempty"`
	Charge        *Charge       `json:"charge,omitempty"`
	Description   string        `json:"description,omitempty"`
}

// InvoiceListParams paginates invoices (optional RFC3339 start/end).
type InvoiceListParams struct {
	Skip  int
	Limit int
	Start string
	End   string
}

// InvoiceList is one page of GET /invoice.
type InvoiceList struct {
	Invoices []Invoice `json:"invoices"`
	PageInfo PageInfo  `json:"pageInfo"`
	Skip     int
	Limit    int
}

// InvoiceDocument is a downloaded invoice PDF or XML (raw bytes).
type InvoiceDocument struct {
	Data        []byte
	ContentType string
}

// InvoiceIntegration holds /invoice/integration fields (shape varies by provider).
type InvoiceIntegration struct {
	Raw map[string]any `json:"-"`
}

func (s *InvoicesService) List(ctx context.Context, params *InvoiceListParams, opts ...RequestOption) (*InvoiceList, error) {
	if params == nil {
		params = &InvoiceListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))
	if params.Start != "" {
		q.Set("start", params.Start)
	}
	if params.End != "" {
		q.Set("end", params.End)
	}
	var out InvoiceList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/invoice", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *InvoicesService) ListAll(ctx context.Context, params *InvoiceListParams, opts ...RequestOption) iter.Seq2[*Invoice, error] {
	return func(yield func(*Invoice, error) bool) {
		p := InvoiceListParams{}
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
			for i := range page.Invoices {
				if !yield(&page.Invoices[i], nil) {
					return
				}
			}
			if !listPageHasMore(page.PageInfo, len(page.Invoices), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}

// InvoiceCreateCustomer is inline customer data on invoice create.
type InvoiceCreateCustomer struct {
	TaxID   string   `json:"taxID,omitempty"`
	Name    string   `json:"name,omitempty"`
	Email   string   `json:"email,omitempty"`
	Phone   string   `json:"phone,omitempty"`
	Address *Address `json:"address,omitempty"`
}

// InvoiceCreateParams creates an invoice; value in cents when set.
type InvoiceCreateParams struct {
	Description   string                 `json:"description,omitempty"`
	BillingDate   *time.Time             `json:"billingDate,omitempty"`
	CorrelationID string                 `json:"correlationID,omitempty"`
	Charge        string                 `json:"charge,omitempty"`
	Value         *Money                 `json:"value,omitempty"`
	CustomerID    string                 `json:"customerId,omitempty"`
	Customer      *InvoiceCreateCustomer `json:"customer,omitempty"`
}

type invoiceEnvelope struct {
	Invoice Invoice `json:"invoice"`
}

func (s *InvoicesService) Create(ctx context.Context, params *InvoiceCreateParams, opts ...RequestOption) (*Invoice, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: InvoiceCreateParams is required")
	}
	var out invoiceEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/invoice", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Invoice, nil
}

func (s *InvoicesService) Cancel(ctx context.Context, correlationID string, opts ...RequestOption) error {
	if correlationID == "" {
		return fmt.Errorf("woovi: correlationID is required")
	}
	path := "/api/v1/invoice/" + url.PathEscape(correlationID) + "/cancel"
	var out struct {
		Success bool `json:"success"`
	}
	return s.client.do(ctx, http.MethodPost, path, nil, map[string]any{}, &out, opts...)
}

func (s *InvoicesService) PDF(ctx context.Context, correlationID string, opts ...RequestOption) (*InvoiceDocument, error) {
	return s.document(ctx, correlationID, "pdf", "application/pdf", opts...)
}

func (s *InvoicesService) XML(ctx context.Context, correlationID string, opts ...RequestOption) (*InvoiceDocument, error) {
	return s.document(ctx, correlationID, "xml", "application/xml", opts...)
}

func (s *InvoicesService) document(ctx context.Context, correlationID, kind, defaultType string, opts ...RequestOption) (*InvoiceDocument, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	path := "/api/v1/invoice/" + url.PathEscape(correlationID) + "/" + kind
	data, contentType, err := s.client.doBytes(ctx, http.MethodGet, path, nil, opts...)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = defaultType
	}
	return &InvoiceDocument{Data: data, ContentType: contentType}, nil
}

func (s *InvoicesService) GetIntegration(ctx context.Context, opts ...RequestOption) (map[string]any, error) {
	var out map[string]any
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/invoice/integration", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *InvoicesService) UpsertIntegration(ctx context.Context, body any, opts ...RequestOption) (map[string]any, error) {
	if body == nil {
		return nil, fmt.Errorf("woovi: integration body is required")
	}
	var out map[string]any
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/invoice/integration", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *InvoicesService) UpdateIntegration(ctx context.Context, body any, opts ...RequestOption) (map[string]any, error) {
	if body == nil {
		return nil, fmt.Errorf("woovi: integration body is required")
	}
	var out map[string]any
	if err := s.client.do(ctx, http.MethodPut, "/api/v1/invoice/integration", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// PatchIntegration patches invoice integration fields (e.g. isActive).
func (s *InvoicesService) PatchIntegration(ctx context.Context, body any, opts ...RequestOption) (map[string]any, error) {
	if body == nil {
		return nil, fmt.Errorf("woovi: integration body is required")
	}
	var out map[string]any
	if err := s.client.do(ctx, http.MethodPatch, "/api/v1/invoice/integration", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// InvoiceCertificateParams uploads an A1 certificate for invoice integration.
type InvoiceCertificateParams struct {
	Pcks12     string `json:"pcks12"`
	Passphrase string `json:"passphrase"`
	Test       bool   `json:"test,omitempty"`
}

// UploadCertificate configures the A1 pkcs12 certificate.
func (s *InvoicesService) UploadCertificate(ctx context.Context, params *InvoiceCertificateParams, opts ...RequestOption) (map[string]any, error) {
	if params == nil || params.Pcks12 == "" || params.Passphrase == "" {
		return nil, fmt.Errorf("woovi: pcks12 and passphrase are required")
	}
	var out map[string]any
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/invoice/integration/certificate", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *InvoicesService) TestIntegration(ctx context.Context, opts ...RequestOption) (map[string]any, error) {
	var out map[string]any
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/invoice/integration/test", nil, map[string]any{}, &out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
