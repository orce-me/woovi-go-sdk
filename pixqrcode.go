package woovi

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode"
)

// PixQRCodesService is /api/v1/qrcode-static.
type PixQRCodesService struct {
	client *Client
}

// PixQRCodeCreateParams creates a static QR; identifier required, value optional (cents).
type PixQRCodeCreateParams struct {
	Name          string `json:"name"`
	Identifier    string `json:"identifier"`
	CorrelationID string `json:"correlationID,omitempty"`
	Value         *Money `json:"value,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

// PixQRCode is a static Pix QR (qrcode-static; value in cents when set).
type PixQRCode struct {
	Name           string    `json:"name"`
	Identifier     string    `json:"identifier"`
	CorrelationID  string    `json:"correlationID,omitempty"`
	PaymentLinkID  string    `json:"paymentLinkID,omitempty"`
	Value          Money     `json:"value,omitempty"`
	Comment        string    `json:"comment,omitempty"`
	CreatedAt      time.Time `json:"createdAt,omitempty"`
	UpdatedAt      time.Time `json:"updatedAt,omitempty"`
	BrCode         string    `json:"brCode,omitempty"`
	PaymentLinkURL string    `json:"paymentLinkUrl,omitempty"`
	QRCodeImage    string    `json:"qrCodeImage,omitempty"`
}

// PixQRCodeListParams paginates GET /qrcode-static.
type PixQRCodeListParams struct {
	Skip  int
	Limit int
}

// PixQRCodeList is one page of GET /qrcode-static.
type PixQRCodeList struct {
	PixQRCodes []PixQRCode `json:"pixQrCodes"`
	PageInfo   PageInfo    `json:"pageInfo"`
	Skip       int
	Limit      int
}

type pixQRCodeEnvelope struct {
	PixQRCode PixQRCode `json:"pixQrCode"`
}

func (s *PixQRCodesService) Create(ctx context.Context, params *PixQRCodeCreateParams, opts ...RequestOption) (*PixQRCode, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: PixQRCodeCreateParams is required")
	}
	if params.Name == "" {
		return nil, fmt.Errorf("woovi: name is required")
	}
	if params.Identifier != "" {
		if err := validatePixQRCodeIdentifier(params.Identifier); err != nil {
			return nil, err
		}
	}
	if params.Value != nil && *params.Value < 0 {
		return nil, fmt.Errorf("woovi: value must not be negative")
	}

	var out pixQRCodeEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/qrcode-static", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.PixQRCode, nil
}

// Get retrieves a static QR code by correlationID or identifier.
func (s *PixQRCodesService) Get(ctx context.Context, id string, opts ...RequestOption) (*PixQRCode, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out pixQRCodeEnvelope
	path := "/api/v1/qrcode-static/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.PixQRCode, nil
}

func (s *PixQRCodesService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	if id == "" {
		return fmt.Errorf("woovi: id is required")
	}
	path := "/api/v1/qrcode-static/" + url.PathEscape(id)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil, nil, opts...)
}

func (s *PixQRCodesService) List(ctx context.Context, params *PixQRCodeListParams, opts ...RequestOption) (*PixQRCodeList, error) {
	if params == nil {
		params = &PixQRCodeListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out PixQRCodeList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/qrcode-static", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *PixQRCodesService) ListAll(ctx context.Context, params *PixQRCodeListParams, opts ...RequestOption) iter.Seq2[*PixQRCode, error] {
	return func(yield func(*PixQRCode, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.PixQRCodes {
				if !yield(&page.PixQRCodes[i], nil) {
					return
				}
			}
		}
	}
}

func (s *PixQRCodesService) ListPages(ctx context.Context, params *PixQRCodeListParams, opts ...RequestOption) iter.Seq2[*PixQRCodeList, error] {
	return func(yield func(*PixQRCodeList, error) bool) {
		p := PixQRCodeListParams{}
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
			if !listPageHasMore(page.PageInfo, len(page.PixQRCodes), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}

func validatePixQRCodeIdentifier(id string) error {
	if len(id) > 25 {
		return fmt.Errorf("woovi: identifier must be at most 25 characters")
	}
	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("woovi: identifier must be alphanumeric")
		}
	}
	return nil
}
