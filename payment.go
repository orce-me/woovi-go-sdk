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

// PixKeyType is CPF, CNPJ, EMAIL, PHONE, RANDOM, or EVP.
type PixKeyType string

const (
	PixKeyTypeCPF    PixKeyType = "CPF"
	PixKeyTypeCNPJ   PixKeyType = "CNPJ"
	PixKeyTypeEmail  PixKeyType = "EMAIL"
	PixKeyTypePhone  PixKeyType = "PHONE"
	PixKeyTypeRandom PixKeyType = "RANDOM"
)

// PaymentStatus is CREATED through CONFIRMED, DENIED, or FAILED.
type PaymentStatus string

const (
	PaymentStatusCreated   PaymentStatus = "CREATED"
	PaymentStatusApproved  PaymentStatus = "APPROVED"
	PaymentStatusConfirmed PaymentStatus = "CONFIRMED"
	PaymentStatusDenied    PaymentStatus = "DENIED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

// PaymentType selects payment kind (BOLETO; omit for Pix).
type PaymentType string

const (
	PaymentTypeBoleto PaymentType = "BOLETO"
)

// PaymentsService is POST/GET /api/v1/payment.
type PaymentsService struct {
	client *Client
}

// PaymentCreateParams creates a payment by Pix key, QR code, or boleto.
// Pix key: Value, DestinationAlias, DestinationAliasType, CorrelationID.
// QR: QRCode, CorrelationID. Boleto: Type=BOLETO, BoletoBarcode, CorrelationID.
type PaymentCreateParams struct {
	Value                Money       `json:"value,omitempty"`
	DestinationAlias     string      `json:"destinationAlias,omitempty"`
	DestinationAliasType PixKeyType  `json:"destinationAliasType,omitempty"`
	QRCode               string      `json:"qrCode,omitempty"`
	Type                 PaymentType `json:"type,omitempty"`
	BoletoBarcode        string      `json:"boletoBarcode,omitempty"`
	CorrelationID        string      `json:"correlationID"`
	Comment              string      `json:"comment,omitempty"`
	SourceAccountID      string      `json:"sourceAccountId,omitempty"`
	PixKeyEndToEndID     string      `json:"pixKeyEndToEndId,omitempty"`
	AutoApprove          *bool       `json:"autoApprove,omitempty"`
}

// Payment is a Pix out / boleto payment (value in cents; correlationID).
type Payment struct {
	Value                Money         `json:"value"`
	Status               PaymentStatus `json:"status"`
	DestinationAlias     string        `json:"destinationAlias,omitempty"`
	DestinationAliasType PixKeyType    `json:"destinationAliasType,omitempty"`
	QRCode               string        `json:"qrCode,omitempty"`
	Type                 PaymentType   `json:"type,omitempty"`
	BoletoBarcode        string        `json:"boletoBarcode,omitempty"`
	Comment              string        `json:"comment,omitempty"`
	CorrelationID        string        `json:"correlationID"`
	SourceAccountID      string        `json:"sourceAccountId,omitempty"`
}

// PaymentDestination is resolved payee data after approve.
type PaymentDestination struct {
	Name    string `json:"name,omitempty"`
	TaxID   string `json:"taxID,omitempty"`
	PixKey  string `json:"pixKey,omitempty"`
	Bank    string `json:"bank,omitempty"`
	Branch  string `json:"branch,omitempty"`
	Account string `json:"account,omitempty"`
}

// PaymentApproveResult is POST /payment/approve (payment, transaction, destination).
type PaymentApproveResult struct {
	Payment     Payment             `json:"payment"`
	Transaction *PaymentTransaction `json:"transaction,omitempty"`
	Destination *PaymentDestination `json:"destination,omitempty"`
}

// PaymentTransaction is the settled Pix movement (endToEndId).
type PaymentTransaction struct {
	Value      Money     `json:"value"`
	EndToEndID string    `json:"endToEndId"`
	Time       time.Time `json:"time"`
}

// PaymentListParams paginates GET /payment.
type PaymentListParams struct {
	Skip  int
	Limit int
}

// PaymentList is one page of GET /payment.
type PaymentList struct {
	Payments []Payment `json:"payments"`
	PageInfo PageInfo  `json:"pageInfo"`
	Skip     int
	Limit    int
}

type paymentEnvelope struct {
	Payment Payment `json:"payment"`
}

func (s *PaymentsService) Create(ctx context.Context, params *PaymentCreateParams, opts ...RequestOption) (*Payment, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: PaymentCreateParams is required")
	}
	if params.CorrelationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	byQR := params.QRCode != ""
	byKey := params.DestinationAlias != ""
	byBoleto := params.Type == PaymentTypeBoleto || params.BoletoBarcode != ""
	switch {
	case byBoleto:
		if params.BoletoBarcode == "" {
			return nil, fmt.Errorf("woovi: boletoBarcode is required")
		}
		if err := validateBoletoBarcode(params.BoletoBarcode); err != nil {
			return nil, err
		}
		if params.Type == "" {
			params.Type = PaymentTypeBoleto
		}
	case byQR:
	case byKey:
		if params.Value <= 0 {
			return nil, fmt.Errorf("woovi: value must be greater than zero")
		}
		if params.DestinationAliasType == "" {
			return nil, fmt.Errorf("woovi: destinationAliasType is required")
		}
	default:
		return nil, fmt.Errorf("woovi: provide destinationAlias, qrCode, or boletoBarcode")
	}

	var out paymentEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/payment", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Payment, nil
}

// Get retrieves a payment by correlationID or id.
func (s *PaymentsService) Get(ctx context.Context, id string, opts ...RequestOption) (*Payment, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out paymentEnvelope
	path := "/api/v1/payment/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Payment, nil
}

func (s *PaymentsService) Approve(ctx context.Context, correlationID string, opts ...RequestOption) (*PaymentApproveResult, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	body := map[string]string{"correlationID": correlationID}
	var out PaymentApproveResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/payment/approve", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PaymentsService) List(ctx context.Context, params *PaymentListParams, opts ...RequestOption) (*PaymentList, error) {
	if params == nil {
		params = &PaymentListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out PaymentList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/payment", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *PaymentsService) ListAll(ctx context.Context, params *PaymentListParams, opts ...RequestOption) iter.Seq2[*Payment, error] {
	return func(yield func(*Payment, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Payments {
				if !yield(&page.Payments[i], nil) {
					return
				}
			}
		}
	}
}

func (s *PaymentsService) ListPages(ctx context.Context, params *PaymentListParams, opts ...RequestOption) iter.Seq2[*PaymentList, error] {
	return func(yield func(*PaymentList, error) bool) {
		p := PaymentListParams{}
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
			if !listPageHasMore(page.PageInfo, len(page.Payments), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}

func normalizePageInfo(info *PageInfo, skip, limit int) {
	if info.Limit == 0 {
		info.Limit = limit
	}
	if info.Skip == 0 && skip > 0 {
		info.Skip = skip
	}
}
