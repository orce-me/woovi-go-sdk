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

// TransactionType classifies a ledger movement.
type TransactionType string

const (
	TransactionTypePayment  TransactionType = "PAYMENT"
	TransactionTypeRefund   TransactionType = "REFUND"
	TransactionTypeWithdraw TransactionType = "WITHDRAW"
)

// TransactionsService is GET /api/v1/transaction.
type TransactionsService struct {
	client *Client
}

// TransactionCharge is charge summary nested on a transaction.
type TransactionCharge struct {
	Status        ChargeStatus `json:"status,omitempty"`
	Customer      string       `json:"customer,omitempty"`
	CorrelationID string       `json:"correlationID,omitempty"`
	CreatedAt     time.Time    `json:"createdAt,omitempty"`
	UpdatedAt     time.Time    `json:"updatedAt,omitempty"`
}

// TransactionWithdraw is withdraw details nested on a transaction (endToEndId).
type TransactionWithdraw struct {
	Value       Money     `json:"value,omitempty"`
	Time        time.Time `json:"time,omitempty"`
	InfoPagador string    `json:"infoPagador,omitempty"`
	EndToEndID  string    `json:"endToEndId,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}

// Transaction is a ledger movement (value in cents; transactionID or endToEndId).
type Transaction struct {
	Customer      *Customer            `json:"customer,omitempty"`
	Payer         *Customer            `json:"payer,omitempty"`
	Charge        *TransactionCharge   `json:"charge,omitempty"`
	Withdraw      *TransactionWithdraw `json:"withdraw,omitempty"`
	Type          TransactionType      `json:"type,omitempty"`
	InfoPagador   string               `json:"infoPagador,omitempty"`
	Value         Money                `json:"value"`
	Time          time.Time            `json:"time,omitempty"`
	TransactionID string               `json:"transactionID,omitempty"`
	EndToEndID    string               `json:"endToEndId,omitempty"`
	GlobalID      string               `json:"globalID,omitempty"`
}

// TransactionListParams filters transactions (RFC3339 range; charge/QR/withdraw ids).
type TransactionListParams struct {
	Skip       int
	Limit      int
	Start      string // RFC3339
	End        string // RFC3339
	Charge     string
	PixQRCode  string
	Withdrawal string
}

// TransactionList is one page of GET /transaction.
type TransactionList struct {
	Transactions []Transaction `json:"transactions"`
	PageInfo     PageInfo      `json:"pageInfo"`
	Skip         int
	Limit        int
}

type transactionEnvelope struct {
	Transaction Transaction `json:"transaction"`
}

// Get retrieves a transaction by Woovi transactionID or bank endToEndId.
func (s *TransactionsService) Get(ctx context.Context, id string, opts ...RequestOption) (*Transaction, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out transactionEnvelope
	path := "/api/v1/transaction/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Transaction, nil
}

func (s *TransactionsService) List(ctx context.Context, params *TransactionListParams, opts ...RequestOption) (*TransactionList, error) {
	if params == nil {
		params = &TransactionListParams{}
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
	if params.Charge != "" {
		q.Set("charge", params.Charge)
	}
	if params.PixQRCode != "" {
		q.Set("pixQrCode", params.PixQRCode)
	}
	if params.Withdrawal != "" {
		q.Set("withdrawal", params.Withdrawal)
	}

	var out TransactionList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/transaction", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *TransactionsService) ListAll(ctx context.Context, params *TransactionListParams, opts ...RequestOption) iter.Seq2[*Transaction, error] {
	return func(yield func(*Transaction, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Transactions {
				if !yield(&page.Transactions[i], nil) {
					return
				}
			}
		}
	}
}

func (s *TransactionsService) ListPages(ctx context.Context, params *TransactionListParams, opts ...RequestOption) iter.Seq2[*TransactionList, error] {
	return func(yield func(*TransactionList, error) bool) {
		p := TransactionListParams{}
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
			if !listPageHasMore(page.PageInfo, len(page.Transactions), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}
