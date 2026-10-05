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

// ChargeStatus is ACTIVE, COMPLETED, EXPIRED, or completed with different payer.
type ChargeStatus string

const (
	ChargeStatusActive                ChargeStatus = "ACTIVE"
	ChargeStatusCompleted             ChargeStatus = "COMPLETED"
	ChargeStatusExpired               ChargeStatus = "EXPIRED"
	ChargeStatusCompletedNotSamePayer ChargeStatus = "COMPLETED_NOT_SAME_CUSTOMER_PAYER"
)

// ChargeType is DYNAMIC (immediate) or OVERDUE (cobrança com vencimento).
type ChargeType string

const (
	ChargeTypeDynamic ChargeType = "DYNAMIC"
	ChargeTypeOverdue ChargeType = "OVERDUE"
)

// ChargesService is POST/GET/PATCH/DELETE /api/v1/charge.
type ChargesService struct {
	client *Client
}

// AdditionalInfo is a key/value pair stored on a charge.
type AdditionalInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TaxID is a CPF/CNPJ object (taxID plus optional type).
type TaxID struct {
	TaxID string `json:"taxID"`
	Type  string `json:"type,omitempty"`
}

// Customer is a stored OpenPix customer (responses and nested objects).
type Customer struct {
	Name          string   `json:"name,omitempty"`
	Email         string   `json:"email,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	TaxID         *TaxID   `json:"taxID,omitempty"`
	CorrelationID string   `json:"correlationID,omitempty"`
	Address       *Address `json:"address,omitempty"`
}

// Address is a Brazilian postal address on customer or charge input.
type Address struct {
	Zipcode      string `json:"zipcode,omitempty"`
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Country      string `json:"country,omitempty"`
}

// ChargeCustomerInput is customer data on charge create. Pass TaxID alone to reuse a customer.
type ChargeCustomerInput struct {
	Name          string   `json:"name,omitempty"`
	Email         string   `json:"email,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	TaxID         string   `json:"taxID,omitempty"`
	CorrelationID string   `json:"correlationID,omitempty"`
	Address       *Address `json:"address,omitempty"`
}

// SplitType selects how a charge split is routed (SPLIT_SUB_ACCOUNT).
type SplitType string

const (
	SplitTypeSubAccount SplitType = "SPLIT_SUB_ACCOUNT"
)

// ChargeSplit routes part of a charge value to a subaccount.
type ChargeSplit struct {
	PixKey    string    `json:"pixKey"`
	Value     Money     `json:"value"`
	SplitType SplitType `json:"splitType"`
}

// ChargeCreateParams creates a charge; correlationID and value (cents) required.
type ChargeCreateParams struct {
	CorrelationID    string               `json:"correlationID"`
	Value            Money                `json:"value"`
	Comment          *string              `json:"comment,omitempty"`
	Customer         *ChargeCustomerInput `json:"customer,omitempty"`
	ExpiresIn        *int                 `json:"expiresIn,omitempty"`
	AdditionalInfo   []AdditionalInfo     `json:"additionalInfo,omitempty"`
	DaysForDueDate   *int                 `json:"daysForDueDate,omitempty"`
	DaysAfterDueDate *int                 `json:"daysAfterDueDate,omitempty"`
	Type             ChargeType           `json:"type,omitempty"`
	Splits           []ChargeSplit        `json:"splits,omitempty"`
	Subaccount       string               `json:"subaccount,omitempty"`
	Cashback         *ChargeCashback      `json:"cashback,omitempty"`
}

// ChargeUpdateParams patches a charge; omitted fields stay unchanged.
type ChargeUpdateParams struct {
	Comment        *string              `json:"comment,omitempty"`
	Customer       *ChargeCustomerInput `json:"customer,omitempty"`
	ExpiresIn      *int                 `json:"expiresIn,omitempty"`
	AdditionalInfo []AdditionalInfo     `json:"additionalInfo,omitempty"`
	Value          *Money               `json:"value,omitempty"`
}

// Charge is a Pix cobrança (value in cents; keyed by correlationID or identifier).
type Charge struct {
	Customer             *Customer        `json:"customer"`
	Value                Money            `json:"value"`
	Identifier           string           `json:"identifier"`
	CorrelationID        string           `json:"correlationID"`
	PaymentLinkID        string           `json:"paymentLinkID"`
	TransactionID        string           `json:"transactionID"`
	Status               ChargeStatus     `json:"status"`
	GiftbackAppliedValue Money            `json:"giftbackAppliedValue"`
	Discount             Money            `json:"discount"`
	ValueWithDiscount    Money            `json:"valueWithDiscount"`
	ExpiresDate          time.Time        `json:"expiresDate"`
	Type                 ChargeType       `json:"type"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
	PaidAt               *time.Time       `json:"paidAt,omitempty"`
	AdditionalInfo       []AdditionalInfo `json:"additionalInfo"`
	ExpiresIn            int              `json:"expiresIn"`
	PixKey               string           `json:"pixKey"`
	BrCode               string           `json:"brCode"`
	PaymentLinkURL       string           `json:"paymentLinkUrl"`
	QRCodeImage          string           `json:"qrCodeImage"`
	GlobalID             string           `json:"globalID"`
	Comment              string           `json:"comment,omitempty"`
	Splits               []ChargeSplit    `json:"splits,omitempty"`
}

// ChargeCreateResult is POST /charge (charge, correlationID, brCode).
type ChargeCreateResult struct {
	Charge        Charge `json:"charge"`
	CorrelationID string `json:"correlationID"`
	BrCode        string `json:"brCode"`
}

// ChargeGetResult wraps a single charge from GET/PATCH /charge/{id}.
type ChargeGetResult struct {
	Charge Charge `json:"charge"`
}

// ChargeListParams paginates charges (optional status filter).
type ChargeListParams struct {
	Skip   int
	Limit  int
	Status ChargeStatus
}

// ChargeList is one page of GET /charge.
type ChargeList struct {
	Charges  []Charge `json:"charges"`
	PageInfo PageInfo `json:"pageInfo"`
	Skip     int      // effective skip for this page
	Limit    int      // effective limit for this page
}

// Create is POST /api/v1/charge.
func (s *ChargesService) Create(ctx context.Context, params *ChargeCreateParams, opts ...RequestOption) (*ChargeCreateResult, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: ChargeCreateParams is required")
	}
	if params.CorrelationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	if params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}

	var out ChargeCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/charge", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a charge by correlation ID or identifier.
func (s *ChargesService) Get(ctx context.Context, id string, opts ...RequestOption) (*Charge, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out ChargeGetResult
	path := "/api/v1/charge/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Charge, nil
}

// Update patches a charge by correlation ID or identifier.
func (s *ChargesService) Update(ctx context.Context, id string, params *ChargeUpdateParams, opts ...RequestOption) (*Charge, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	if params == nil {
		return nil, fmt.Errorf("woovi: ChargeUpdateParams is required")
	}
	var out ChargeGetResult
	path := "/api/v1/charge/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodPatch, path, nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Charge, nil
}

// Delete removes a charge by correlation ID or identifier.
func (s *ChargesService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	if id == "" {
		return fmt.Errorf("woovi: id is required")
	}
	path := "/api/v1/charge/" + url.PathEscape(id)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil, nil, opts...)
}

// ChargeImage is a downloaded charge QR PNG (raw bytes).
type ChargeImage struct {
	Data        []byte
	ContentType string
}

// BRCodeImage downloads the charge QR PNG by payment link id. size optional, 600–4096.
func (s *ChargesService) BRCodeImage(ctx context.Context, id string, size int, opts ...RequestOption) (*ChargeImage, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	if size != 0 && (size < 600 || size > 4096) {
		return nil, fmt.Errorf("woovi: size must be between 600 and 4096")
	}
	q := url.Values{}
	if size > 0 {
		q.Set("size", strconv.Itoa(size))
	}
	path := "/openpix/charge/brcode/image/" + url.PathEscape(id) + ".png"
	data, contentType, err := s.client.doBytes(ctx, http.MethodGet, path, q, opts...)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "image/png"
	}
	return &ChargeImage{Data: data, ContentType: contentType}, nil
}

// QRCodeBase64 downloads the charge QR as a base64 data URL.
func (s *ChargesService) QRCodeBase64(ctx context.Context, id string, size int, opts ...RequestOption) (string, error) {
	if id == "" {
		return "", fmt.Errorf("woovi: id is required")
	}
	if size != 0 && (size < 600 || size > 4096) {
		return "", fmt.Errorf("woovi: size must be between 600 and 4096")
	}
	q := url.Values{}
	if size > 0 {
		q.Set("size", strconv.Itoa(size))
	}
	path := "/api/image/qrcode/base64/" + url.PathEscape(id)
	var out struct {
		Success     bool   `json:"success"`
		ImageBase64 string `json:"imageBase64"`
	}
	if err := s.client.do(ctx, http.MethodGet, path, q, nil, &out, opts...); err != nil {
		return "", err
	}
	if out.ImageBase64 == "" {
		return "", fmt.Errorf("woovi: empty qrcode base64 response")
	}
	return out.ImageBase64, nil
}

func (s *ChargesService) List(ctx context.Context, params *ChargeListParams, opts ...RequestOption) (*ChargeList, error) {
	if params == nil {
		params = &ChargeListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))
	if params.Status != "" {
		q.Set("status", string(params.Status))
	}

	var out ChargeList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/charge", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *ChargesService) ListAll(ctx context.Context, params *ChargeListParams, opts ...RequestOption) iter.Seq2[*Charge, error] {
	return func(yield func(*Charge, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Charges {
				if !yield(&page.Charges[i], nil) {
					return
				}
			}
		}
	}
}

// ListPages iterates page envelopes for checkpointing.
func (s *ChargesService) ListPages(ctx context.Context, params *ChargeListParams, opts ...RequestOption) iter.Seq2[*ChargeList, error] {
	return func(yield func(*ChargeList, error) bool) {
		p := ChargeListParams{}
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

			if !listPageHasMore(page.PageInfo, len(page.Charges), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}
