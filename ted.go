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

// TEDStatus is the lifecycle of a TED transfer.
type TEDStatus string

const (
	TEDStatusPending    TEDStatus = "PENDING"
	TEDStatusScheduled  TEDStatus = "SCHEDULED"
	TEDStatusProcessing TEDStatus = "PROCESSING"
	TEDStatusCompleted  TEDStatus = "COMPLETED"
	TEDStatusFailed     TEDStatus = "FAILED"
	TEDStatusRefunded   TEDStatus = "REFUNDED"
)

// TEDType classifies TED purpose (PAYMENT, WITHDRAW, refund variants).
type TEDType string

const (
	TEDTypePayment        TEDType = "PAYMENT"
	TEDTypeWithdraw       TEDType = "WITHDRAW"
	TEDTypeRefundSent     TEDType = "REFUND_SENT"
	TEDTypeRefundReceived TEDType = "REFUND_RECEIVED"
)

// TEDDirection is IN or OUT.
type TEDDirection string

const (
	TEDDirectionIn  TEDDirection = "IN"
	TEDDirectionOut TEDDirection = "OUT"
)

// TEDAccountType is Bacen account type (CACC, SLRY, SVGS, TRAN).
type TEDAccountType string

const (
	TEDAccountTypeCACC TEDAccountType = "CACC"
	TEDAccountTypeSLRY TEDAccountType = "SLRY"
	TEDAccountTypeSVGS TEDAccountType = "SVGS"
	TEDAccountTypeTRAN TEDAccountType = "TRAN"
)

// TEDsService is /api/v1/ted.
type TEDsService struct {
	client *Client
}

// TEDParty is sender or receiver bank data on a TED.
type TEDParty struct {
	Name        string         `json:"name,omitempty"`
	Document    string         `json:"document,omitempty"`
	ISPB        string         `json:"ispb,omitempty"`
	Agency      int            `json:"agency,omitempty"`
	Account     int64          `json:"account,omitempty"`
	AccountType TEDAccountType `json:"accountType,omitempty"`
}

// TED is a TED transfer (value in cents; correlationID / NUOP).
type TED struct {
	CorrelationID string       `json:"correlationID"`
	NUOP          string       `json:"nuop,omitempty"`
	Status        TEDStatus    `json:"status,omitempty"`
	Type          TEDType      `json:"type,omitempty"`
	Direction     TEDDirection `json:"direction,omitempty"`
	Value         Money        `json:"value,omitempty"`
	MoveDate      string       `json:"moveDate,omitempty"`
	AccountID     string       `json:"accountId,omitempty"`
	Sender        *TEDParty    `json:"sender,omitempty"`
	Receiver      *TEDParty    `json:"receiver,omitempty"`
	ErrorCode     *string      `json:"errorCode"`
	Reason        *string      `json:"reason"`
	BCBCode       *string      `json:"bcbCode"`
	CreatedAt     time.Time    `json:"createdAt,omitempty"`
	UpdatedAt     time.Time    `json:"updatedAt,omitempty"`
}

// TEDCreateParams creates a TED out (value in cents; receiver required).
type TEDCreateParams struct {
	CorrelationID  string   `json:"correlationID"`
	Value          Money    `json:"value"`
	AccountID      string   `json:"accountId"`
	Receiver       TEDParty `json:"receiver"`
	MoveDate       string   `json:"moveDate,omitempty"` // YYYY-MM-DD
	ClientFinality *int     `json:"clientFinality,omitempty"`
	Description    string   `json:"description,omitempty"`
	Schedule       *bool    `json:"schedule,omitempty"`
}

// TEDListParams filters TEDs (status, type, direction, RFC3339 range).
type TEDListParams struct {
	Skip          int
	Limit         int
	Status        TEDStatus
	Type          TEDType
	Direction     TEDDirection
	CorrelationID string
	AccountID     string
	Start         string // RFC3339
	End           string // RFC3339
}

// TEDList is one page of GET /ted.
type TEDList struct {
	TEDs     []TED    `json:"teds"`
	PageInfo PageInfo `json:"pageInfo"`
	Skip     int
	Limit    int
}

type tedEnvelope struct {
	TED TED `json:"ted"`
}

// Create sends a TED. Reusing correlationID returns the existing TED.
func (s *TEDsService) Create(ctx context.Context, params *TEDCreateParams, opts ...RequestOption) (*TED, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: TEDCreateParams is required")
	}
	if params.CorrelationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	if params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}
	if params.AccountID == "" {
		return nil, fmt.Errorf("woovi: accountId is required")
	}
	if params.Receiver.Document == "" || params.Receiver.ISPB == "" {
		return nil, fmt.Errorf("woovi: receiver document and ispb are required")
	}

	var out tedEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/ted", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.TED, nil
}

func (s *TEDsService) Get(ctx context.Context, correlationID string, opts ...RequestOption) (*TED, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	var out tedEnvelope
	path := "/api/v1/ted/" + url.PathEscape(correlationID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.TED, nil
}

func (s *TEDsService) List(ctx context.Context, params *TEDListParams, opts ...RequestOption) (*TEDList, error) {
	if params == nil {
		params = &TEDListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		return nil, fmt.Errorf("woovi: limit must be at most 100")
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))
	if params.Status != "" {
		q.Set("status", string(params.Status))
	}
	if params.Type != "" {
		q.Set("type", string(params.Type))
	}
	if params.Direction != "" {
		q.Set("direction", string(params.Direction))
	}
	if params.CorrelationID != "" {
		q.Set("correlationID", params.CorrelationID)
	}
	if params.AccountID != "" {
		q.Set("accountId", params.AccountID)
	}
	if params.Start != "" {
		q.Set("start", params.Start)
	}
	if params.End != "" {
		q.Set("end", params.End)
	}

	var out TEDList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/ted", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *TEDsService) ListAll(ctx context.Context, params *TEDListParams, opts ...RequestOption) iter.Seq2[*TED, error] {
	return func(yield func(*TED, error) bool) {
		p := TEDListParams{}
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
			for i := range page.TEDs {
				if !yield(&page.TEDs[i], nil) {
					return
				}
			}
			if !listPageHasMore(page.PageInfo, len(page.TEDs), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}

func (s *TEDsService) Refund(ctx context.Context, correlationID string, opts ...RequestOption) (*TED, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	var out tedEnvelope
	path := "/api/v1/ted/" + url.PathEscape(correlationID) + "/refund"
	if err := s.client.do(ctx, http.MethodPost, path, nil, map[string]any{}, &out, opts...); err != nil {
		return nil, err
	}
	return &out.TED, nil
}
