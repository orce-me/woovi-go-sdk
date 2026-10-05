package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GiftbacksService is GET /api/v1/giftback/balance/{taxID} (live; not in public OpenAPI yet).
type GiftbacksService struct {
	client *Client
}

// GiftbackBalance is GET /giftback/balance/{taxID} (balance in cents).
type GiftbackBalance struct {
	Balance Money  `json:"balance"`
	Status  string `json:"status,omitempty"`
}

// Balance returns Giftback balance for taxID. Requires GIFTBACK_BALANCE_GET.
func (s *GiftbacksService) Balance(ctx context.Context, taxID string, opts ...RequestOption) (*GiftbackBalance, error) {
	if taxID == "" {
		return nil, fmt.Errorf("woovi: taxID is required")
	}
	var out GiftbackBalance
	path := "/api/v1/giftback/balance/" + url.PathEscape(taxID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
