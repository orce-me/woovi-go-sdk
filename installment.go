package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// InstallmentStatus is the lifecycle of a Pix Automático installment.
type InstallmentStatus string

const (
	InstallmentStatusActive    InstallmentStatus = "ACTIVE"
	InstallmentStatusScheduled InstallmentStatus = "SCHEDULED"
	InstallmentStatusCompleted InstallmentStatus = "COMPLETED"
	InstallmentStatusCanceled  InstallmentStatus = "CANCELED"
)

// InstallmentsService is /api/v1/installments (Pix Automático).
type InstallmentsService struct {
	client *Client
}

// InstallmentCobr is the cobrança generated for one installment.
type InstallmentCobr struct {
	IdentifierID  string    `json:"identifierId,omitempty"`
	RecurrencyID  string    `json:"recurrencyId,omitempty"`
	InstallmentID string    `json:"installmentId,omitempty"`
	Status        string    `json:"status,omitempty"`
	Value         Money     `json:"value,omitempty"`
	CreatedAt     time.Time `json:"createdAt,omitempty"`
}

// Installment is one Pix Automático installment (value in cents).
type Installment struct {
	DateGenerateCharge          time.Time         `json:"dateGenerateCharge,omitempty"`
	Expiration                  int               `json:"expiration,omitempty"`
	InstallmentNumber           int               `json:"installmentNumber,omitempty"`
	Value                       Money             `json:"value"`
	Status                      InstallmentStatus `json:"status,omitempty"`
	CreatedAt                   time.Time         `json:"createdAt,omitempty"`
	Cobr                        *InstallmentCobr  `json:"cobr,omitempty"`
	PaymentSubscriptionGlobalID string            `json:"paymentSubscriptionGlobalID,omitempty"`
	GlobalID                    string            `json:"globalID,omitempty"`
}

// InstallmentCobrParams optionally overrides value on create/retry cobr.
type InstallmentCobrParams struct {
	Value *Money `json:"value,omitempty"`
}

// InstallmentList is GET /installments.
type InstallmentList struct {
	Installments []Installment `json:"installments"`
	PageInfo     PageInfo      `json:"pageInfo"`
	Skip         int
	Limit        int
}

func (s *InstallmentsService) Get(ctx context.Context, id string, opts ...RequestOption) (*Installment, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out Installment
	path := "/api/v1/installments/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *InstallmentsService) CreateCobr(ctx context.Context, id string, params *InstallmentCobrParams, opts ...RequestOption) (*Installment, error) {
	return s.cobr(ctx, http.MethodPost, id, "/cobr", params, opts...)
}

func (s *InstallmentsService) RetryCobr(ctx context.Context, id string, params *InstallmentCobrParams, opts ...RequestOption) (*Installment, error) {
	return s.cobr(ctx, http.MethodPost, id, "/cobr/retry", params, opts...)
}

func (s *InstallmentsService) cobr(ctx context.Context, method, id, suffix string, params *InstallmentCobrParams, opts ...RequestOption) (*Installment, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var body any
	if params != nil {
		body = params
	}
	var out Installment
	path := "/api/v1/installments/" + url.PathEscape(id) + suffix
	if err := s.client.do(ctx, method, path, nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SubscriptionsService) ListInstallments(ctx context.Context, subscriptionID string, skip, limit int, opts ...RequestOption) (*InstallmentList, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("woovi: subscriptionID is required")
	}
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out InstallmentList
	path := "/api/v1/subscriptions/" + url.PathEscape(subscriptionID) + "/installments"
	if err := s.client.do(ctx, http.MethodGet, path, q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}
