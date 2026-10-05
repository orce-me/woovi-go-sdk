package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode"
)

// BoletosService is POST /api/v1/boleto and boleto transaction APIs.
type BoletosService struct {
	client *Client
}

// BoletoEntity is issuer or beneficiary data on a validated boleto.
type BoletoEntity struct {
	Code  string `json:"code,omitempty"`
	Name  string `json:"name,omitempty"`
	TaxID string `json:"taxID,omitempty"`
}

// Boleto is POST /boleto/validate output (barcode, value in cents, parties).
type Boleto struct {
	Barcode          string       `json:"barcode"`
	ExpiresDate      time.Time    `json:"expiresDate"`
	TotalValue       Money        `json:"totalValue"`
	IssuingEntity    BoletoEntity `json:"issuingEntity"`
	FinalBeneficiary BoletoEntity `json:"finalBeneficiary"`
}

// BoletoTransactionType is BOLETO_IN or BOLETO_OUT.
type BoletoTransactionType string

const (
	BoletoTransactionTypeIn  BoletoTransactionType = "BOLETO_IN"
	BoletoTransactionTypeOut BoletoTransactionType = "BOLETO_OUT"
)

// BoletoTransactionStatus is the lifecycle of a boleto settlement.
type BoletoTransactionStatus string

const (
	BoletoTransactionStatusCreated    BoletoTransactionStatus = "CREATED"
	BoletoTransactionStatusProcessing BoletoTransactionStatus = "PROCESSING"
	BoletoTransactionStatusPending    BoletoTransactionStatus = "PENDING"
	BoletoTransactionStatusConfirmed  BoletoTransactionStatus = "CONFIRMED"
	BoletoTransactionStatusRejected   BoletoTransactionStatus = "REJECTED"
)

// BoletoTransaction is a settled boleto movement (BOLETO_SETTLED webhook).
type BoletoTransaction struct {
	BoletoTransactionID string                  `json:"boletoTransactionID"`
	Type                BoletoTransactionType   `json:"type,omitempty"`
	Status              BoletoTransactionStatus `json:"status,omitempty"`
	Value               Money                   `json:"value,omitempty"`
	Fee                 Money                   `json:"fee,omitempty"`
	CreatedAt           time.Time               `json:"createdAt,omitempty"`
	SettledAt           time.Time               `json:"settledAt,omitempty"`
	FinesValue          Money                   `json:"finesValue,omitempty"`
	InterestsValue      Money                   `json:"interestsValue,omitempty"`
	DiscountValue       Money                   `json:"discountValue,omitempty"`
	Charge              *Charge                 `json:"charge,omitempty"`
}

type boletoValidateEnvelope struct {
	Boleto Boleto `json:"boleto"`
}

type boletoTransactionEnvelope struct {
	BoletoTransaction BoletoTransaction `json:"boletoTransaction"`
}

func (s *BoletosService) Validate(ctx context.Context, barcode string, opts ...RequestOption) (*Boleto, error) {
	if err := validateBoletoBarcode(barcode); err != nil {
		return nil, err
	}
	body := map[string]string{"barcode": barcode}
	var out boletoValidateEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/boleto/validate", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Boleto, nil
}

// BoletoTransactionListParams filters boleto settlements (RFC3339 start/end).
type BoletoTransactionListParams struct {
	Type         BoletoTransactionType
	Status       BoletoTransactionStatus
	Start        string
	End          string
	SettledStart string
	SettledEnd   string
	Skip         int
	Limit        int
}

// BoletoTransactionList is GET /boleto-transaction.
type BoletoTransactionList struct {
	Status             string              `json:"status,omitempty"`
	BoletoTransactions []BoletoTransaction `json:"boletoTransactions"`
	PageInfo           PageInfo            `json:"pageInfo"`
	Skip               int
	Limit              int
}

// GetTransaction retrieves a boleto transaction by id (BOLETO_SETTLED).
func (s *BoletosService) GetTransaction(ctx context.Context, boletoTransactionID string, opts ...RequestOption) (*BoletoTransaction, error) {
	if boletoTransactionID == "" {
		return nil, fmt.Errorf("woovi: boletoTransactionID is required")
	}
	var out boletoTransactionEnvelope
	path := "/api/v1/boleto-transaction/" + url.PathEscape(boletoTransactionID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.BoletoTransaction, nil
}

func (s *BoletosService) ListTransactions(ctx context.Context, params *BoletoTransactionListParams, opts ...RequestOption) (*BoletoTransactionList, error) {
	if params == nil {
		params = &BoletoTransactionListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))
	if params.Type != "" {
		q.Set("type", string(params.Type))
	}
	if params.Status != "" {
		q.Set("status", string(params.Status))
	}
	if params.Start != "" {
		q.Set("start", params.Start)
	}
	if params.End != "" {
		q.Set("end", params.End)
	}
	if params.SettledStart != "" {
		q.Set("settledStart", params.SettledStart)
	}
	if params.SettledEnd != "" {
		q.Set("settledEnd", params.SettledEnd)
	}
	var out BoletoTransactionList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/boleto-transaction", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func validateBoletoBarcode(barcode string) error {
	if barcode == "" {
		return fmt.Errorf("woovi: barcode is required")
	}
	n := 0
	for _, r := range barcode {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("woovi: barcode must contain only digits")
		}
		n++
	}
	switch n {
	case 44, 47, 48:
		return nil
	default:
		return fmt.Errorf("woovi: barcode must have 44, 47, or 48 digits")
	}
}
