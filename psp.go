package woovi

import (
	"context"
	"net/http"
	"net/url"
)

// PSPsService is GET /api/v1/psp.
type PSPsService struct {
	client *Client
}

// PSP is a Brazilian Pix participant (ISPB / COMPE).
type PSP struct {
	Name  string `json:"name"`
	ISPB  string `json:"ispb"`
	Code  string `json:"code,omitempty"`
	Compe string `json:"compe,omitempty"`
}

// PSPListParams filters GET /psp by ISPB, name, or COMPE.
type PSPListParams struct {
	ISPB  string
	Name  string
	Compe string
}

func (s *PSPsService) List(ctx context.Context, params *PSPListParams, opts ...RequestOption) ([]PSP, error) {
	q := url.Values{}
	if params != nil {
		if params.ISPB != "" {
			q.Set("ispb", params.ISPB)
		}
		if params.Name != "" {
			q.Set("name", params.Name)
		}
		if params.Compe != "" {
			q.Set("compe", params.Compe)
		}
	}
	var out struct {
		Success bool  `json:"success"`
		PSPs    []PSP `json:"psps"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/psp", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	return out.PSPs, nil
}
