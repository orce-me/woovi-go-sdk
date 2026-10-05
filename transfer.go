package woovi

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// TransfersService is POST /api/v1/transfer.
type TransfersService struct {
	client *Client
}

// TransferCreateParams moves value (cents) between two Woovi account Pix keys.
type TransferCreateParams struct {
	Value         Money  `json:"value"`
	FromPixKey    string `json:"fromPixKey"`
	ToPixKey      string `json:"toPixKey"`
	CorrelationID string `json:"correlationID,omitempty"`
}

// Transfer is POST /transfer result (value in cents).
type Transfer struct {
	Value         Money     `json:"value"`
	Time          time.Time `json:"time,omitempty"`
	CorrelationID string    `json:"correlationID,omitempty"`
	FromPixKey    string    `json:"fromPixKey,omitempty"`
	ToPixKey      string    `json:"toPixKey,omitempty"`
}

// Create transfers value between two Woovi account Pix keys.
func (s *TransfersService) Create(ctx context.Context, params *TransferCreateParams, opts ...RequestOption) (*Transfer, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: TransferCreateParams is required")
	}
	if params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}
	if params.FromPixKey == "" || params.ToPixKey == "" {
		return nil, fmt.Errorf("woovi: fromPixKey and toPixKey are required")
	}

	var out Transfer
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/transfer", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
