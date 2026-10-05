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

// RefundStatus is IN_PROCESSING, CONFIRMED, or REJECTED.
type RefundStatus string

const (
	RefundStatusInProcessing RefundStatus = "IN_PROCESSING"
	RefundStatusConfirmed    RefundStatus = "CONFIRMED"
	RefundStatusRejected     RefundStatus = "REJECTED"
)

// RefundsService is /api/v1/refund (by endToEndId).
type RefundsService struct {
	client *Client
}

// RefundCreateParams refunds a Pix by transaction endToEndId (value in cents).
type RefundCreateParams struct {
	CorrelationID         string `json:"correlationID"`
	TransactionEndToEndID string `json:"transactionEndToEndId"`
	Value                 Money  `json:"value"`
	Comment               string `json:"comment,omitempty"`
}

// ChargeRefundCreateParams refunds a charge. Omit Value to refund the remaining amount.
type ChargeRefundCreateParams struct {
	CorrelationID string `json:"correlationID"`
	Value         *Money `json:"value,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

// Refund is a Pix refund (value in cents; correlationID / endToEndId).
type Refund struct {
	Status        RefundStatus `json:"status"`
	Value         Money        `json:"value"`
	CorrelationID string       `json:"correlationID"`
	RefundID      string       `json:"refundId,omitempty"`
	EndToEndID    string       `json:"endToEndId,omitempty"`
	Time          time.Time    `json:"time,omitempty"`
	Comment       string       `json:"comment,omitempty"`
}

// RefundListParams paginates GET /refund.
type RefundListParams struct {
	Skip  int
	Limit int
}

// RefundList is one page of GET /refund.
type RefundList struct {
	Refunds  []Refund `json:"refunds"`
	PageInfo PageInfo `json:"pageInfo"`
	Skip     int
	Limit    int
}

type refundEnvelope struct {
	Refund Refund `json:"refund"`
}

type chargeRefundsEnvelope struct {
	Refunds []Refund `json:"refunds"`
}

// Create refunds a received Pix by endToEndId.
func (s *RefundsService) Create(ctx context.Context, params *RefundCreateParams, opts ...RequestOption) (*Refund, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: RefundCreateParams is required")
	}
	if params.CorrelationID == "" || params.TransactionEndToEndID == "" {
		return nil, fmt.Errorf("woovi: correlationID and transactionEndToEndId are required")
	}
	if params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}

	var out refundEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/refund", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Refund, nil
}

// Get retrieves a refund by id or correlationID.
func (s *RefundsService) Get(ctx context.Context, id string, opts ...RequestOption) (*Refund, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out refundEnvelope
	path := "/api/v1/refund/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Refund, nil
}

func (s *RefundsService) List(ctx context.Context, params *RefundListParams, opts ...RequestOption) (*RefundList, error) {
	if params == nil {
		params = &RefundListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out RefundList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/refund", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *RefundsService) ListAll(ctx context.Context, params *RefundListParams, opts ...RequestOption) iter.Seq2[*Refund, error] {
	return func(yield func(*Refund, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Refunds {
				if !yield(&page.Refunds[i], nil) {
					return
				}
			}
		}
	}
}

func (s *RefundsService) ListPages(ctx context.Context, params *RefundListParams, opts ...RequestOption) iter.Seq2[*RefundList, error] {
	return func(yield func(*RefundList, error) bool) {
		p := RefundListParams{}
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
			if !listPageHasMore(page.PageInfo, len(page.Refunds), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}

func (s *ChargesService) Refund(ctx context.Context, chargeID string, params *ChargeRefundCreateParams, opts ...RequestOption) (*Refund, error) {
	if chargeID == "" {
		return nil, fmt.Errorf("woovi: chargeID is required")
	}
	if params == nil {
		return nil, fmt.Errorf("woovi: ChargeRefundCreateParams is required")
	}
	if params.CorrelationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	if params.Value != nil && *params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}

	var out refundEnvelope
	path := "/api/v1/charge/" + url.PathEscape(chargeID) + "/refund"
	if err := s.client.do(ctx, http.MethodPost, path, nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Refund, nil
}

func (s *ChargesService) ListRefunds(ctx context.Context, chargeID string, opts ...RequestOption) ([]Refund, error) {
	if chargeID == "" {
		return nil, fmt.Errorf("woovi: chargeID is required")
	}
	var out chargeRefundsEnvelope
	path := "/api/v1/charge/" + url.PathEscape(chargeID) + "/refund"
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return out.Refunds, nil
}
