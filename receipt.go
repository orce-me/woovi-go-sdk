package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ReceiptType selects which receipt PDF to export.
type ReceiptType string

const (
	ReceiptTypePixIn     ReceiptType = "pix-in"
	ReceiptTypePixOut    ReceiptType = "pix-out"
	ReceiptTypePixRefund ReceiptType = "pix-refund"
)

// ReceiptsService is GET /api/v1/receipt/{type}/{endToEndId}.
type ReceiptsService struct {
	client *Client
}

// Receipt is a downloaded receipt PDF (raw bytes).
type Receipt struct {
	Data        []byte
	ContentType string
}

// Get downloads a receipt PDF for a transaction endToEndId.
func (s *ReceiptsService) Get(ctx context.Context, receiptType ReceiptType, endToEndID string, opts ...RequestOption) (*Receipt, error) {
	if receiptType == "" {
		return nil, fmt.Errorf("woovi: receiptType is required")
	}
	if endToEndID == "" {
		return nil, fmt.Errorf("woovi: endToEndID is required")
	}
	path := "/api/v1/receipt/" + url.PathEscape(string(receiptType)) + "/" + url.PathEscape(endToEndID)
	data, contentType, err := s.client.doBytes(ctx, http.MethodGet, path, nil, opts...)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "application/pdf"
	}
	return &Receipt{Data: data, ContentType: contentType}, nil
}
